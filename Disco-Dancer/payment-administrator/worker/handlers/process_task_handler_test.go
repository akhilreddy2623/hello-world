package handlers

import (
	"fmt"
	"testing"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/enums"
	repositoryMock "geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository/mocks"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-worker/handlers/mocks"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name  string
	input commonMessagingModels.ExecuteTaskRequest
	want  error
}

func Test_ValidateProcessPaymentHandlerInputs(t *testing.T) {
	date, _ := time.Parse(commonMessagingModels.TimeFormat, "2024-03-27")
	paymentRepositoryInterface := repositoryMock.PaymentRepositoryInterface{}
	processTaskHandlerInterface := mocks.ProcessTaskHandlerInterface{}

	testCases := []TestCase{
		{
			name: "PaymentMethodTypeShouldBeValid",
			input: commonMessagingModels.ExecuteTaskRequest{
				Version:         1,
				Component:       "administrator",
				TaskName:        "processpayments",
				TaskDate:        commonMessagingModels.JsonDate(date),
				TaskExecutionId: 1,
				ExecutionParameters: commonMessagingModels.ExecutionParameters{
					PaymentMethodType:  "none",
					PaymentRequestType: "all",
					PaymentFrequency:   "all",
				},
			},
			want: fmt.Errorf("invalid input payment method type '%s'", "none"),
		},
		{
			name: "PaymentRequestTypeShouldBeValid",
			input: commonMessagingModels.ExecuteTaskRequest{
				Version:         1,
				Component:       "administrator",
				TaskName:        "processpayments",
				TaskDate:        commonMessagingModels.JsonDate(date),
				TaskExecutionId: 1,
				ExecutionParameters: commonMessagingModels.ExecutionParameters{
					PaymentMethodType:  "all",
					PaymentRequestType: "none",
					PaymentFrequency:   "all",
				},
			},
			want: fmt.Errorf("invalid input payment request type '%s'", "none"),
		},
		{
			name: "PaymentFrequencyShouldBeValid",
			input: commonMessagingModels.ExecuteTaskRequest{
				Version:         1,
				Component:       "administrator",
				TaskName:        "processpayments",
				TaskDate:        commonMessagingModels.JsonDate(date),
				TaskExecutionId: 1,
				ExecutionParameters: commonMessagingModels.ExecutionParameters{
					PaymentMethodType:  "all",
					PaymentRequestType: "all",
					PaymentFrequency:   "none",
				},
			},
			want: fmt.Errorf("invalid input payment frequency '%s'", "none"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			processTaskHandlerInterface.On("PublishTaskErrorResponse", testCase.input, "invalidField", testCase.want, "", 0).Return(nil)
			got := processPaymentTask(testCase.input, &paymentRepositoryInterface, &processTaskHandlerInterface)

			assert.Equal(t, testCase.want, got)
		})
	}

}

func Test_ProcessPaymentHandlerSuccess(t *testing.T) {
	date, _ := time.Parse(commonMessagingModels.TimeFormat, "2024-03-27")
	executeTaskRequest := commonMessagingModels.ExecuteTaskRequest{
		Version:         1,
		Component:       "administrator",
		TaskName:        "processpayments",
		TaskDate:        commonMessagingModels.JsonDate(date),
		TaskExecutionId: 1,
		ExecutionParameters: commonMessagingModels.ExecutionParameters{
			PaymentMethodType:  "all",
			PaymentRequestType: "all",
			PaymentFrequency:   "all",
		},
	}
	dbexecuteTaskRequest := commonMessagingModels.ExecuteTaskRequestDb{
		Version:         1,
		Component:       1,
		TaskName:        1,
		TaskDate:        date,
		TaskExecutionId: 1,
		ExecutionParametersDb: commonMessagingModels.ExecutionParametersDb{
			PaymentMethodType:  3,
			PaymentRequestType: 6,
			PaymentFrequency:   3,
		},
	}
	paymentRepositoryInterface := repositoryMock.PaymentRepositoryInterface{}
	processTaskHandlerInterface := mocks.ProcessTaskHandlerInterface{}
	var count int = 5
	var countPtr *int = &count
	paymentRepositoryInterface.On("ProcessPayments", dbexecuteTaskRequest).Return(countPtr, nil)
	processTaskHandlerInterface.On("PublishTaskResponse", executeTaskRequest, enums.TaskInprogress, "", 0).Return(nil)
	processTaskHandlerInterface.On("PublishTaskResponse", executeTaskRequest, enums.TaskCompleted, "", *countPtr).Return(nil)
	err := processPaymentTask(executeTaskRequest, &paymentRepositoryInterface, &processTaskHandlerInterface)
	assert.Nil(t, err)
}

