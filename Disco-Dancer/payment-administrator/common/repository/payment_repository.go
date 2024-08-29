package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/clients"
	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
	"geico.visualstudio.com/Billing/plutus/proto/paymentpreference"
	"github.com/jackc/pgx/v5"
)

const (
	user = "payment-administrator"

	ErrorUnableToGetPaymentPreferences = "unable to get payment preferences from payment vault"

	IncomingPaymentRequestAlreadyExist = "unable to process payment, as previous payment request with same tenantid and tenantRequestId already exist"

	updatePaymentStatusQuery                     = `UPDATE public.payment SET "Status" = $%d WHERE "PaymentId" IN (%s)`
	updatePaymentRequestStatusQuery              = `UPDATE public.incoming_payment_request SET "Status" = $%d WHERE "RequestId" IN (%s)`
	updatePaymentRequestStatusFromPaymentIdQuery = `UPDATE public.incoming_payment_request SET "Status" = $%d WHERE "RequestId" IN (SELECT "RequestId" FROM public.payment WHERE "PaymentId" IN (%s))`
)

var repositoryLog = logging.GetLogger("payment-administrator-repository")
var ExecutorExecuteRequestTopic string
var DataHubPaymentEventsTopic string

//go:generate mockery --name PaymentRepositoryInterface
type PaymentRepositoryInterface interface {
	MakePayment(incomingPaymentRequest *dbmodels.IncomingPaymentRequest) ([]dbmodels.Payment, error)
	ProcessPayments(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) (*int, error)
	UpdatePaymentStatus(executePaymentResponses commonMessagingModels.ExecutePaymentResponse) error
	GetTenantInformationForPaymentId(paymentId int64) (int64, int64, json.RawMessage, error)
}

type PaymentRepository struct {
}

func (PaymentRepository) MakePayment(incomingPaymentRequest *dbmodels.IncomingPaymentRequest) ([]dbmodels.Payment, error) {

	if isIncomingPaymentRequestAlreadyExist(incomingPaymentRequest.TenantId, incomingPaymentRequest.TenantRequestId) {
		repositoryLog.Info(context.Background(), IncomingPaymentRequestAlreadyExist)
		return nil, errors.New(IncomingPaymentRequestAlreadyExist)
	}

	if incomingPaymentRequest.PaymentRequestTypeEnum == enums.CustomerChoice {
		return addPaymentInfo(incomingPaymentRequest)
	} else if incomingPaymentRequest.PaymentRequestTypeEnum == enums.InsuranceAutoAuctions {

		paymentPreferenceList, err := clients.GetPaymentPreference(
			incomingPaymentRequest.UserId,
			incomingPaymentRequest.ProductIdentifier,
			incomingPaymentRequest.TransactionTypeEnum.String(),
			incomingPaymentRequest.PaymentRequestTypeEnum.String())

		if err != nil {
			repositoryLog.Info(context.Background(), "grpc call for GetPaymentPreference failed, for paymentId '%d.", incomingPaymentRequest.RequestId)
			return nil, errors.New("grpc call to GetPaymentPreference failed, for userId " + incomingPaymentRequest.UserId)
		} else {
			repositoryLog.Info(context.Background(), "grpc call for GetPaymentPreference successfull, for paymentId '%d.", incomingPaymentRequest.RequestId)
		}

		if len(paymentPreferenceList) == 0 {
			// Bug 8906138 - If paymentpreference not found, then do not insert the record. Commented below line
			// addIncomingPaymentRequest(incomingPaymentRequest, enums.Errored.EnumIndex())
			repositoryLog.Info(context.Background(), "unable to get payment preferences from payment vault for userid %s", incomingPaymentRequest.UserId)
			return nil, errors.New(ErrorUnableToGetPaymentPreferences)
		} else {
			for _, paymentPreference := range paymentPreferenceList {
				if paymentPreference.Split == 100 {
					incomingPaymentRequest.PaymentMethodType = enums.GetPaymentMethodTypeEnumFromString(paymentPreference.PaymentMethodType)
				}
			}

			return addPaymentInfo(incomingPaymentRequest)
		}
	}

	return nil, nil
}

