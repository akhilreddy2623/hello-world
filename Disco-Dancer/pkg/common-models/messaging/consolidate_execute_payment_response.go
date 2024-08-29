package messaging

import "time"

type ConsolidatedExecutePaymentResponse struct {
	Version                int
	ConsolidatedId         int64
	Status                 string
	SettlementIdentifier   string
	PaymentDate            time.Time
	Amount                 float32
	Last4AccountIdentifier string
	PaymentRequestType     string
}