func Test_ValidateProcessWorkdayHandlerInputs(t *testing.T) {
	date, _ := time.Parse(commonMessagingModels.TimeFormat, "2024-03-27")
	workdayRepository := repositoryMock.WorkdayRepositoryInterface{}
	processTaskHandlerInterface := mocks.ProcessTaskHandlerInterface{}

	testCases := []TestCase{
		{
			name: "PaymentRequestTypeShouldBeValid",
			input: commonMessagingModels.ExecuteTaskRequest{
				Version:         1,
				Component:       "administrator",
				TaskName:        "sendworkdaydata",
				TaskDate:        commonMessagingModels.JsonDate(date),
				TaskExecutionId: 1,
				ExecutionParameters: commonMessagingModels.ExecutionParameters{
					PaymentRequestType: "none",
					WorkdayFeed:        "processall",
				},
			},
			want: fmt.Errorf("invalid input payment request type '%s'", "none"),
		},
		{
			name: "WorkdayFeedTypeShouldBeValid",
			input: commonMessagingModels.ExecuteTaskRequest{
				Version:         1,
				Component:       "administrator",
				TaskName:        "sendworkdaydata",
				TaskDate:        commonMessagingModels.JsonDate(date),
				TaskExecutionId: 1,
				ExecutionParameters: commonMessagingModels.ExecutionParameters{
					PaymentRequestType: "all",
					WorkdayFeed:        "none",
				},
			},
			want: fmt.Errorf("invalid input workday feed type '%s'", "none"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			processTaskHandlerInterface.On("PublishTaskErrorResponse", testCase.input, "invalidField", testCase.want, "", 0).Return(nil)
			got := sendWorkdayDataTask(testCase.input, &workdayRepository, &processTaskHandlerInterface)

			assert.Equal(t, testCase.want, got)
		})
	}
}

func Test_ProcessWorkdayHandlerSuccess(t *testing.T) {
	date, _ := time.Parse(commonMessagingModels.TimeFormat, "2024-03-27")
	executeTaskRequest := commonMessagingModels.ExecuteTaskRequest{
		Version:         1,
		Component:       "administrator",
		TaskName:        "sendworkdaydata",
		TaskDate:        commonMessagingModels.JsonDate(date),
		TaskExecutionId: 1,
		ExecutionParameters: commonMessagingModels.ExecutionParameters{
			PaymentRequestType: "all",
			WorkdayFeed:        "processall",
		},
	}
	dbexecuteTaskRequest := commonMessagingModels.ExecuteTaskRequestDb{
		Version:         1,
		Component:       1,
		TaskName:        4,
		TaskDate:        date,
		TaskExecutionId: 1,
		ExecutionParametersDb: commonMessagingModels.ExecutionParametersDb{
			PaymentRequestType: 6,
			WorkdayFeed:        1,
		},
	}
	workdayRepositoryInterface := repositoryMock.WorkdayRepositoryInterface{}
	processTaskHandlerInterface := mocks.ProcessTaskHandlerInterface{}
	var count int = 0
	var countPtr *int = &count
	workdayRepositoryInterface.On("UpdateIsSentToWorkdayFalse", dbexecuteTaskRequest).Return(nil)
	workdayRepositoryInterface.On("GetWorkdayFeedRows", dbexecuteTaskRequest).Return(nil, nil)
	workdayRepositoryInterface.On("UpdatePaymentWorkdayFeedStatus", nil)
	processTaskHandlerInterface.On("PublishTaskResponse", executeTaskRequest, enums.TaskInprogress, "", 0).Return(nil)
	processTaskHandlerInterface.On("PublishTaskResponse", executeTaskRequest, enums.TaskCompleted, "", *countPtr).Return(nil)
	processTaskHandlerInterface.On("CallWorkdayAPI", nil).Return(nil)
	err := sendWorkdayDataTask(executeTaskRequest, &workdayRepositoryInterface, &processTaskHandlerInterface)
	assert.Nil(t, err)
}
