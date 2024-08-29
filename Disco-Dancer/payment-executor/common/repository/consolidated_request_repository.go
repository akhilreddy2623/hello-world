package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	dbModels "geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
	"github.com/jackc/pgx/v5"
)

// TODO: Need to add all the payment request types which needs to be consolidated
var paymentRequestTypeToConsolidate = [...]string{"insuranceautoauctions"}

type ConsolidatedRequestRepositoryInterface interface {
	ConsolidatePaymentRequests(executePaymentRequest dbModels.ExecuteTaskRequest) (int, error)
}

type ConsolidatedRequestRepository struct {
	// Add any necessary dependencies or fields here
}

// Main method or entry point for execution requests consolidation logic
func (ConsolidatedRequestRepository) ConsolidatePaymentRequests(settlePaymentsRequest dbModels.ExecuteTaskRequest) (int, error) {
	var totalCount int = 0

	for i := range paymentRequestTypeToConsolidate {
		processExecutionRequests, err := getExecutionRequestsBasedOnPaymentRequestType(
			settlePaymentsRequest,
			enums.GetPaymentRequestTypeEnum(paymentRequestTypeToConsolidate[i]),
			enums.New.EnumIndex())

		totalCount = totalCount + len(processExecutionRequests)

		if len(processExecutionRequests) == 0 {
			repositoryLog.Info(context.Background(), "Payment-Executor - No payments to consolidate for payment request type: %s", paymentRequestTypeToConsolidate[i])
			return totalCount, nil
		}

		if err != nil {
			repositoryLog.Error(context.Background(), err, "Payment-Executor - error in getting the list of payments to consolidate")
		}

		// Update the status from NEW to Consolidated as soon as the records are fetched from execution_request table
		err = updateExecutionRequestStatus(processExecutionRequests, enums.Consolidated)

		if err != nil {
			return totalCount, err
		}

		// Create a map of execution requests
		// where key is the account number and routing number. e.g. Acc1:ABA1
		// value is the list of execution requests and the consolidated amount for that key. Example as below:
		// map (Acc1:ABA1, list[(dbModels.ExecutionRequest) & ConsolidatedAmount])
		// map (Acc2:ABA2, list[(dbModels.ExecutionRequest) & ConsolidatedAmount])
		// map (Acc3:ABA3, list[(dbModels.ExecutionRequest) & ConsolidatedAmount])
		processExecutionRequestsMap := getExecutionRequestsMapWithConsolidatedAmount(processExecutionRequests)

		// Loop through the above list of aggregated payments to insert
		// one row for consolidated execution request
		// for each acc number and routing number combination for each execution request type
		for _, consolidatedExecutionRequest := range processExecutionRequestsMap {

			/*
				Step # 1- Get the ConsolidatedId first from the consolidated_request table
				Step # 2- If the ConsolidatedId is not found, then insert the record in consolidated_request table
				Step # 3- Else update the record with the new amount in consolidated_request table
			*/

			// Step # 1
			paymentsConsolidatedInfo, err := getPaymentsConsolidatedInfoFromDb(consolidatedExecutionRequest.ExecutionRequestCollection[0])

			if err != nil {
				return totalCount, err
			}

			// Step # 2 & Step # 3
			if paymentsConsolidatedInfo.ConsolidatedId == 0 {
				paymentsConsolidatedInfo.ConsolidatedId, err = insertConsolidatedRequestData(&consolidatedExecutionRequest.ExecutionRequestCollection[0], consolidatedExecutionRequest.ConsolidatedAmount)
				if err != nil {
					return totalCount, err
				}

			} else {

				updatedConsolidatedAmount := paymentsConsolidatedInfo.Amount + consolidatedExecutionRequest.ConsolidatedAmount

				err = updateConsolidatedAmount(paymentsConsolidatedInfo.ConsolidatedId, updatedConsolidatedAmount)
				if err != nil {
					return totalCount, err
				}
			}

			err = updateConsolidatedIdInExecutionRequest(consolidatedExecutionRequest, paymentsConsolidatedInfo.ConsolidatedId)
			if err != nil {
				return totalCount, err
			}
		}
	}

	return totalCount, nil
}

