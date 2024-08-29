package db

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type ExecuteTaskRequest struct {
	Version             int
	Component           enums.ComponentType
	TaskName            enums.TaskType
	TaskDate            time.Time
	TaskExecutionId     int
	ExecutionParameters ExecutionParameters
}

type ExecutionParameters struct {
	PaymentMethodType  enums.PaymentMethodType
	PaymentRequestType enums.PaymentRequestType
	PaymentFrequency   enums.PaymentFrequency
}