func (PaymentRepository) ProcessPayments(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) (*int, error) {
	//TODO: batchCount, totalErrorCount, totalKafkaErrorCount Count should come from config Management later
	var batchCount int = 5000
	var totalCount int = 0
	var totalErrorCount int = 10
	var totalKafkaErrorCount = 5
	for {
		processPaymentRequests, err := getPayments(executeTaskRequest, batchCount)

		if err != nil {
			totalErrorCount--
			if totalErrorCount > 0 {
				continue
			} else {
				repositoryLog.Error(context.Background(), err, "get functionality of batch processing error count exceeded maximum allowed count 10, cannot proceed with payment processing")
				return &totalCount, err
			}
		}

		if len(processPaymentRequests) == 0 {
			return &totalCount, nil
		}

		processPaymentRequestsOut, err := processPayments(processPaymentRequests)
		if err != nil {
			totalErrorCount--
			if processPaymentRequestsOut != nil {
				updatePaymentsStatus(processPaymentRequestsOut, enums.Accepted)
			}
			if totalErrorCount > 0 {
				continue
			} else {
				repositoryLog.Error(context.Background(), err, "process functionality of batch processing error count exceeded maximum allowed count 10, cannot proceed with payment processing")
				return &totalCount, err
			}
		}
		totalCount = totalCount + len(processPaymentRequestsOut)

		publishMessagesToExecutorAndDataHub(processPaymentRequestsOut, &totalKafkaErrorCount, &totalCount)

		if totalKafkaErrorCount <= 0 {
			err = errors.New("publish functionality of batch processing error count exceeded maximum allowed count 5, cannot proceed with payment processing")
			repositoryLog.Error(context.Background(), err, "")
			return &totalCount, err
		}
	}

}

func (PaymentRepository) UpdatePaymentStatus(executePaymentResponses commonMessagingModels.ExecutePaymentResponse) error {
	transaction, err := database.NewDbContext().Database.Begin(context.Background())
	defer transaction.Rollback(context.Background())

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in starting transaction for updating payment status")
		return err
	}

	paymentIds := []int64{executePaymentResponses.PaymentId}
	if err := updatePaymentInfoStatus(transaction, paymentIds, updatePaymentRequestStatusFromPaymentIdQuery, executePaymentResponses.Status); err != nil {
		return err
	}
	if err := updatePaymentInfoStatus(transaction, paymentIds, updatePaymentStatusQuery, executePaymentResponses.Status); err != nil {
		return err
	}

	return transaction.Commit(context.Background())
}

func (PaymentRepository) GetTenantInformationForPaymentId(paymentId int64) (int64, int64, json.RawMessage, error) {
	var tenantId, tenantRequestId int64
	var metadata json.RawMessage

	getPaymentRequest := `SELECT "TenantId", p."TenantRequestId", i."Metadata"
						  FROM public.payment AS p 
						  INNER JOIN public.incoming_payment_request AS i
						  ON p."RequestId" = i."RequestId"
						  WHERE "PaymentId" = $1`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentRequest,
		paymentId)

	err := row.Scan(&tenantId, &tenantRequestId, &metadata)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error reading from payment tables")
		return 0, 0, nil, err
	}

	return tenantId, tenantRequestId, metadata, nil
}

