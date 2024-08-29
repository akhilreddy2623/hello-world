package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"geico.visualstudio.com/Billing/plutus/api"
	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	appmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/app"
	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

const (
	invalidField               = "invalidField"
	unableToProcessPayment     = "unableToProcessPayment"
	unableToProcessWorkdayFeed = "unableToProcessWorkdayFeed"
)

var log = logging.GetLogger("payment-administrator-handlers")
var ExecuteTaskResponseTopic string
var ExecutePaymentRequestTopic string
var DataHubPaymentEventsTopic string

//go:generate mockery --name ProcessTaskHandlerInterface
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

	CallWorkdayAPI(workdayFeed []*dbmodels.WorkdayFeed) error
}

type ProcessTaskHandler struct {
}

var ProcessTaskRequestHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in paymentplatform.internal.executetaskrequests topic: '%s'", *message.Body)

	var processTaskHandler ProcessTaskHandlerInterface = ProcessTaskHandler{}
	exeuteTaskRequest := commonMessagingModels.ExecuteTaskRequest{}
	if err := json.Unmarshal([]byte(*message.Body), &exeuteTaskRequest); err != nil {
		log.Error(context.Background(), err, "unable to unmrshal exeuteTaskRequest")
		return err
	}

	if enums.GetComponentTypeEnum(exeuteTaskRequest.Component) == enums.Administrator && enums.GetTaskTypeEnum(exeuteTaskRequest.TaskName) == enums.ProcessPayments {
		var paymentRepository repository.PaymentRepositoryInterface = repository.PaymentRepository{}
		return processPaymentTask(exeuteTaskRequest, paymentRepository, processTaskHandler)
	}
	if enums.GetComponentTypeEnum(exeuteTaskRequest.Component) == enums.Administrator && enums.GetTaskTypeEnum(exeuteTaskRequest.TaskName) == enums.SendWorkdayData {
		var workdayRepository repository.WorkdayRepositoryInterface = repository.WorkdayRepository{}
		return sendWorkdayDataTask(exeuteTaskRequest, workdayRepository, processTaskHandler)
	}
	return nil
}

func (ProcessTaskHandler) PublishTaskResponse(
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

func (ProcessTaskHandler) PublishTaskErrorResponse(
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

func (ProcessTaskHandler) CallWorkdayAPI(workdayFeed []*dbmodels.WorkdayFeed) error {
	bearerToken, err := generateBearerTokenForWorkday()
	if err != nil {
		return err
	}

	authorization := fmt.Sprintf("bearer %s", *bearerToken)
	formattedWorkDayFeed := buildWorkdayFeedAccumilatedRequest(workdayFeed)
	workdayRequstUrl := commonFunctions.GetConfigHandler().GetString("PaymentPlatform.Workday.Url", "")
	workdayRetryCount := commonFunctions.GetConfigHandler().GetInt("PaymentPlatform.Workday.RetryCount", 0)
	workdayTimeout := commonFunctions.GetConfigHandler().GetInt("PaymentPlatform.Workday.Timeout", 0)

	timeoutDuration := time.Duration(workdayTimeout * int(time.Second))

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.Workday,
		Url:           workdayRequstUrl,
		Request:       formattedWorkDayFeed,
		Authorization: authorization,
		Timeout:       timeoutDuration,
	}

	apiResponse, err := api.RetryApiCall(api.PostRestApiCall, apiRequest, workdayRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "call failed to %s", workdayRequstUrl)
		return err
	}

	// TODO: Handle actual response from workday
	bodyString := string(apiResponse.Response)

	log.Info(context.Background(), "Response From Workday API: %s", bodyString)
	return nil
}

func processPaymentTask(
	executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
	paymentRepository repository.PaymentRepositoryInterface,
	processTaskHandler ProcessTaskHandlerInterface) error {

	isSendWorkdayDataTask := false
	taskRq, err := validateAndGetTaskRequest(executeTaskRequest, isSendWorkdayDataTask)

	if err != nil {
		processTaskErr := processTaskHandler.PublishTaskErrorResponse(executeTaskRequest, invalidField, err, ExecuteTaskResponseTopic, 0)
		if processTaskErr != nil {
			return processTaskErr
		}
		return err
	}

	err = processTaskHandler.PublishTaskResponse(executeTaskRequest, enums.TaskInprogress, ExecuteTaskResponseTopic, 0)
	if err != nil {
		return err
	}

	repository.ExecutorExecuteRequestTopic = ExecutePaymentRequestTopic
	repository.DataHubPaymentEventsTopic = DataHubPaymentEventsTopic
	count, err := paymentRepository.ProcessPayments(*taskRq)
	if err != nil {
		processTaskErr := processTaskHandler.PublishTaskErrorResponse(executeTaskRequest, unableToProcessPayment, err, ExecuteTaskResponseTopic, *count)
		if processTaskErr != nil {
			return processTaskErr
		}
		return err
	}

	err = processTaskHandler.PublishTaskResponse(executeTaskRequest, enums.TaskCompleted, ExecuteTaskResponseTopic, *count)
	if err != nil {
		return err
	}

	return nil
}

