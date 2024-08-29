package enums

import (
	"bytes"
	"encoding/json"
	"strings"
)

type PaymentFrequency int

const (
	NonePaymentFrequency PaymentFrequency = iota
	OneTime
	Recurrring
	AllPaymentFrequency
)

func (p PaymentFrequency) String() string {
	return [...]string{"", "onetime", "recurrring", "all"}[p]
}

func (p PaymentFrequency) EnumIndex() int {
	return int(p)
}

func GetPaymentFrequecyEnum(paymentFrequency string) PaymentFrequency {
	switch strings.ToLower(paymentFrequency) {
	case "onetime":
		return OneTime
	case "recurring":
		return Recurrring
	case "all":
		return AllPaymentFrequency
	}
	return NonePaymentFrequency
}

var toPaymentFrequencyId = map[string]PaymentFrequency{
	"onetime":   OneTime,
	"recurring": Recurrring,
	"all":       AllPaymentFrequency,
}

var toPaymentFrequecyString = map[PaymentFrequency]string{
	OneTime:             "onetime",
	Recurrring:          "recurring",
	AllPaymentFrequency: "all",
}

func (p PaymentFrequency) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toPaymentFrequecyString[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PaymentFrequency) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toPaymentFrequencyId[strings.ToLower(j)]
	return nil
}
