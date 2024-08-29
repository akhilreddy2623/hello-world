package inbound_settlement_postprocessor

import (
	"context"
	"encoding/json"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/repository"
	"github.com/jackc/pgx/v5"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	dbModels "geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
)

const (
	updatedByUserId = "ach_onetime_ack_postprocessor"
)

var log = logging.GetLogger("payment-executor-postprocessors")

/*
Step 1 - Fetch records from execution_request in a batch say of 5000
Step 2 - Start processing these records one by one

	Step 2(a) - Update the record status as Completed in execution_request table
	Step 2(b) - Publish the payment complete event to topic paymentPlatform.executePaymentResponses

Step 3 - If the single record is processed successfully then commit the transaction
Step 4 - If the publishing of single record fails then rollback the transaction
*/

func AchOnetimeAckPostProcessor() error {

	publishingPaymentError := PublishPaymentCompleteEvent()
	if publishingPaymentError != nil {
		log.Error(context.Background(), publishingPaymentError, "Payment Executor - Error in publishing payment complete event")
	}

	publishingConsolidatedError := PublishConsolidatedAmountEvent()
	if publishingConsolidatedError != nil {
		log.Error(context.Background(), publishingConsolidatedError, "Payment Executor - Error in publishing consolidated amount event")
	}

	if publishingPaymentError != nil {
		return publishingPaymentError
	}
	if publishingConsolidatedError != nil {
		return publishingConsolidatedError
	}

	return nil
}

func PublishConsolidatedAmountEvent() error {
	log.Info(context.Background(), "Payment Executor - Publishing Payment Consolidated Event. Entering...")

	executeConsolidateRequestsData, err := getConsolidationRequestsToMarkAsComplete(enums.InProgress)

	if err != nil {
		return err
	}

	if len(executeConsolidateRequestsData) == 0 {
		log.Info(context.Background(), "Payment Executor - All consolidated requests are maked as completed. Exiting...")
		return nil
	}

	if err := publishConsolidatedAmountEventAndMarkRecordsAsComplete(executeConsolidateRequestsData); err != nil {
		return err
	}

	return nil
}

func PublishPaymentCompleteEvent() error {
	log.Info(context.Background(), "Payment Executor - Publishing Payment Complete Event")

	// Open a transaction and get payments in batches
	// For each payment, publish a payment complete event
	// Update the status of payment to be completed
	// If all payments are processed successfully, commit the transaction

	var batchCount int = 5000

	for {
		executePaymentRequestsData, err := getExecutionRequestsToMarkAsComplete(enums.InProgress, batchCount)

		if err != nil {
			return err
		}

		if len(executePaymentRequestsData) == 0 {
			log.Info(context.Background(), "Payment Executor - All payment requests are maked as completed. Exiting...")
			return nil
		}

		err = publishPaymentCompleteEventAndMarkRecordsAsComplete(executePaymentRequestsData)
		if err != nil {
			return err
		}

	}

}

func publishPaymentCompleteEventAndMarkRecordsAsComplete(executePaymentRequestsData []*dbModels.ExecutionRequest) error {

	for _, item := range executePaymentRequestsData {
		publishPaymentCompleteEventAndMarkRecordsAsCompleteSingle(item)
	}

	return nil
}

