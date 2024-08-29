package repository

import (
	"context"
	"fmt"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	dbModels "geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
)

var repositoryLog = logging.GetLogger("payment-executor-repository")

const (
	user                          = "payment-executor"
	paymentIdAlreadyExistsMessage = "PaymentId already exists in execution_request table. PaymentId: %d"
)

type ExecutionRequestRepositoryInterface interface {
	ExecuteRequests(executionRequest commonMessagingModels.ProcessPaymentRequest) error
	GetExecutionRequestsByConsolidatedId(consolidatedId int64, status *int) ([]*dbModels.ExecutionRequest, error)
}

type ExecutionRequestRepository struct {
	// Add any necessary dependencies or fields here
}

func (ExecutionRequestRepository) ExecuteRequests(executionRequest commonMessagingModels.ProcessPaymentRequest) error {
	repositoryLog.Info(context.Background(), "Execute Payment Request - Insert Record into execution_request table")

	if isIncomingPaymentRequestAlreadyExist(executionRequest.PaymentId) {
		repositoryLog.Info(context.Background(), paymentIdAlreadyExistsMessage, executionRequest.PaymentId)
		return fmt.Errorf(paymentIdAlreadyExistsMessage, executionRequest.PaymentId)
	}

	return insertExecutionRequestData(executionRequest)
}

func (ExecutionRequestRepository) GetExecutionRequestsByConsolidatedId(consolidatedId int64, status *int) ([]*dbModels.ExecutionRequest, error) {
	repositoryLog.Info(context.Background(), "Fetching Execution Requests for ConsolidatedId: %d", consolidatedId)

	query := `SELECT 
				"ExecutionRequestId",	
				"PaymentId",
				COALESCE("SettlementIdentifier", ''),
				"Amount"::numeric::float,
				"PaymentDate",
				"Last4AccountIdentifier",
				"PaymentRequestType",
				"PaymentMethodType"
			FROM public.execution_request
			WHERE 
				"ConsolidatedId" = $1
				AND "Status"= $2`

	rows, err := database.NewDbContext().Database.Query(
		context.Background(),
		query,
		consolidatedId,
		status,
	)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Error fetching execution requests for ConsolidatedId: %d", consolidatedId)
		return nil, err
	}
	defer rows.Close()

	var executionRequests []*dbModels.ExecutionRequest

	for rows.Next() {
		var executionRequest dbModels.ExecutionRequest
		err := rows.Scan(
			&executionRequest.ExecutionRequestId,
			&executionRequest.PaymentId,
			&executionRequest.SettlementIdentifier,
			&executionRequest.Amount,
			&executionRequest.PaymentDate,
			&executionRequest.Last4AccountIdentifier,
			&executionRequest.PaymentRequestType,
			&executionRequest.PaymentMethodType)

		if err != nil {
			repositoryLog.Error(context.Background(), err, "Error scanning execution requests for ConsolidatedId: %d", consolidatedId)
			return nil, err
		}

		executionRequests = append(executionRequests, &executionRequest)
	}

	return executionRequests, nil
}

func isIncomingPaymentRequestAlreadyExist(paymentId int64) bool {

	isPaymentIdAlreadyExist := false
	getExecutionRequest := `SELECT EXISTS
                             (SELECT 1 
							  FROM public.execution_request 
							  WHERE "PaymentId" = $1)`

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getExecutionRequest,
		paymentId)

	err := row.Scan(&isPaymentIdAlreadyExist)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "isIncomingPaymentRequestAlreadyExist - error fetching data from execution_request table")
	}

	return isPaymentIdAlreadyExist

}

func insertExecutionRequestData(executionRequest commonMessagingModels.ProcessPaymentRequest) error {

	executionRequestQuery := `INSERT INTO public.execution_request 
	("TenantId",
	"PaymentId",
	"ConsolidatedId",
	"AccountIdentifier",
	"RoutingNumber",
	"Amount",
	"PaymentDate",
	"PaymentFrequency",
	"TransactionType",
	"PaymentRequestType",
	"PaymentMethodType",
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
	 VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := database.NewDbContext().Database.Exec(
		context.Background(),
		executionRequestQuery,
		executionRequest.TenantId,
		executionRequest.PaymentId,
		nil,
		executionRequest.AccountIdentifier,
		executionRequest.RoutingNumber,
		executionRequest.Amount,
		time.Time(executionRequest.PaymentDate),
		enums.GetPaymentFrequecyEnum(executionRequest.PaymentFrequency),
		enums.GetTransactionTypeEnum(executionRequest.TransactionType),
		enums.GetPaymentRequestTypeEnum(executionRequest.PaymentRequestType),
		enums.GetPaymentMethodTypeEnumFromString(executionRequest.PaymentMethodType),
		executionRequest.PaymentExtendedData,
		0,
		enums.New.EnumIndex(),
		nil, // TODO - Will discuss about settlement identifier to be a GUID or some system generated number which can be used in files
		time.Now(),
		user,
		time.Now(),
		user,
		executionRequest.Last4AccountIdentifier,
	)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error inserting data into execution_request table")
		return err
	}

	return nil
}
