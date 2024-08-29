package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/repository"

	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	mutexlock "geico.visualstudio.com/Billing/plutus/lock"
	dbModels "geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
	settlement "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound"
)

const (
	invalidField              = "invalidField"
	consolidatePaymentRqError = "consolidatePaymentRqError"
	errorInSettlementProcess  = "errorInSettlementProcess"
)

var ExecuteTaskResponseTopic string

type ProcessTaskHandlerInterface interface {
	PublishTaskResponse(
		executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
		taskStatus enums.TaskStatus,
		executeTaskResponseTopic string,
		count int) error

	PublishTaskErrorResponse(
		executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
		errorType string,
		err error,
		executetaskResponseTopic string,
		count int) error
}

type ExecuteTaskHandlerStruct struct {
}

// Function to read message for settle payment request and kick off the process
var PaymentSettlementTaskHandler = func(ctx context.Context, message *kafkamessaging.Message) error {

	log.Info(ctx, "PaymentExecutor - Received message in paymentplatform.internal.executetaskrequests topic to settle payments: '%s'", *message.Body)

	var processTaskHandler ProcessTaskHandlerInterface = ExecuteTaskHandlerStruct{}
	paymentSettlementTaskRequest := commonMessagingModels.ExecuteTaskRequest{}

	err := json.Unmarshal([]byte(*message.Body), &paymentSettlementTaskRequest)
	if err != nil {
		log.Error(ctx, err, "Error in unmarshalling paymentSettlementTaskRequest message")
		return err
	}

	var paymentSettlementTaskRequestRepository repository.ConsolidatedRequestRepositoryInterface = repository.ConsolidatedRequestRepository{}
	return ExecutePaymentRequestSettlementTask(ctx, paymentSettlementTaskRequest, paymentSettlementTaskRequestRepository, processTaskHandler)
}

func ExecutePaymentRequestSettlementTask(
	ctx context.Context,
	paymentSettlementTaskRequest commonMessagingModels.ExecuteTaskRequest,
	consolidatedRequestRepository repository.ConsolidatedRequestRepositoryInterface,
	processTaskHandler ProcessTaskHandlerInterface) error {

	paymentSettlementTaskRq, validationError := validateAndGetTaskRequest(paymentSettlementTaskRequest)

	// TODO - enums.SettleACHPayments may be changed to SettlePayments?
	if paymentSettlementTaskRq != nil && (paymentSettlementTaskRq.Component != enums.Executor || paymentSettlementTaskRq.TaskName != enums.SettleACHPayments) {
		return nil
	}

	if validationError != nil {
		errorResponse := processTaskHandler.PublishTaskErrorResponse(paymentSettlementTaskRequest, invalidField, validationError, ExecuteTaskResponseTopic, 0)
		if errorResponse != nil {
			return errorResponse
		}
		log.Error(ctx, validationError, "error validating the payment settlement task request")
		return validationError
	}

	err := processTaskHandler.PublishTaskResponse(paymentSettlementTaskRequest, enums.TaskInprogress, ExecuteTaskResponseTopic, 0)
	if err != nil {
		return err
	}

	var count int
	var errInfo error

	// Create a new instance of the PSqlLock.
	consolidatePaymentRequestsLock, _ := mutexlock.NewPSqlLockService()

	// Run the task with the lock.
	lockErr := consolidatePaymentRequestsLock.RunWithLock(
		ctx,
		func() error {
			count, errInfo = consolidatedRequestRepository.ConsolidatePaymentRequests(*paymentSettlementTaskRq)
			return errInfo
		},
		mutexlock.ResourceLockId(enums.ConsolidatePaymentRequestsLock))

	if lockErr != nil {
		err := processTaskHandler.PublishTaskErrorResponse(paymentSettlementTaskRequest, consolidatePaymentRqError, lockErr, ExecuteTaskResponseTopic, count)
		if err != nil {
			return err
		}
		log.Error(ctx, lockErr, "error processing settle payments task")
		return lockErr
	}

	// Create outbound file for ach one time payment requests
	err = processSettlement(ctx, *paymentSettlementTaskRq)
	if err != nil {
		errorResponse := processTaskHandler.PublishTaskErrorResponse(paymentSettlementTaskRequest, errorInSettlementProcess, err, ExecuteTaskResponseTopic, count)
		if errorResponse != nil {
			return errorResponse
		}
	}

	err = processTaskHandler.PublishTaskResponse(paymentSettlementTaskRequest, enums.TaskCompleted, ExecuteTaskResponseTopic, count)
	if err != nil {
		return err
	}

	return nil
}