// TODO add logic for handling autopay payments
func processPayments(processPaymentRequests []*dbmodels.ProcessPaymentRequest) ([]*dbmodels.ProcessPaymentRequest, error) {

	err := updatePaymentsStatus(processPaymentRequests, enums.InProgress)
	if err != nil {
		return nil, err
	}

	var erroredPaymentRequests []*dbmodels.ProcessPaymentRequest
	var successPaymentRequests []*dbmodels.ProcessPaymentRequest
	//var splitPaymentRequests []*dbmodels.ProcessPaymentRequest

	for i := len(processPaymentRequests) - 1; i >= 0; i-- {
		processPaymentRequest := processPaymentRequests[i]
		paymentPreferenceList, err := clients.GetPaymentPreference(
			processPaymentRequest.UserId,
			processPaymentRequest.ProductIdentifier,
			processPaymentRequest.TransactionType.String(),
			processPaymentRequest.PaymentRequestType.String())

		if err != nil || len(paymentPreferenceList) == 0 {
			repositoryLog.Info(context.Background(), "grpc call for GetPaymentPreference failed, for paymentId '%d.", processPaymentRequest.PaymentId)
			erroredPaymentRequests = append(erroredPaymentRequests, processPaymentRequest)
			processPaymentRequests = append(processPaymentRequests[:i], processPaymentRequests[i+1:]...)
		} else {
			//var isSplit bool
			//amount := processPaymentRequest.Amount
			for _, paymentPreference := range paymentPreferenceList {
				paymentMethodTypeEnum := enums.GetPaymentMethodTypeEnumFromString(paymentPreference.PaymentMethodType)

				if processPaymentRequest.PaymentMethodType == enums.ACH && paymentMethodTypeEnum == enums.Card {
					continue
				} else if processPaymentRequest.PaymentMethodType == enums.Card && paymentMethodTypeEnum == enums.ACH {
					continue
				}
				if paymentPreference.Split == 100 {
					populatePaymentPreferenceInfo(processPaymentRequest, paymentPreference)
					successPaymentRequests = append(successPaymentRequests, processPaymentRequest)
					break
				} else {
					repositoryLog.Info(context.Background(), "split amount not equal to 100 is not supported, for paymentId '%d.", processPaymentRequest.PaymentId)
					erroredPaymentRequests = append(erroredPaymentRequests, processPaymentRequest)
					processPaymentRequests = append(processPaymentRequests[:i], processPaymentRequests[i+1:]...)
					break
				}
				//TODO: Add the split logic for day 2
				// } else if !isSplit {
				// 	processPaymentRequest.Amount = float32(math.Round(float64((float32(paymentPreference.Split)/100*amount))*100) / 100)
				// 	populatePaymentPreferenceInfo(processPaymentRequest, paymentPreference)
				// 	isSplit = true
				// 	successPaymentRequests = append(successPaymentRequests, processPaymentRequest)
				// } else {
				// 	var executePaymentRequestNew dbmodels.ProcessPaymentRequest = *processPaymentRequest
				// 	executePaymentRequestNew.Amount = float32(math.Round(float64((float32(paymentPreference.Split)/100*amount))*100) / 100)
				// 	populatePaymentPreferenceInfo(&executePaymentRequestNew, paymentPreference)
				// 	splitPaymentRequests = append(splitPaymentRequests, &executePaymentRequestNew)
				// }

			}

		}

	}

	// if len(splitPaymentRequests) > 0 {
	// 	processPaymentRequests = append(processPaymentRequests, splitPaymentRequests...)
	// }

	if len(erroredPaymentRequests) > 0 {
		if err := updatePaymentsStatus(erroredPaymentRequests, enums.Errored); err != nil {
			return nil, err
		}
	}

	// Note: Commented this to keep the payment status in "Inprogress"
	// if len(successPaymentRequests) > 0 {
	// 	if err := updatePaymentsStatus(successPaymentRequests, enums.Completed); err != nil {
	// 		return nil, err
	// 	}
	// }

	return processPaymentRequests, nil
}

