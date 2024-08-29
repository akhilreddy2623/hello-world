package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/task-manager-common/repository"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	appmodels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/app"
	dbmodels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/db"
)

var log = logging.GetLogger("task-manager-api-handlers")

// TODO: Should these be global constants?
const (
	invalidJsonMessage            = "invalidJson"
	invalidFieldMessage           = "invalidField"
	invalidJsonResponseMessage    = "invalid response, unable to encode the json payload"
	taskDependenciesNotMetMessage = "taskDependenciesNotMet"
	persistanceErrorMessage       = "unableToStoreRequest"
	messagingErrorMessage         = "unableToPublishMessage"
)

type Taskhandler struct {
}

func (Taskhandler) ExecuteTask(w http.ResponseWriter, r *http.Request) {
	log.Info(context.Background(), "Received request for ExecuteTask ", r)
	// TODO: 1. Validate Request JSON and input parameters : Done
	// TODO: 2. Make sure Task Dependencies are met : Done
	// TODO: 3	Insert Values in database : Done
	// TODO: 4. Write to Kafka Topic : Done
	var executeTaskRequest appmodels.ExecuteTaskRequest

	err := json.NewDecoder(r.Body).Decode(&executeTaskRequest)
	log.Info(context.Background(), "Decoding Done")
	if err != nil {
		log.Info(context.Background(), err, invalidJsonMessage) //TODO: update after invalid input logging decision
		executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, 0, enums.Errored.String(),
			createErrorResponse(invalidJsonMessage, err.Error(), http.StatusBadRequest))
		writeResponse(w, executeTaskResponse, executeTaskResponse.ErrorDetails.StatusCode)
		return
	}

	validationError := validateExecuteTaskRequestParemeters(executeTaskRequest)
	log.Info(context.Background(), "Validation Done")
	if validationError != nil {
		log.Info(context.Background(), validationError, invalidFieldMessage) //TODO: update after invalid input logging decision
		executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, 0, enums.Errored.String(), validationError)
		writeResponse(w, executeTaskResponse, executeTaskResponse.ErrorDetails.StatusCode)
		return
	}

	scheduledTaskDBRecord, dependencyError := areTaskDependenciesMet(executeTaskRequest)
	log.Info(context.Background(), "Dependency Check Done ")
	if dependencyError != nil {
		log.Info(context.Background(), dependencyError, taskDependenciesNotMetMessage) //TODO: update after invalid input logging decision
		executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, 0, enums.Errored.String(), dependencyError)
		writeResponse(w, executeTaskResponse, executeTaskResponse.ErrorDetails.StatusCode)
		return
	}

	taskExecutionDBRecord, persistanceError := persistTaskExecutionRequestInDB(executeTaskRequest)
	log.Info(context.Background(), "Persistance Done ")
	if persistanceError != nil {
		log.Error(context.Background(), errors.New(persistanceError.Message), persistanceErrorMessage)
		executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, 0, enums.Errored.String(), persistanceError)
		writeResponse(w, executeTaskResponse, executeTaskResponse.ErrorDetails.StatusCode)
		return
	}

	messagingError := publishMessageToTopic(*taskExecutionDBRecord, *scheduledTaskDBRecord)
	log.Info(context.Background(), "Publish To Kafka Done ")
	if messagingError != nil {
		log.Error(context.Background(), errors.New(messagingError.Message), messagingErrorMessage)
		executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, 0, enums.Errored.String(), messagingError)
		writeResponse(w, executeTaskResponse, executeTaskResponse.ErrorDetails.StatusCode)
		return
	}
	executeTaskResponse := getExecuteTaskResponse(executeTaskRequest.TaskId, taskExecutionDBRecord.TaskExecutionId, enums.Accepted.String(), nil)
	writeResponse(w, executeTaskResponse, http.StatusOK)
	log.Info(context.Background(), "Completed request for ExecuteTask")
}

func (Taskhandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Info(context.Background(), "Received request for GetTaskStatus ", r)

	// TODO: Implement GetTaskStatus
	// Not part of Day 1 Scope but good to have

	writeResponse(w, "", http.StatusOK)
	log.Info(context.Background(), "Completed request for GetTaskStatus ", "")
}

func validateExecuteTaskRequestParemeters(executeTaskRequest appmodels.ExecuteTaskRequest) *commonAppModels.ErrorResponse {

	if executeTaskRequest.TaskId < 1 {
		return createErrorResponse(invalidFieldMessage, "taskId should be a valid positive integer", http.StatusBadRequest)
	} else if !isValidDate(executeTaskRequest.Year, executeTaskRequest.Month, executeTaskRequest.Day) {
		return createErrorResponse(invalidFieldMessage, "Year, Month and Date should be a valid date", http.StatusBadRequest)
	} else if executeTaskRequest.ExecutionParameters.PaymentMethodType == enums.NonePaymentMethodType {
		return createErrorResponse(invalidFieldMessage, "PaymentMethodType should be ach, card or all", http.StatusBadRequest)
	} else if executeTaskRequest.ExecutionParameters.PaymentRequestType == enums.NonePaymentRequestType {
		return createErrorResponse(invalidFieldMessage, "PaymentRequestType should be customerchoice, insuranceautoauctions, sweep, incentive, commission or all", http.StatusBadRequest)
	} else if executeTaskRequest.ExecutionParameters.PaymentFrequency == enums.NonePaymentFrequency {
		return createErrorResponse(invalidFieldMessage, "PaymentFrequency should be onetime, recurring or all", http.StatusBadRequest)
	}
	return nil
}