func UpdateSelltlementIdentifer(settlementData []dbModels.SettlementPayment) error {

	valueStrings, valueArgs := prepareUpdateValues(settlementData)

	var updateSettlementIdQuery = `WITH updates(consolidatedId, settlementIdentifier) AS (
		values %s
	)
	UPDATE public.consolidated_request
	SET "SettlementIdentifier" = u.settlementIdentifier
	FROM updates u
	WHERE "ConsolidatedId" = u.consolidatedId::bigint`

	updateSettlementIdQuery = fmt.Sprintf(updateSettlementIdQuery, strings.Join(valueStrings, ","))

	rows, err := database.NewDbContext().Database.Query(context.Background(), updateSettlementIdQuery, valueArgs...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error in updating the Selltlement Identifer of consolidated_request")
		return err
	}

	defer rows.Close()

	return nil
}

// Below method aggregates the execution requests based on the account number and routing number
// and returns a map of execution requests with the consolidated amount
func getExecutionRequestsMapWithConsolidatedAmount(processExecutionRequests []*dbModels.ExecutionRequest) map[string]*dbModels.ListExecutionRequest {

	paymentMap := make(map[string]*dbModels.ListExecutionRequest)

	for _, eachPaymentRq := range processExecutionRequests {

		var accountIdentifier string = eachPaymentRq.AccountIdentifier
		var RoutingNumber string = eachPaymentRq.RoutingNumber

		var paymentRequestType string = strconv.Itoa(eachPaymentRq.PaymentRequestType.EnumIndex())

		// TODO - Do some thinking if we want to hash the data (Not Day 1)
		var key = paymentRequestType + ":" + accountIdentifier + ":" + RoutingNumber

		if _, found := paymentMap[key]; found {
			paymentMap[key].ExecutionRequestCollection = append(paymentMap[key].ExecutionRequestCollection, *eachPaymentRq)
			paymentMap[key].ConsolidatedAmount += eachPaymentRq.Amount
		} else {
			paymentMap[key] = &dbModels.ListExecutionRequest{
				ExecutionRequestCollection: []dbModels.ExecutionRequest{*eachPaymentRq},
				ConsolidatedAmount:         eachPaymentRq.Amount,
			}
		}
	}

	return paymentMap
}

// Insert the records in consolidated table for each execution request type for the first time for a particular payment date
func insertConsolidatedRequestData(executionRequest *dbModels.ExecutionRequest, consolidatedAmount float32) (int64, error) {

	insertConsolidatedRequestQuery :=
		`INSERT INTO public.consolidated_request (
			"AccountIdentifier",
			"RoutingNumber",
			"Amount",
			"PaymentDate",
			"PaymentRequestType",
			"PaymentExtendedData",
			"RetryCount",
			"Status",
			"SettlementIdentifier",
			"CreatedDate",
			"CreatedBy",
			"UpdatedDate",
			"UpdatedBy",
			"Last4AccountIdentifier"
		)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING "ConsolidatedId"`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		insertConsolidatedRequestQuery,
		executionRequest.AccountIdentifier,
		executionRequest.RoutingNumber,
		consolidatedAmount,
		executionRequest.PaymentDate,
		executionRequest.PaymentRequestType.EnumIndex(),
		executionRequest.PaymentExtendedData,
		0,
		enums.New.EnumIndex(),
		nil,
		time.Now(),
		user,
		time.Now(),
		user,
		executionRequest.Last4AccountIdentifier,
	)

	var consolidatedId int64
	err := row.Scan(&consolidatedId)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error inserting records to consolidated_request table")
		return 0, err
	} else {
		repositoryLog.Info(context.Background(), "Payment-Executor - inserted records in consolidated_request table for ConsolidatedId: %d", consolidatedId)
	}

	return consolidatedId, nil
}

func updateConsolidatedAmount(retrievedConsolidatedId int64, consolidatedAmount float32) error {

	updateConsolidatedAmountQuery :=
		`UPDATE public.consolidated_request
		SET
			"Amount" = $1,
			"UpdatedDate" = $2
		WHERE
			"ConsolidatedId" = $3`

	_, err := database.NewDbContext().Database.Exec(
		context.Background(),
		updateConsolidatedAmountQuery,
		consolidatedAmount,
		time.Now(),
		retrievedConsolidatedId)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error updating records in consolidated_request table for ConsolidatedId: %d", retrievedConsolidatedId)
		return err
	} else {
		repositoryLog.Info(context.Background(), "Payment-Executor - updated records in consolidated_request table for ConsolidatedId: %d", retrievedConsolidatedId)
	}

	return nil
}