func publishMessagesToExecutorAndDataHub(processPaymentRequests []*dbmodels.ProcessPaymentRequest, totalKafkaErrorCount *int, totalCount *int) error {

	for _, processPaymentRequest := range processPaymentRequests {

		processPaymentRq, err := getProcessPaymentRequestJson(*processPaymentRequest)
		if err != nil {
			*totalKafkaErrorCount--
			repositoryLog.Error(context.Background(), err, "error before publishing the message to kafka, resetting payment id '%d' status", processPaymentRequest.PaymentId)
			updatePaymentStatusSingle(processPaymentRequest, enums.Accepted)
			*totalCount--
			continue
		}
		err = messaging.KafkaPublish(ExecutorExecuteRequestTopic, *processPaymentRq)
		if err != nil {
			*totalKafkaErrorCount--
			repositoryLog.Error(context.Background(), err, "error while publishing the message to kafka, resetting payment id '%d' status", processPaymentRequest.PaymentId)
			updatePaymentStatusSingle(processPaymentRequest, enums.Accepted)
			*totalCount--
		} else {
			publishPaymentEventToDataHub(context.Background(), *processPaymentRequest)
		}

	}
	return nil

}

func getProcessPaymentRequestJson(processPaymentRequest dbmodels.ProcessPaymentRequest) (*string, error) {
	processPaymentRq := commonMessagingModels.ProcessPaymentRequest{
		TenantId:               processPaymentRequest.TenantId,
		PaymentId:              processPaymentRequest.PaymentId,
		PaymentFrequency:       processPaymentRequest.PaymentFrequency.String(),
		TransactionType:        processPaymentRequest.TransactionType.String(),
		PaymentRequestType:     processPaymentRequest.PaymentRequestType.String(),
		PaymentMethodType:      processPaymentRequest.PaymentMethodType.String(),
		PaymentExtendedData:    processPaymentRequest.PaymentExtendedData,
		AccountIdentifier:      processPaymentRequest.AccountIdentifier,
		Last4AccountIdentifier: processPaymentRequest.Last4AccountIdentifier,
		RoutingNumber:          processPaymentRequest.RoutingNumber,
		Amount:                 processPaymentRequest.Amount,
		PaymentDate:            commonMessagingModels.JsonDate(processPaymentRequest.PaymentDate),
	}
	processPaymentRequestJson, err := json.Marshal(processPaymentRq)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to marshal processPaymentRequest")
		return nil, err
	}

	processPaymentRequestJsonStr := string(processPaymentRequestJson)
	return &processPaymentRequestJsonStr, nil
}

func updatePaymentStatusSingle(processPaymentRequest *dbmodels.ProcessPaymentRequest, paymentStatus enums.PaymentStatus) error {
	var processPaymentRequests []*dbmodels.ProcessPaymentRequest
	processPaymentRequests = append(processPaymentRequests, processPaymentRequest)
	return updatePaymentsStatus(processPaymentRequests, paymentStatus)
}

func updatePaymentsStatus(processPaymentRequests []*dbmodels.ProcessPaymentRequest, paymentStatus enums.PaymentStatus) error {

	transaction, err := database.NewDbContext().Database.Begin(context.Background())
	defer transaction.Rollback(context.Background())

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in starting transaction for updating payment status")
		return err
	}

	paymentIds := getpaymentIds(processPaymentRequests)
	if err := updatePaymentInfoStatus(transaction, paymentIds, updatePaymentStatusQuery, paymentStatus); err != nil {
		return err
	}

	requestIds := getpaymentRequestIds(processPaymentRequests)
	if err := updatePaymentInfoStatus(transaction, requestIds, updatePaymentRequestStatusQuery, paymentStatus); err != nil {
		return err
	}

	transaction.Commit(context.Background())
	return nil
}

func updatePaymentInfoStatus(transaction pgx.Tx, ids []int64, query string, paymentStatus enums.PaymentStatus) error {

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query = fmt.Sprintf(query, len(ids)+1, strings.Join(placeholders, ","))

	args := make([]interface{}, len(ids)+1)
	for i, id := range ids {
		args[i] = id
	}
	args[len(ids)] = paymentStatus.EnumIndex()

	_, err := transaction.Exec(context.Background(), query, args...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Unable to update payment status")
		return err
	}

	return nil
}

func getpaymentIds(processPaymentRequests []*dbmodels.ProcessPaymentRequest) []int64 {
	var paymentIds []int64
	for _, item := range processPaymentRequests {
		paymentIds = append(paymentIds, item.PaymentId)
	}
	return paymentIds
}