func areTaskDependenciesMet(executeTaskRequest appmodels.ExecuteTaskRequest) (*dbmodels.ScheduledTask, *commonAppModels.ErrorResponse) {
	var scheduledTaskRepository repository.ScheduledTaskRepositoryInterface = repository.ScheduledTaskRepository{}
	executionDate, dateParseError := convertStringToDate(executeTaskRequest.Year, executeTaskRequest.Month, executeTaskRequest.Day)

	if dateParseError != nil {
		log.Info(context.Background(), dateParseError) //TODO: update after invalid input logging descision
		return nil, createErrorResponse(invalidFieldMessage, dateParseError.Error(), http.StatusBadRequest)
	}

	scheduledTaskDBRecord, err := scheduledTaskRepository.AreTaskDependenciesMet(executeTaskRequest.TaskId, executionDate)

	if err != nil {
		return nil, createErrorResponse(taskDependenciesNotMetMessage, err.Error(), http.StatusFailedDependency)
	}
	return scheduledTaskDBRecord, nil
}

func persistTaskExecutionRequestInDB(executeTaskRequest appmodels.ExecuteTaskRequest) (*dbmodels.TaskExecution, *commonAppModels.ErrorResponse) {
	var scheduledTaskRepository repository.ScheduledTaskRepositoryInterface = repository.ScheduledTaskRepository{}
	taskExecution, conversionError := getTaskExecutionDBModel(executeTaskRequest)

	if conversionError != nil {
		return nil, createErrorResponse(taskDependenciesNotMetMessage, conversionError.Error(), http.StatusFailedDependency)
	}

	persistanceError := scheduledTaskRepository.PersistTaskExecutionRequestInDB(taskExecution)
	if persistanceError != nil {
		return nil, createErrorResponse(taskDependenciesNotMetMessage, persistanceError.Error(), http.StatusFailedDependency)
	}

	return taskExecution, nil
}

func publishMessageToTopic(taskExecutionDBRecord dbmodels.TaskExecution, scheduledTaskDBRecord dbmodels.ScheduledTask) *commonAppModels.ErrorResponse {
	configHandler := commonFunctions.GetConfigHandler()
	executeTaskRequestsTopicName := configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")

	executeTaskRequest := commonMessagingModels.ExecuteTaskRequest{
		Version:         1,
		Component:       scheduledTaskDBRecord.Component,
		TaskName:        scheduledTaskDBRecord.TaskName,
		TaskDate:        commonMessagingModels.JsonDate(taskExecutionDBRecord.ExecutionDate), //TODO: Check why are we using a new Type here
		TaskExecutionId: taskExecutionDBRecord.TaskExecutionId,
		ExecutionParameters: commonMessagingModels.ExecutionParameters{
			PaymentMethodType:  taskExecutionDBRecord.ExecutionParameters.PaymentMethodType.String(),
			PaymentRequestType: taskExecutionDBRecord.ExecutionParameters.PaymentRequestType.String(),
			PaymentFrequency:   taskExecutionDBRecord.ExecutionParameters.PaymentFrequency.String(),
		},
	}

	executeTaskRequestJson, marshalError := json.Marshal(executeTaskRequest)
	if marshalError != nil {
		log.Error(context.Background(), marshalError, "error in marshalling schedulePaymentRequest")
		return createErrorResponse(taskDependenciesNotMetMessage, marshalError.Error(), http.StatusInternalServerError)
	}
	publishError := messaging.KafkaPublish(executeTaskRequestsTopicName, string(executeTaskRequestJson))
	if publishError != nil {
		log.Error(context.Background(), publishError, "error in marshalling schedulePaymentRequest")
		return createErrorResponse(taskDependenciesNotMetMessage, publishError.Error(), http.StatusInternalServerError)
	}
	return nil
}

func getTaskExecutionDBModel(executeTaskRequest appmodels.ExecuteTaskRequest) (*dbmodels.TaskExecution, error) {
	executionDate, dateParseError := convertStringToDate(executeTaskRequest.Year, executeTaskRequest.Month, executeTaskRequest.Day)

	if dateParseError != nil {
		return nil, dateParseError
	}
	return &dbmodels.TaskExecution{
		TaskId:              executeTaskRequest.TaskId,
		ExecutionDate:       executionDate,
		ExecutionParameters: executeTaskRequest.ExecutionParameters,
		StartTime:           time.Now(),
		Status:              enums.TaskStaged,
	}, nil
}

func getExecuteTaskResponse(taskId int, taskExecutionId int, status string, errorDetails *commonAppModels.ErrorResponse) appmodels.ExecuteTaskResponse {
	return appmodels.ExecuteTaskResponse{
		TaskId:          taskId,
		TaskExecutionId: taskExecutionId,
		Status:          status,
		ErrorDetails:    errorDetails,
	}
}

// TODO Should write resaponse be a solution function as well?
func writeResponse(w http.ResponseWriter, r any, statusCode int) {
	response, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Error(context.Background(), err, invalidJsonResponseMessage)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
func createErrorResponse(errorType string, errorMessage string, statusCode int) *commonAppModels.ErrorResponse {
	return &commonAppModels.ErrorResponse{
		Type:       errorType,
		Message:    errorMessage,
		StatusCode: statusCode,
	}
}

// TODO: discuss and move below 2 functions to a util class
func isValidDate(year string, month string, day string) bool {
	_, err := time.Parse("2006-01-02", year+"-"+month+"-"+day)
	if err != nil {
		log.Info(context.Background(), err) //TODO: update after invalid input logging descision
	}

	return err == nil
}

func convertStringToDate(year string, month string, day string) (time.Time, error) {
	dateString := year + "-" + month + "-" + day //TODO: Is this the best way to do this?
	date, dateParseError := time.Parse("2006-01-02", dateString)
	if dateParseError != nil {
		log.Info(context.Background(), dateParseError) //TODO: discuss if private methods should be logging the errors
	}
	return date, dateParseError
}
