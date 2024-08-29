package enums

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

type PaymentMethodType int

const (
	NonePaymentMethodType PaymentMethodType = iota
	ACH
	Card
	AllPaymentMethodType
)

func (p PaymentMethodType) String() string {
	return [...]string{
		"",
		"ach",
		"card",
		"all"}[p]
}

func (p PaymentMethodType) EnumIndex() int {
	return int(p)
}

func GetPaymentMethodTypeEnum(paymentMethodType int) PaymentMethodType {
	switch paymentMethodType {
	case 1:
		return ACH
	case 2:
		return Card
	case 3:
		return AllPaymentMethodType
	}
	return NonePaymentMethodType
}

func GetPaymentMethodTypeEnumFromString(paymentMethodType string) PaymentMethodType {
	switch strings.ToLower(paymentMethodType) {
	case "ach":
		return ACH
	case "card":
		return Card
	case "all":
		return AllPaymentMethodType
	}
	return NonePaymentMethodType
}

var toID = map[string]PaymentMethodType{
	"ach":  ACH,
	"card": Card,
	"all":  AllPaymentMethodType,
}

var toString = map[PaymentMethodType]string{
	ACH:                  "ach",
	Card:                 "card",
	AllPaymentMethodType: "all",
}

func (p PaymentMethodType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PaymentMethodType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toID[strings.ToLower(j)]
	return nil
}

// Helps showing available enum values for PaymentRequestType on Error Message
func (p *PaymentMethodType) GetAllowedEnumValues() string {
	values := make([]string, 0, len(toString))
	for _, value := range toString {
		values = append(values, value)
	}

	sort.Strings(values) // sort values in increasing order

	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", and " + values[len(values)-1]
	}
}