func getpaymentRequestIds(processPaymentRequests []*dbmodels.ProcessPaymentRequest) []int64 {
	var paymentRequestIds []int64
	for _, item := range processPaymentRequests {
		paymentRequestIds = append(paymentRequestIds, item.RequestId)
	}
	return paymentRequestIds
}

func populatePaymentPreferenceInfo(processPaymentRequest *dbmodels.ProcessPaymentRequest, paymentPreference *paymentpreference.PaymentPreference) {
	processPaymentRequest.PaymentMethodType = enums.GetPaymentMethodTypeEnumFromString(paymentPreference.PaymentMethodType)
	processPaymentRequest.PaymentExtendedData = paymentPreference.PaymentExtendedData
	processPaymentRequest.AccountIdentifier = paymentPreference.AccountIdentifier
	processPaymentRequest.RoutingNumber = paymentPreference.RoutingNumber
	processPaymentRequest.Last4AccountIdentifier = paymentPreference.Last4AccountIdentifier
}

func getPayments(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb, batchCount int) ([]*dbmodels.ProcessPaymentRequest, error) {

	var processPaymentRequests []*dbmodels.ProcessPaymentRequest
	rows, err := getPaymentRows(executeTaskRequest, batchCount)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		processPaymentRequest := dbmodels.ProcessPaymentRequest{
			PaymentMethodType: executeTaskRequest.ExecutionParametersDb.PaymentMethodType,
		}
		err := rows.Scan(
			&processPaymentRequest.TenantId,
			&processPaymentRequest.PaymentId,
			&processPaymentRequest.PaymentFrequency,
			&processPaymentRequest.TransactionType,
			&processPaymentRequest.PaymentRequestType,
			&processPaymentRequest.Amount,
			&processPaymentRequest.PaymentDate,
			&processPaymentRequest.UserId,
			&processPaymentRequest.ProductIdentifier,
			&processPaymentRequest.RequestId)
		if err != nil {
			repositoryLog.Error(context.Background(), err, "error fetching results from getPayments query")
			return nil, err
		}

		processPaymentRequests = append(processPaymentRequests, &processPaymentRequest)
	}

	return processPaymentRequests, nil
}

func getPaymentRows(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb, batchCount int) (pgx.Rows, error) {
	var isPaymentFrequencyAdditionalClause bool
	var isPaymentRequestTypeAdditionalClause bool
	var rows pgx.Rows
	var err error

	getPayments := `SELECT i."TenantId",
							p."PaymentId", 
							i."PaymentFrequency",
							i."TransactionType",
							i."PaymentRequestType",
							p."Amount"::numeric::float,
							p."PaymentDate",
							p."UserId",
							i."ProductIdentifier",
							i."RequestId"
					FROM public.payment AS p 
					INNER JOIN public.incoming_payment_request AS i
					ON p."RequestId" = i."RequestId"
					WHERE p."PaymentDate" <= $1 AND p."Status" = $2`

	if executeTaskRequest.ExecutionParametersDb.PaymentFrequency != enums.AllPaymentFrequency {
		isPaymentFrequencyAdditionalClause = true
	}
	if executeTaskRequest.ExecutionParametersDb.PaymentRequestType != enums.AllPaymentRequestType {
		isPaymentRequestTypeAdditionalClause = true
	}

	if !isPaymentFrequencyAdditionalClause && !isPaymentRequestTypeAdditionalClause {
		getPayments = getPayments + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getPayments,
			executeTaskRequest.TaskDate,
			enums.Accepted.EnumIndex())
	} else if isPaymentFrequencyAdditionalClause && isPaymentRequestTypeAdditionalClause {
		getPayments = getPayments + " AND i.\"PaymentFrequency\" = $3 AND i.\"PaymentRequestType\" = $4"
		getPayments = getPayments + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getPayments,
			executeTaskRequest.TaskDate,
			enums.Accepted.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentFrequency.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentRequestType.EnumIndex())
	} else if isPaymentFrequencyAdditionalClause {
		getPayments = getPayments + " AND i.\"PaymentFrequency\" = $3"
		getPayments = getPayments + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getPayments,
			executeTaskRequest.TaskDate,
			enums.Accepted.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentFrequency.EnumIndex())
	} else if isPaymentRequestTypeAdditionalClause {
		getPayments = getPayments + " AND i.\"PaymentRequestType\" = $3"
		getPayments = getPayments + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getPayments,
			executeTaskRequest.TaskDate,
			enums.Accepted.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentRequestType.EnumIndex())
	}
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error executing getPayments query")
		return nil, err
	}

	return rows, err
}

