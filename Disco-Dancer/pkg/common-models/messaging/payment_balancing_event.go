package messaging

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
)

type PaymentEvent struct {
	Version            int
	PaymentId          int64
	Amount             float32
	PaymentDate        time.Time
	PaymentRequestType enums.PaymentRequestType
	PaymentMethodType  enums.PaymentMethodType
	EventType          enums.PaymentEventType
	EventDateTime      time.Time
	SettlementAmount   float32
}
