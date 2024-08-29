package app

import (
	"geico.visualstudio.com/Billing/plutus/enums"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
)

type ExecuteTaskRequest struct {
	TaskId              int
	Year                string
	Month               string
	Day                 string
	ExecutionParameters TaskExecutionParemeters
}

type ExecuteTaskResponse struct {
	TaskId          int                            `json:"taskId"`
	TaskExecutionId int                            `json:"taskExecutionId,omitempty"`
	Status          string                         `json:"status"`
	ErrorDetails    *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}

type TaskExecutionParemeters struct {
	PaymentMethodType  enums.PaymentMethodType
	PaymentRequestType enums.PaymentRequestType
	PaymentFrequency   enums.PaymentFrequency
}