func isIncomingPaymentRequestAlreadyExist(tenantId int64, tenantRequestId int64) bool {
	isIncomingRequestAlreadyExist := false
	getPaymentRequest := `SELECT EXISTS
                             (SELECT 1 
							  FROM public.incoming_payment_request 
							  WHERE "TenantId" = $1 AND "TenantRequestId"= $2)`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentRequest,
		tenantId,
		tenantRequestId)
	err := row.Scan(&isIncomingRequestAlreadyExist)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error fetching incoming_payment_request table")
	}

	return isIncomingRequestAlreadyExist
}

func addPaymentInfo(incomingPaymentRequest *dbmodels.IncomingPaymentRequest) ([]dbmodels.Payment, error) {
	err := addIncomingPaymentRequest(incomingPaymentRequest, enums.Accepted.EnumIndex())
	if err != nil {
		return nil, err
	}
	payments, err := addPayments(incomingPaymentRequest)
	if err != nil {
		setIncomingPaymentRequestToErrorStatus(incomingPaymentRequest.RequestId)
		return nil, err
	}
	return payments, nil
}

func addIncomingPaymentRequest(incomingPaymentRequest *dbmodels.IncomingPaymentRequest, status int) error {

	addPaymentRequest := `INSERT INTO public.incoming_payment_request
	("TenantRequestId",
	 "TenantId",
	 "CallerApp", 
	 "AccountId", 
	 "UserId", 
	 "PaymentFrequency",
	 "TransactionType", 
	 "PaymentExtractionSchedule",
	 "ProductIdentifier",
	 "Amount",
	 "PaymentRequestType",
	 "Metadata",
	 "PaymentDate",
	 "Status",
	 "CreatedDate",
	 "CreatedBy",
	 "UpdatedDate",
	 "UpdatedBy"
	)
	VALUES(
	$1, $2, $3,	$4, $5, $6, $7,	$8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
	)
	RETURNING "RequestId"`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		addPaymentRequest,
		incomingPaymentRequest.TenantRequestId,
		incomingPaymentRequest.TenantId,
		incomingPaymentRequest.CallerAppEnum.EnumIndex(),
		incomingPaymentRequest.AccountId,
		incomingPaymentRequest.UserId,
		incomingPaymentRequest.PaymentFrequencyEnum.EnumIndex(),
		incomingPaymentRequest.TransactionTypeEnum.EnumIndex(),
		incomingPaymentRequest.PaymentExtractionScheduleJson,
		incomingPaymentRequest.ProductIdentifier,
		nil,
		incomingPaymentRequest.PaymentRequestTypeEnum.EnumIndex(),
		incomingPaymentRequest.Metadata,
		nil,
		status,
		time.Now(),
		user,
		time.Now(),
		user,
	)

	err := row.Scan(&incomingPaymentRequest.RequestId)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error inserting incoming_payment_request table")
		return err
	}
	return nil
}

