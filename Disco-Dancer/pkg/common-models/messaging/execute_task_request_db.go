package messaging

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type ExecuteTaskRequestDb struct {
	Version               int
	Component             enums.ComponentType
	TaskName              enums.TaskType
	TaskDate              time.Time
	TaskExecutionId       int
	ExecutionParametersDb ExecutionParametersDb
}

type ExecutionParametersDb struct {
	PaymentMethodType  enums.PaymentMethodType
	PaymentRequestType enums.PaymentRequestType
	PaymentFrequency   enums.PaymentFrequency
	WorkdayFeed        enums.WorkdayFeed
}