func publishPaymentCompleteEventAndMarkRecordsAsCompleteSingle(executePaymentRequestsData *dbModels.ExecutionRequest) error {

	updateExecutionRequestStatusQuery :=
		`UPDATE public.execution_request 
		SET 
			"Status" = $1,
			"UpdatedBy" = $2,
			"UpdatedDate" = $3
		WHERE 
			"ExecutionRequestId" = $4`

	// Open a transaction
	transaction, err := database.NewDbContext().Database.Begin(context.Background())
	defer transaction.Rollback(context.Background())

	if err != nil {
		log.Error(context.Background(), err, "Payment Executor - error in starting transaction for updating execution_request status as Completed")
		return err
	}

	_, err = transaction.Exec(
		context.Background(),
		updateExecutionRequestStatusQuery,
		enums.Completed.EnumIndex(),
		updatedByUserId,
		time.Now(),
		executePaymentRequestsData.ExecutionRequestId)

	if err != nil {
		log.Error(context.Background(), err, "Payment Executor - error updating records in consolidated_request table for ExecutionRequestId: %d", executePaymentRequestsData.ExecutionRequestId)
		return err
	} else {
		log.Info(context.Background(), "Payment Executor - updated records in consolidated_request table for ExecutionRequestId: %d", executePaymentRequestsData.ExecutionRequestId)

		// Publish Payment Complete Event
		configHandler := commonFunctions.GetConfigHandler()

		var topicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")

		messageToPublish := getMessageToPublish(executePaymentRequestsData)

		err = messaging.KafkaPublish(topicName, messageToPublish)

		if err == nil {
			err = transaction.Commit(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "Payment Executor - error in commiting transaction for paymentId '%d'", executePaymentRequestsData.PaymentId)
			}

			// Publish Payment Settled Event
			messageToPublish := getPaymentSettledMessageAndPublish(executePaymentRequestsData)
			var PaymentEventTopicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
			err = messaging.KafkaPublish(PaymentEventTopicName, messageToPublish)
			if err != nil {
				log.Error(context.Background(), err, "Payment Executor - error in publishing message as Completed for PaymentId '%d'", executePaymentRequestsData.PaymentId)
			}

		} else {
			err = transaction.Rollback(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "Payment Executor - error in publishing message as Completed for PaymentId '%d'", executePaymentRequestsData.PaymentId)
			}
		}
	}

	return nil
}

func getMessageToPublish(executePaymentRequestsData *dbModels.ExecutionRequest) string {

	message := ExecutePaymentResponse{
		Version:                1,
		PaymentId:              executePaymentRequestsData.PaymentId,
		Status:                 enums.Completed.String(),
		SettlementIdentifier:   executePaymentRequestsData.SettlementIdentifier,
		PaymentDate:            executePaymentRequestsData.PaymentDate,
		Amount:                 executePaymentRequestsData.Amount,
		Last4AccountIdentifier: executePaymentRequestsData.Last4AccountIdentifier,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in marshalling message with PaymentId '%d'", executePaymentRequestsData.PaymentId)
	}
	return string(messageJson)
}

func getPaymentSettledMessageAndPublish(executePaymentRequestsData *dbModels.ExecutionRequest) string {

	message := commonMessagingModels.PaymentEvent{
		Version:            1,
		PaymentId:          executePaymentRequestsData.PaymentId,
		Amount:             executePaymentRequestsData.Amount,
		PaymentDate:        executePaymentRequestsData.PaymentDate,
		PaymentRequestType: executePaymentRequestsData.PaymentRequestType,
		PaymentMethodType:  executePaymentRequestsData.PaymentMethodType,
		EventType:          enums.PaymentSettled,
		EventDateTime:      time.Now(),
		SettlementAmount:   executePaymentRequestsData.Amount,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in marshalling message with PaymentId '%d'", executePaymentRequestsData.PaymentId)
	}
	return string(messageJson)
}

func getExecutionRequestsToMarkAsComplete(paymentStatus enums.PaymentStatus, batchCount int) ([]*dbModels.ExecutionRequest, error) {

	var executePaymentRequests []*dbModels.ExecutionRequest

	rows, err := executePaymentRequestsFromDb(paymentStatus, batchCount)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var executionRequest dbModels.ExecutionRequest
		err := rows.Scan(
			&executionRequest.ExecutionRequestId,
			&executionRequest.PaymentId,
			&executionRequest.SettlementIdentifier,
			&executionRequest.Amount,
			&executionRequest.PaymentDate,
			&executionRequest.Last4AccountIdentifier)

		if err != nil {
			log.Error(context.Background(), err, "Payment-Executor - error in scanning payments to update status to Complete from execution_request")
			return nil, err
		}

		executePaymentRequests = append(executePaymentRequests, &executionRequest)
	}

	return executePaymentRequests, nil
}