func addPayments(incomingPaymentRequest *dbmodels.IncomingPaymentRequest) ([]dbmodels.Payment, error) {
	addPayment := `INSERT INTO public.payment
	("RequestId",
	 "TenantRequestId",
	 "ProductIdentifier", 
	 "AccountId", 
	 "UserId", 
	 "PaymentFrequency",
	 "PaymentDate", 
	 "Amount",
	 "Status",
	 "CreatedDate",
	 "CreatedBy",
	 "UpdatedDate",
	 "UpdatedBy"
	)
	VALUES(
	$1, $2, $3,	$4, $5, $6, $7,	$8, $9, $10, $11, $12, $13
	) RETURNING "PaymentId"`

	db := database.NewDbContext().Database
	tx, err := db.Begin(context.Background())
	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to start transaction for adding into payment table")
		return nil, err
	}

	var payments []dbmodels.Payment

	for _, paymentScheduleItem := range incomingPaymentRequest.PaymentExtractionSchedule {

		payment := dbmodels.Payment{
			RequestId:            incomingPaymentRequest.RequestId,
			TenantRequestId:      incomingPaymentRequest.TenantRequestId,
			ProductIdentifier:    incomingPaymentRequest.ProductIdentifier,
			AccountId:            incomingPaymentRequest.AccountId,
			UserId:               incomingPaymentRequest.UserId,
			PaymentFrequencyEnum: incomingPaymentRequest.PaymentFrequencyEnum,
			PaymentDate:          time.Time(paymentScheduleItem.Date),
			Amount:               paymentScheduleItem.Amount,
			Status:               enums.Accepted.EnumIndex(),
		}

		dbError := tx.QueryRow(context.Background(), addPayment,
			payment.RequestId,
			payment.TenantRequestId,
			payment.ProductIdentifier,
			payment.AccountId,
			payment.UserId,
			payment.PaymentFrequencyEnum.EnumIndex(),
			payment.PaymentDate,
			payment.Amount,
			payment.Status,
			time.Now(),
			user,
			time.Now(),
			user,
		).Scan(&payment.PaymentId)

		if dbError != nil {
			repositoryLog.Error(context.Background(), err, "unable to execute query for adding into payment table")
			return nil, err
		}

		payments = append(payments, payment)
	}
	err = tx.Commit(context.Background())
	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to commit query for payment table")
		return nil, err
	}
	defer tx.Rollback(context.Background())
	return payments, nil
}

func setIncomingPaymentRequestToErrorStatus(requestId int64) {
	updatePaymentRequestStatus := `UPDATE public.incoming_payment_request 
	                               SET "Status" = $2
	                               WHERE "RequestId" = $1 `

	_, err := database.NewDbContext().Database.Exec(context.Background(), updatePaymentRequestStatus, requestId, enums.Errored.EnumIndex())
	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to update payment request row to errored status with requestId %d", requestId)
	}

}

func publishPaymentEventToDataHub(ctx context.Context, request dbmodels.ProcessPaymentRequest) {
	processPaymentEvent, err := createPaymentEventJson(request)
	if err != nil {
		repositoryLog.Error(ctx, err, "error before publishing the payment event message to datahub, payment id '%d'", request.PaymentId)
	}

	err = messaging.KafkaPublish(DataHubPaymentEventsTopic, *processPaymentEvent)
	if err != nil {
		repositoryLog.Error(ctx, err, "error while publishing the payment event message to datahub, payment id '%d'", request.PaymentId)
	}
}

func createPaymentEventJson(processPaymentRequest dbmodels.ProcessPaymentRequest) (*string, error) {
	paymentEvent := commonMessagingModels.PaymentEvent{
		Version:            1,
		PaymentId:          processPaymentRequest.PaymentId,
		Amount:             processPaymentRequest.Amount,
		PaymentDate:        processPaymentRequest.PaymentDate,
		PaymentRequestType: processPaymentRequest.PaymentRequestType,
		PaymentMethodType:  processPaymentRequest.PaymentMethodType,
		EventType:          enums.SentByAdminstrator,
		EventDateTime:      time.Now(),
	}
	processPaymentRequestJson, err := json.Marshal(paymentEvent)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to marshal processPaymentRequest")
		return nil, err
	}

	processPaymentRequestJsonStr := string(processPaymentRequestJson)
	return &processPaymentRequestJsonStr, nil
}
