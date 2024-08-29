package handlers

import (
	"context"
	"testing"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	repositoryMock "geico.visualstudio.com/Billing/plutus/task-manager-common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_ExecuteTaskResponseHandlerSuccess(t *testing.T) {
	var taskExecutionId int = 2

	taskResponse := commonMessagingModels.ExecuteTaskResponse{
		Version:               1,
		TaskExecutionId:       taskExecutionId,
		Status:                "completed",
		ProcessedRecordsCount: 1,
	}

	ScheduledTaskRepositoryInterface := repositoryMock.ScheduledTaskRepositoryInterface{}
	ScheduledTaskRepositoryInterface.On("UpdateTaskStatus", taskResponse).Return(nil)
	err := ProcessExecuteTaskResponse(context.Background(), taskResponse, &ScheduledTaskRepositoryInterface)
	assert.Nil(t, err)
}