func executePaymentRequestsFromDb(paymentStatus enums.PaymentStatus, batchCount int) (pgx.Rows, error) {

	var rows pgx.Rows
	var err error

	getExecutionRequestsQuery :=
		`SELECT 
			"ExecutionRequestId",	
			"PaymentId",
			COALESCE("SettlementIdentifier", ''),
			"Amount"::numeric::float,
			"PaymentDate",
			"Last4AccountIdentifier"
		FROM public.execution_request
		WHERE 
			"PaymentDate" <= $1
			AND "Status"= $2 
		Order By 1 ASC
		LIMIT $3`

	rows, err = database.NewDbContext().Database.Query(
		context.Background(),
		getExecutionRequestsQuery,
		time.Now(),
		paymentStatus,
		batchCount)

	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in fetching payments to update status to Complete from execution_request")
		return nil, err
	}

	return rows, nil
}

func getConsolidationRequestsToMarkAsComplete(paymentStatus enums.PaymentStatus) ([]*dbModels.ConsolidatedRequest, error) {

	var consolidatePaymentRequests []*dbModels.ConsolidatedRequest

	rows, err := getConsolidateRequestsFromDb(paymentStatus)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var consolidateRequest dbModels.ConsolidatedRequest
		err := rows.Scan(
			&consolidateRequest.ConsolidatedId,
			&consolidateRequest.SettlementIdentifier,
			&consolidateRequest.Amount,
			&consolidateRequest.PaymentDate,
			&consolidateRequest.Last4AccountIdentifier)

		if err != nil {
			log.Error(context.Background(), err, "Payment-Executor - error in scanning payments to update status to Complete from consolidated_request")
			return nil, err
		}

		consolidatePaymentRequests = append(consolidatePaymentRequests, &consolidateRequest)
	}

	return consolidatePaymentRequests, nil
}

func getConsolidateRequestsFromDb(paymentStatus enums.PaymentStatus) (pgx.Rows, error) {

	var rows pgx.Rows
	var err error

	getExecutionRequestsQuery :=
		`SELECT 
			"ConsolidatedId",	
			COALESCE("SettlementIdentifier", ''),
			"Amount"::numeric::float,
			"PaymentDate",
			"Last4AccountIdentifier"
		FROM public.consolidated_request
		WHERE 
			"PaymentDate" <= $1
			AND "Status"= $2 
		Order By 1 ASC`

	rows, err = database.NewDbContext().Database.Query(
		context.Background(),
		getExecutionRequestsQuery,
		time.Now(),
		paymentStatus)

	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in fetching consolidate payments to update status to Complete in consolidated_request")
		return nil, err
	}

	return rows, nil
}

func publishConsolidatedAmountEventAndMarkRecordsAsComplete(executeConsolidateRequestsData []*dbModels.ConsolidatedRequest) error {

	for _, item := range executeConsolidateRequestsData {
		publishConsolidatedCompleteEventAndMarkRecordsAsCompleteSingle(item)
	}

	return nil
}