func sendWorkdayDataTask(executeTaskRequest commonMessagingModels.ExecuteTaskRequest,
	workdayRepository repository.WorkdayRepositoryInterface,
	processTaskHandler ProcessTaskHandlerInterface) error {

	isSendWorkdayDataTask := true
	taskRq, err := validateAndGetTaskRequest(executeTaskRequest, isSendWorkdayDataTask)

	if err != nil {
		processTaskErr := processTaskHandler.PublishTaskErrorResponse(executeTaskRequest, invalidField, err, ExecuteTaskResponseTopic, 0)
		if processTaskErr != nil {
			return processTaskErr
		}
		return err
	}

	err = processTaskHandler.PublishTaskResponse(executeTaskRequest, enums.TaskInprogress, ExecuteTaskResponseTopic, 0)
	if err != nil {
		return err
	}

	count, err := processWorkdayFeed(processTaskHandler, workdayRepository, *taskRq)
	if err != nil {
		processTaskErr := processTaskHandler.PublishTaskErrorResponse(executeTaskRequest, unableToProcessWorkdayFeed, err, ExecuteTaskResponseTopic, *count)
		if processTaskErr != nil {
			return processTaskErr
		}
		return err
	}

	err = processTaskHandler.PublishTaskResponse(executeTaskRequest, enums.TaskCompleted, ExecuteTaskResponseTopic, *count)
	if err != nil {
		return err
	}
	return nil
}

func processWorkdayFeed(processTaskHandler ProcessTaskHandlerInterface,
	workdayRepository repository.WorkdayRepositoryInterface,
	executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) (*int, error) {
	totalCount := 0
	if executeTaskRequest.ExecutionParametersDb.WorkdayFeed == enums.NoneWorkDayFeed {
		log.Info(context.Background(), "Processing Workday Feed for none")
		return &totalCount, nil
	}

	if executeTaskRequest.ExecutionParametersDb.WorkdayFeed == enums.ProcessAll {
		log.Info(context.Background(), "Processing Workday Feed for all")
		err := workdayRepository.UpdateIsSentToWorkdayFalse(executeTaskRequest)
		if err != nil {
			return &totalCount, err
		}
	}

	if executeTaskRequest.ExecutionParametersDb.WorkdayFeed == enums.ProcessNotSent {
		log.Info(context.Background(), "Processing Workday Feed for not sent")
	}

	for {
		workdayFeedRows, err := workdayRepository.GetWorkdayFeedRows(executeTaskRequest)
		if err != nil {
			return &totalCount, err
		}
		if len(workdayFeedRows) == 0 {
			return &totalCount, nil
		}
		err = processTaskHandler.CallWorkdayAPI(workdayFeedRows)
		if err != nil {
			return &totalCount, err
		}
		workdayRepository.UpdatePaymentWorkdayFeedStatus(workdayFeedRows)
		totalCount = totalCount + len(workdayFeedRows)
	}
}

func buildWorkdayFeedAccumilatedRequest(workdayFeed []*dbmodels.WorkdayFeed) []byte {

	var formattedWorkdayFeed []appmodels.WorkdayFeed
	for _, workdayFeedItem := range workdayFeed {

		metadata := appmodels.WorkdayFeedMetaData{}
		if err := json.Unmarshal([]byte(workdayFeedItem.Metadata), &metadata); err != nil {
			log.Error(context.Background(), err, "unable to unmrshal workdayFeedMetaData")
		}

		formattedWorkdayFeedItem := appmodels.WorkdayFeed{
			VendorType:       metadata.VendorType,
			Amount:           workdayFeedItem.Amount,
			PaymentDate:      workdayFeedItem.PaymentDate.Format("2006-01-02"),
			AtlasCheckNumber: metadata.CheckNumber,
			ClaimNumber:      metadata.ClaimNumber,
			PublicId:         metadata.PublicId,
		}

		formattedWorkdayFeed = append(formattedWorkdayFeed, formattedWorkdayFeedItem)
	}

	formattedWorkdayFeedJson, err := json.Marshal(formattedWorkdayFeed)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal formattedWorkdayFeed")
	}
	return formattedWorkdayFeedJson
}