// Below method queries the database to retrieve payments based on the execution request type in batches
func getExecutionRequestsBasedOnPaymentRequestType(
	settlePaymentsRequest dbModels.ExecuteTaskRequest, paymentRequestTypeEnum enums.PaymentRequestType, paymentStatus int) ([]*dbModels.ExecutionRequest, error) {

	var processExecutionRequests []*dbModels.ExecutionRequest

	// Query the database to get the payments in NEW status
	// to consolidate from execution_request table
	rows, err := getPaymentsToConsolidateFromDb(settlePaymentsRequest, paymentStatus, paymentRequestTypeEnum)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - get payments to consolidate from db failed")
		return nil, err
	}

	for rows.Next() {
		var executionRequest dbModels.ExecutionRequest
		err := rows.Scan(
			&executionRequest.ExecutionRequestId,
			&executionRequest.TenantId,
			&executionRequest.PaymentId,
			&executionRequest.ConsolidatedId,
			&executionRequest.AccountIdentifier,
			&executionRequest.RoutingNumber,
			&executionRequest.Amount,
			&executionRequest.PaymentDate,
			&executionRequest.PaymentFrequency,
			&executionRequest.TransactionType,
			&executionRequest.PaymentRequestType,
			&executionRequest.PaymentMethodType,
			&executionRequest.PaymentExtendedData,
			&executionRequest.Status,
			&executionRequest.Last4AccountIdentifier)

		if err != nil {
			repositoryLog.Error(context.Background(), err, "Payment-Executor - error in scanning execution request")
			return nil, err
		}

		processExecutionRequests = append(processExecutionRequests, &executionRequest)
	}

	return processExecutionRequests, nil
}

// getPaymentsToConsolidateFromDb queries the database to retrieve payments that need to be consolidated.
func getPaymentsToConsolidateFromDb(settlePaymentsRequest dbModels.ExecuteTaskRequest, newPaymentStatus int, paymentRequestTypeEnum enums.PaymentRequestType) (pgx.Rows, error) {

	var rows pgx.Rows
	var err error

	// TODO - get the batch count value from config manager
	var batchCount int = 5000

	getPaymentsToConsolidateQuery :=
		`SELECT 
			"ExecutionRequestId",
			"TenantId",
			"PaymentId",
			COALESCE("ConsolidatedId", 0),
			"AccountIdentifier",
			"RoutingNumber",
			"Amount"::numeric::float,
			"PaymentDate",
			"PaymentFrequency",
			"TransactionType",
			"PaymentRequestType",
			"PaymentMethodType",
			"PaymentExtendedData",
			"Status",
			"Last4AccountIdentifier"
		FROM public.execution_request
		WHERE 
			"PaymentDate" <= $1
			AND "Status"=$2  
			AND "PaymentRequestType"=$3 
		Order By 1 ASC
		LIMIT $4`

	rows, err = database.NewDbContext().Database.Query(
		context.Background(),
		getPaymentsToConsolidateQuery,
		settlePaymentsRequest.TaskDate,
		newPaymentStatus,
		paymentRequestTypeEnum,
		batchCount)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error in fetching payments to consolidate from payment_executor")
		return nil, err
	}

	return rows, nil
}

// getPaymentsConsolidatedInfoFromDb queries the database to retrieve consolidated payment information
// based on the payment request type, account number, routing number and payment date.
func getPaymentsConsolidatedInfoFromDb(paymentsConsolidated dbModels.ExecutionRequest) (dbModels.ConsolidatedRequest, error) {

	var paymentsConsolidatedModel dbModels.ConsolidatedRequest

	getPaymentsConsolidatedQuery :=
		`SELECT 
			"ConsolidatedId", 
			"AccountIdentifier", 
			"RoutingNumber",
			"PaymentRequestType", 
			"Amount"::numeric::float, 
			"PaymentDate", 
			"PaymentExtendedData"
		FROM 
			public.consolidated_request 
		WHERE 
			"PaymentDate" = $1 
			AND "AccountIdentifier" = $2 
			AND "RoutingNumber" = $3
			AND "PaymentRequestType" = $4
			AND "Status" not in ($5, $6)`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentsConsolidatedQuery,
		paymentsConsolidated.PaymentDate,
		paymentsConsolidated.AccountIdentifier,
		paymentsConsolidated.RoutingNumber,
		paymentsConsolidated.PaymentRequestType.EnumIndex(),
		enums.InProgress.EnumIndex(),
		enums.Completed.EnumIndex())

	err := row.Scan(&paymentsConsolidatedModel.ConsolidatedId,
		&paymentsConsolidatedModel.AccountIdentifier,
		&paymentsConsolidatedModel.RoutingNumber,
		&paymentsConsolidatedModel.PaymentRequestType,
		&paymentsConsolidatedModel.Amount,
		&paymentsConsolidatedModel.PaymentDate,
		&paymentsConsolidatedModel.PaymentExtendedData)

	if paymentsConsolidatedModel.ConsolidatedId == 0 {
		return paymentsConsolidatedModel, nil
	}

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error in fetching data from consolidated_request table")
	}

	return paymentsConsolidatedModel, err
}

