package enums

import (
	"bytes"
	"encoding/json"
	"strings"
)

type PaymentStatus int

const (
	NonePaymentStatus PaymentStatus = iota
	New
	Accepted
	Errored
	InProgress
	Completed
	Consolidated
	Settled
)

func (p PaymentStatus) String() string {
	return [...]string{
		"",
		"new",
		"accepted",
		"errored",
		"inprogress",
		"completed",
		"consolidated",
		"settled"}[p]
}

func (p PaymentStatus) EnumIndex() int {
	return int(p)
}

var toPaymentStatus = map[string]PaymentStatus{
	"":             NonePaymentStatus,
	"new":          New,
	"accepted":     Accepted,
	"errored":      Errored,
	"inprogress":   InProgress,
	"completed":    Completed,
	"consolidated": Consolidated,
	"settled":      Settled,
}

func (p PaymentStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(strings.ToUpper(p.String()))
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PaymentStatus) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toPaymentStatus[strings.ToLower(j)]
	return nil
}