func publishConsolidatedCompleteEventAndMarkRecordsAsCompleteSingle(item *dbModels.ConsolidatedRequest) error {

	updateConsolidatedRequestStatusQuery :=
		`UPDATE public.consolidated_request 
		SET 
			"Status" = $1,
			"UpdatedBy" = $2,
			"UpdatedDate" = $3
		WHERE 
			"ConsolidatedId" = $4`

	// Open a transaction
	transaction, err := database.NewDbContext().Database.Begin(context.Background())
	defer transaction.Rollback(context.Background())

	if err != nil {
		log.Error(context.Background(), err, "Payment Executor - error in starting transaction for updating consolidated_request status as Completed")
		return err
	}

	_, err = transaction.Exec(
		context.Background(),
		updateConsolidatedRequestStatusQuery,
		enums.Completed.EnumIndex(),
		updatedByUserId,
		time.Now(),
		item.ConsolidatedId)

	if err != nil {
		log.Error(context.Background(), err, "Payment Executor - error updating records in consolidated_request table for ConsolidatedId: %d", item.ConsolidatedId)
		return err
	} else {
		log.Info(context.Background(), "Payment Executor - updated records in consolidated_request table for ConsolidatedId: %d", item.ConsolidatedId)

		// Publish Consolidate Request Complete Event
		configHandler := commonFunctions.GetConfigHandler()
		var topicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.ConsolidatedPayments", "")

		messageToPublish := getConsolidatedMessageToPublish(item)
		err = messaging.KafkaPublish(topicName, messageToPublish)

		if err == nil {
			err = transaction.Commit(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "Payment Executor - error in commiting transaction for ConsolidatedId '%d'", item.ConsolidatedId)
			}
		} else {
			err = transaction.Rollback(context.Background())
			if err != nil {
				log.Error(context.Background(), err, "Payment Executor - error in publishing message as Completed for ConsolidatedId '%d'", item.ConsolidatedId)
			}
		}

		// Fetch and publish related execution requests
		err = publishExecutionRequestsForConsolidatedId(item.ConsolidatedId)
		if err != nil {
			log.Error(context.Background(), err, "Payment Executor - error in publishing related execution requests for ConsolidatedId '%d'", item.ConsolidatedId)
			return err
		}

	}

	return nil
}

func getConsolidatedMessageToPublish(consolidatedRequest *dbModels.ConsolidatedRequest) string {

	message := commonMessagingModels.ConsolidatedExecutePaymentResponse{
		Version:                1,
		ConsolidatedId:         consolidatedRequest.ConsolidatedId,
		Status:                 enums.Completed.String(),
		SettlementIdentifier:   consolidatedRequest.SettlementIdentifier,
		PaymentDate:            consolidatedRequest.PaymentDate,
		Amount:                 consolidatedRequest.Amount,
		Last4AccountIdentifier: consolidatedRequest.Last4AccountIdentifier,
		PaymentRequestType:     consolidatedRequest.PaymentRequestType.String(),
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in marshalling message with ConsolidatedId '%d'", consolidatedRequest.ConsolidatedId)
	}
	return string(messageJson)
}

type ExecutePaymentResponse struct {
	Version                int
	PaymentId              int64
	Status                 string
	SettlementIdentifier   string
	PaymentDate            time.Time
	Amount                 float32
	Last4AccountIdentifier string
}

func publishExecutionRequestsForConsolidatedId(consolidatedId int64) error {
	// Initialize the repository
	executionRequestRepo := repository.ExecutionRequestRepository{}

	// Fetch the execution requests by ConsolidatedId
	status := enums.Consolidated.EnumIndex()
	executionRequests, err := executionRequestRepo.GetExecutionRequestsByConsolidatedId(consolidatedId, &status)
	if err != nil {
		log.Error(context.Background(), err, "Error fetching execution requests for ConsolidatedId: %d", consolidatedId)
		return err
	}

	for _, executionRequest := range executionRequests {
		messageToPublish := getMessageToPublish(executionRequest)
		configHandler := commonFunctions.GetConfigHandler()
		var topicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")

		err = messaging.KafkaPublish(topicName, messageToPublish)
		if err != nil {
			log.Error(context.Background(), err, "Payment Executor - error in publishing execution request for PaymentId '%d'", executionRequest.PaymentId)
			return err
		}

		// Publish Payment Settled Event
		messageToPublish = getPaymentSettledMessageAndPublish(executionRequest)
		var paymentEventTopicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
		err = messaging.KafkaPublish(paymentEventTopicName, messageToPublish)
		if err != nil {
			log.Error(context.Background(), err, "Payment Executor - error in publishing payment settled event for PaymentId '%d'", executionRequest.PaymentId)
			return err
		}
	}

	return nil
}
