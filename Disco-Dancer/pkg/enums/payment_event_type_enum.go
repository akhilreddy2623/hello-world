package enums

import (
	"bytes"
	"encoding/json"
	"strings"
)

type PaymentEventType int

const (
	NoEvent PaymentEventType = iota
	SavedInAdminstrator
	SentByAdminstrator
	ReceivedByExecutor
	SentToBank
	PaymentSettled
	Cancelled
	Reversed
)

func (p PaymentEventType) String() string {
	return [...]string{
		"noevent",
		"savedinadminstrator",
		"sentbyadminstrator",
		"receivedbyexecutor",
		"senttobank",
		"paymentsettled",
		"cancelled",
		"reversed"}[p]
}

func (p PaymentEventType) EnumIndex() int {
	return int(p)
}

var toPaymentEvent = map[string]PaymentEventType{
	"noevent":             NoEvent,
	"savedinadminstrator": SavedInAdminstrator,
	"sentbyadminstrator":  SentByAdminstrator,
	"receivedbyexecutor":  ReceivedByExecutor,
	"senttobank":          SentToBank,
	"paymentsettled":      PaymentSettled,
	"cancelled":           Cancelled,
	"reversed":            Reversed,
}

func (p PaymentEventType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(strings.ToUpper(p.String()))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PaymentEventType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toPaymentEvent[strings.ToLower(j)]
	return nil
}