// Method to update execution_request status based on certain condition
// Condition 1 - Update the status of execution_request to Consolidated for the NEW records which wer fetched to consolidate
func updateExecutionRequestStatus(processExecutionRequests []*dbModels.ExecutionRequest, paymentStatus enums.PaymentStatus) error {

	var updateExecutionRequestStatusQuery = `UPDATE public.execution_request SET "Status" = $%d WHERE "ExecutionRequestId" IN (%s)`

	executionRequestIds := getExecutionRequestIds(processExecutionRequests)

	placeholders := make([]string, len(executionRequestIds))
	for i := range executionRequestIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	updateExecutionRequestStatusQuery = fmt.Sprintf(updateExecutionRequestStatusQuery, len(executionRequestIds)+1, strings.Join(placeholders, ","))

	args := make([]interface{}, len(executionRequestIds)+1)

	for i, id := range executionRequestIds {
		args[i] = id
	}
	args[len(executionRequestIds)] = paymentStatus.EnumIndex()

	rows, err := database.NewDbContext().Database.Query(context.Background(), updateExecutionRequestStatusQuery, args...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error in updating the status of execution_request to Consolidated for the execution request ids")
		return err
	}
	defer rows.Close()

	return nil
}

func updateConsolidatedIdInExecutionRequest(consolidatedPaymentRequest *dbModels.ListExecutionRequest, consolidatedId int64) error {

	updateConsolidatedIdInExecutionRequestQuery := `UPDATE public.execution_request SET "ConsolidatedId" = $%d WHERE "ExecutionRequestId" IN (%s)`

	var executionRequestIds = getExecutionRequestIdsFromCollection(consolidatedPaymentRequest.ExecutionRequestCollection)

	placeholders := make([]string, len(executionRequestIds))
	for i := range executionRequestIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	updateConsolidatedIdInExecutionRequestQuery = fmt.Sprintf(updateConsolidatedIdInExecutionRequestQuery, len(executionRequestIds)+1, strings.Join(placeholders, ","))

	args := make([]interface{}, len(executionRequestIds)+1)

	for i, id := range executionRequestIds {
		args[i] = id
	}

	args[len(executionRequestIds)] = consolidatedId

	rows, err := database.NewDbContext().Database.Query(context.Background(), updateConsolidatedIdInExecutionRequestQuery, args...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Payment-Executor - error in updating the consolidatedId in execution_request table")
		return err
	}
	defer rows.Close()

	return nil
}

func getExecutionRequestIdsFromCollection(executionRequests []dbModels.ExecutionRequest) []int64 {
	var executionRequestIds []int64

	for _, item := range executionRequests {
		executionRequestIds = append(executionRequestIds, item.ExecutionRequestId)
	}

	return executionRequestIds
}

func getExecutionRequestIds(processExecutionRequests []*dbModels.ExecutionRequest) []int64 {

	var executionRequestIds []int64

	for _, item := range processExecutionRequests {
		executionRequestIds = append(executionRequestIds, item.ExecutionRequestId)
	}

	return executionRequestIds
}

func prepareUpdateValues(settlementData []dbModels.SettlementPayment) ([]string, []interface{}) {
	var valueStrings []string
	var valueArgs []interface{}

	for i, record := range settlementData {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d::bigint, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, record.ConsolidatedId, record.SettlementIdentifier)
	}

	return valueStrings, valueArgs
}