func getErrorResponse(errorType string, err error) commonAppModels.ErrorResponse {
	errorResponse := commonAppModels.ErrorResponse{
		Type:    errorType,
		Message: err.Error(),
	}
	return errorResponse
}

func generateBearerTokenForWorkday() (*string, error) {
	configHandler := commonFunctions.GetConfigHandler()
	requstUrl := configHandler.GetString("PaymentPlatform.BearerToken.Url", "")
	grantType := configHandler.GetString("PaymentPlatform.BearerToken.GrantType", "")
	scope := configHandler.GetString("PaymentPlatform.BearerToken.Scope.Workday", "")
	clientId := configHandler.GetString("PaymentPlatform.BearerToken.ClientId", "")
	clientSecret := configHandler.GetString("PaymentPlatform.BearerToken.ClientSecret", "")

	request := commonAppModels.BearerTokenRequest{
		Url:          requstUrl,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        scope,
		GrantType:    grantType,
	}

	apiResponse, err := api.RetryApiCall(api.GenerateBearerTokenApiCall, request, 5)
	if err != nil {
		log.Error(context.Background(), err, "call failed to generate bearer token for url %s", requstUrl)
		return nil, err
	}
	return &apiResponse.Access_Token, nil
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

func validateAndGetTaskRequest(executeTaskRequest commonMessagingModels.ExecuteTaskRequest, isSendWorkdayDataTask bool) (*commonMessagingModels.ExecuteTaskRequestDb, error) {

	executeTaskReq := commonMessagingModels.ExecuteTaskRequestDb{
		Version:         executeTaskRequest.Version,
		Component:       enums.GetComponentTypeEnum(executeTaskRequest.Component),
		TaskName:        enums.GetTaskTypeEnum(executeTaskRequest.TaskName),
		TaskDate:        time.Time(executeTaskRequest.TaskDate),
		TaskExecutionId: executeTaskRequest.TaskExecutionId,
	}

	executionParam, err := validateAndGetExecutionParameters(executeTaskRequest.ExecutionParameters, isSendWorkdayDataTask)
	if err != nil {
		return nil, err
	}

	executeTaskReq.ExecutionParametersDb = *executionParam
	return &executeTaskReq, nil
}

func validateAndGetExecutionParameters(executionParam commonMessagingModels.ExecutionParameters, isSendWorkdayDataTask bool) (*commonMessagingModels.ExecutionParametersDb, error) {

	executionParameters := commonMessagingModels.ExecutionParametersDb{
		PaymentMethodType:  enums.GetPaymentMethodTypeEnumFromString(executionParam.PaymentMethodType),
		PaymentRequestType: enums.GetPaymentRequestTypeEnum(executionParam.PaymentRequestType),
		PaymentFrequency:   enums.GetPaymentFrequecyEnum(executionParam.PaymentFrequency),
		WorkdayFeed:        enums.GetWorkDayFeedEnum(executionParam.WorkdayFeed),
	}

	if !isSendWorkdayDataTask {
		if executionParameters.PaymentMethodType == enums.NonePaymentMethodType {
			return nil, fmt.Errorf("invalid input payment method type '%s'", executionParam.PaymentMethodType)
		} else if executionParameters.PaymentRequestType == enums.NonePaymentRequestType {
			return nil, fmt.Errorf("invalid input payment request type '%s'", executionParam.PaymentRequestType)
		} else if executionParameters.PaymentFrequency == enums.NonePaymentFrequency {
			return nil, fmt.Errorf("invalid input payment frequency '%s'", executionParam.PaymentFrequency)
		}
	} else {
		if executionParameters.PaymentRequestType == enums.NonePaymentRequestType {
			return nil, fmt.Errorf("invalid input payment request type '%s'", executionParam.PaymentRequestType)
		} else if executionParameters.WorkdayFeed == enums.NoneWorkDayFeed {
			return nil, fmt.Errorf("invalid input workday feed type '%s'", executionParam.WorkdayFeed)
		}
	}

	return &executionParameters, nil
}
