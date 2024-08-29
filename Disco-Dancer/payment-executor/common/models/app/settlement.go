package app

import "time"

type SettlementPayment struct {
	AccountIdentifier    string
	RoutingNumber        string
	AccountName          string
	Amount               float32
	ConsolidatedId       int64
	PaymentRequestType   int16
	PaymentDate          time.Time
	SettlementIdentifier string
}
