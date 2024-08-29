package db

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type SettlementPayment struct {
	AccountIdentifier    string
	RoutingNumber        string
	PaymentExtendedData  PaymentExtendedData
	Amount               float32
	ConsolidatedId       int64
	PaymentRequestType   int16
	PaymentDate          time.Time
	SettlementIdentifier string
}

type PaymentExtendedData struct {
	FirstName           string
	LastName            string
	ACHAccountType      enums.ACHAccountType
	ZipCode             string
	CVV                 string
	CardExpirationMonth string
	CardExpirationYear  string
}
