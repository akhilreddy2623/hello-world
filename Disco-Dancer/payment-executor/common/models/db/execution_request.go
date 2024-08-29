package db

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type ExecutionRequest struct {
	ExecutionRequestId     int64
	TenantId               int64
	PaymentId              int64
	ConsolidatedId         int64
	AccountIdentifier      string
	Last4AccountIdentifier string
	RoutingNumber          string
	Amount                 float32
	PaymentDate            time.Time
	PaymentFrequency       enums.PaymentFrequency
	TransactionType        enums.TransactionType
	PaymentRequestType     enums.PaymentRequestType
	PaymentMethodType      enums.PaymentMethodType
	PaymentExtendedData    string
	Status                 enums.PaymentStatus
	SettlementIdentifier   string
}

type ListExecutionRequest struct {
	ExecutionRequestCollection []ExecutionRequest
	ConsolidatedAmount         float32
}
