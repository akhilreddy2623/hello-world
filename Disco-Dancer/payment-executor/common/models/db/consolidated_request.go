package db

import (
	"encoding/json"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type ConsolidatedRequest struct {
	ConsolidatedId         int64
	AccountIdentifier      string
	RoutingNumber          string
	Last4AccountIdentifier string
	Amount                 float32
	PaymentDate            time.Time
	PaymentRequestType     enums.PaymentRequestType
	PaymentExtendedData    json.RawMessage
	Status                 enums.PaymentStatus
	SettlementIdentifier   string
}
