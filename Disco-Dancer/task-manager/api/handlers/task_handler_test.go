package handlers

import (
	"net/http"
	"testing"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	appmodels "geico.visualstudio.com/Billing/plutus/task-manager-common/models/app"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name  string
	input appmodels.ExecuteTaskRequest
	want  *commonAppModels.ErrorResponse
}

func Test_ValidateRequestParameters(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Test for Invalid TaskId",
			input: appmodels.ExecuteTaskRequest{
				TaskId: -1},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "taskId should be a valid positive integer",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid Year",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "10"},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "Year, Month and Date should be a valid date",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid Month",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "2030",
				Month:  "15",
			},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "Year, Month and Date should be a valid date",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid Day",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "2030",
				Month:  "10",
				Day:    "40",
			},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "Year, Month and Date should be a valid date",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid PaymentMethodType",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "2030",
				Month:  "10",
				Day:    "15",
				ExecutionParameters: appmodels.TaskExecutionParemeters{
					PaymentMethodType: enums.NonePaymentMethodType,
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "PaymentMethodType should be ach, card or all",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid PaymentRequestType",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "2030",
				Month:  "10",
				Day:    "15",
				ExecutionParameters: appmodels.TaskExecutionParemeters{
					PaymentMethodType:  enums.AllPaymentMethodType,
					PaymentRequestType: enums.NonePaymentRequestType,
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "PaymentRequestType should be customerchoice, insuranceautoauctions, sweep, incentive, commission or all",
				StatusCode: http.StatusBadRequest},
		},
		{
			name: "Test for Invalid PaymentFrequency",
			input: appmodels.ExecuteTaskRequest{
				TaskId: 1,
				Year:   "2030",
				Month:  "10",
				Day:    "15",
				ExecutionParameters: appmodels.TaskExecutionParemeters{
					PaymentMethodType:  enums.AllPaymentMethodType,
					PaymentRequestType: enums.AllPaymentRequestType,
					PaymentFrequency:   enums.NonePaymentFrequency,
				},
			},
			want: &commonAppModels.ErrorResponse{
				Type:       invalidFieldMessage,
				Message:    "PaymentFrequency should be onetime, recurring or all",
				StatusCode: http.StatusBadRequest},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := validateExecuteTaskRequestParemeters(testCase.input)
			assert.Equal(t, testCase.want.Message, got.Message)
			assert.Equal(t, testCase.want.Type, got.Type)
			assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
		})
	}

}