func processSettlement(ctx context.Context, executeTaskRequest dbModels.ExecuteTaskRequest) error {

	executionParameters :=
		dbModels.ExecutionParameters{
			PaymentMethodType:  executeTaskRequest.ExecutionParameters.PaymentMethodType,
			PaymentRequestType: executeTaskRequest.ExecutionParameters.PaymentRequestType,
			PaymentFrequency:   executeTaskRequest.ExecutionParameters.PaymentFrequency,
		}

	// Create a new instance of the PSqlLock for file processing
	createACHOneTimeFileLock, _ := mutexlock.NewPSqlLockService()

	err := createACHOneTimeFileLock.RunWithLock(
		ctx,
		func() error {
			settlement.DataHubPaymentEventsTopic = DataHubPaymentEventsTopic
			return settlement.ProcessOutbound(executeTaskRequest.TaskDate, executionParameters)
		},
		mutexlock.ResourceLockId(enums.CreateACHOneTimeFileLock))

	if err != nil {
		log.Error(ctx, err, "error processing create ach one time file task")
		return err
	}
	return nil
}

func validateAndGetTaskRequest(paymentSettlementTaskRequest commonMessagingModels.ExecuteTaskRequest) (*dbModels.ExecuteTaskRequest, error) {

	executeTaskReq := dbModels.ExecuteTaskRequest{
		Version:         paymentSettlementTaskRequest.Version,
		Component:       enums.GetComponentTypeEnum(paymentSettlementTaskRequest.Component),
		TaskName:        enums.GetTaskTypeEnum(paymentSettlementTaskRequest.TaskName),
		TaskDate:        time.Time(paymentSettlementTaskRequest.TaskDate),
		TaskExecutionId: paymentSettlementTaskRequest.TaskExecutionId,
	}

	executionParam, err := validateAndGetExecutionParameters(paymentSettlementTaskRequest.ExecutionParameters)
	if err != nil {
		return nil, err
	}

	executeTaskReq.ExecutionParameters = *executionParam
	return &executeTaskReq, nil
}

func (ExecuteTaskHandlerStruct) PublishTaskResponse(
	executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
	taskStatus enums.TaskStatus,
	schedulerTaskResponseTopic string,
	count int) error {
	executeTaskResponse, err := getExecuteTaskResponseJson(executeTaskRequest, taskStatus, count, nil)
	if err != nil {
		return err
	}
	err = messaging.KafkaPublish(schedulerTaskResponseTopic, *executeTaskResponse)
	return err
}

func (ExecuteTaskHandlerStruct) PublishTaskErrorResponse(
	executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
	errorType string,
	err error,
	schedulerTaskResponseTopic string,
	count int) error {
	errorResponse := getErrorResponse(errorType, err)
	executeTaskResponse, err := getExecuteTaskResponseJson(executeTaskRequest, enums.TaskErrored, count, &errorResponse)
	if err != nil {
		return err
	}
	err = messaging.KafkaPublish(schedulerTaskResponseTopic, *executeTaskResponse)
	return err
}

func getErrorResponse(errorType string, err error) commonAppModels.ErrorResponse {
	errorResponse := commonAppModels.ErrorResponse{
		Type:    errorType,
		Message: err.Error(),
	}
	return errorResponse
}

func getExecuteTaskResponseJson(
	executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
	taskStatus enums.TaskStatus,
	processedRecordsCount int,
	errorResponse *commonAppModels.ErrorResponse) (*string, error) {

	executeTaskResponse := commonMessagingModels.ExecuteTaskResponse{
		Version:               executeTaskRequest.Version,
		TaskExecutionId:       executeTaskRequest.TaskExecutionId,
		Status:                taskStatus.String(),
		ProcessedRecordsCount: processedRecordsCount,
		ErrorDetails:          errorResponse,
	}

	executeTaskResponseJson, err := json.Marshal(executeTaskResponse)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal schedulePaymentResponse")
		return nil, err
	}

	executeTaskResponseJsonStr := string(executeTaskResponseJson)
	return &executeTaskResponseJsonStr, nil
}

func validateAndGetExecutionParameters(executionParam commonMessagingModels.ExecutionParameters) (*dbModels.ExecutionParameters, error) {

	executionParameters := dbModels.ExecutionParameters{
		PaymentMethodType:  enums.GetPaymentMethodTypeEnumFromString(executionParam.PaymentMethodType),
		PaymentRequestType: enums.GetPaymentRequestTypeEnum(executionParam.PaymentRequestType),
		PaymentFrequency:   enums.GetPaymentFrequecyEnum(executionParam.PaymentFrequency),
	}

	if executionParameters.PaymentMethodType == enums.NonePaymentMethodType {
		return nil, fmt.Errorf("invalid input payment method type '%s'", executionParam.PaymentMethodType)
	} else if executionParameters.PaymentRequestType == enums.NonePaymentRequestType {
		return nil, fmt.Errorf("invalid input payment request type '%s'", executionParam.PaymentRequestType)
	} else if executionParameters.PaymentFrequency == enums.NonePaymentFrequency {
		return nil, fmt.Errorf("invalid input payment frequency '%s'", executionParam.PaymentFrequency)
	}

	return &executionParameters, nil
}
