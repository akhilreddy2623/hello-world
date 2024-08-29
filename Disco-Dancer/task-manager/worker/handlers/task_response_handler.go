package handlers

import (
	"context"
	"encoding/json"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/task-manager-common/repository"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var log = logging.GetLogger("Task-manager-worker")

var ExecuteTaskResponseHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in dv.paymentplatform.internal.executetaskresponses topic: '%s'", *message.Body)

	var scheduledTaskRepository repository.ScheduledTaskRepositoryInterface = repository.ScheduledTaskRepository{}

	executeTaskResponse := commonMessagingModels.ExecuteTaskResponse{}
	if err := json.Unmarshal([]byte(*message.Body), &executeTaskResponse); err != nil {
		log.Error(context.Background(), err, "unable to unmarshal executetaskresponses")
		return err
	}

	return ProcessExecuteTaskResponse(ctx, executeTaskResponse, scheduledTaskRepository)
}

func ProcessExecuteTaskResponse(
	ctx context.Context,
	executeTaskResponse commonMessagingModels.ExecuteTaskResponse,
	scheduledTaskRepository repository.ScheduledTaskRepositoryInterface) error {

	if err := scheduledTaskRepository.UpdateTaskStatus(executeTaskResponse); err != nil {
		log.Error(context.Background(), err, "unable to update Task status")
		return err
	}
	return nil
}
