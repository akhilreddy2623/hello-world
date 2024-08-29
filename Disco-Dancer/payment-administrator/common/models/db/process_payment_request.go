package db

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type ProcessPaymentRequest struct {
	TenantId               int64
	PaymentId              int64
	PaymentFrequency       enums.PaymentFrequency
	TransactionType        enums.TransactionType
	PaymentRequestType     enums.PaymentRequestType
	PaymentMethodType      enums.PaymentMethodType
	PaymentExtendedData    string
	AccountIdentifier      string
	Last4AccountIdentifier string
	RoutingNumber          string
	Amount                 float32
	PaymentDate            time.Time
	UserId                 string
	ProductIdentifier      string
	RequestId              int64
}
