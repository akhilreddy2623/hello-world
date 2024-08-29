package enums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type PaymentRequestType int

const (
	NonePaymentRequestType PaymentRequestType = iota
	CustomerChoice
	InsuranceAutoAuctions
	Sweep
	Incentive
	Commission
	AllPaymentRequestType
)

func (p PaymentRequestType) String() string {
	return [...]string{"", "customerchoice", "iaa", "sweep", "incentive", "commission", "all"}[p]
}

func (p PaymentRequestType) EnumIndex() int {
	return int(p)
}

func GetPaymentRequestTypeEnum(paymentRequestType string) PaymentRequestType {
	switch strings.ToLower(paymentRequestType) {
	case "customerchoice":
		return CustomerChoice
	case "iaa", "insuranceautoauctions":
		return InsuranceAutoAuctions
	case "sweep":
		return Sweep
	case "incentive":
		return Incentive
	case "commission":
		return Commission
	case "all":
		return AllPaymentRequestType
	}
	return NonePaymentRequestType
}

var toPaymentRequestTypeId = map[string]PaymentRequestType{
	"customerchoice": CustomerChoice,
	"iaa":            InsuranceAutoAuctions,
	"sweep":          Sweep,
	"incentive":      Incentive,
	"commission":     Commission,
	"all":            AllPaymentRequestType,
}

var toPaymentRequestTypeString = map[PaymentRequestType]string{
	CustomerChoice:        "customerchoice",
	InsuranceAutoAuctions: "iaa",
	Sweep:                 "sweep",
	Incentive:             "incentive",
	Commission:            "commission",
	AllPaymentRequestType: "all",
}

func (p PaymentRequestType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toPaymentRequestTypeString[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PaymentRequestType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toPaymentRequestTypeId[strings.ToLower(j)]
	return nil
}

// Helps showing available enum values for PaymentRequestType on Error Message
func (PaymentRequestType) GetAllowedEnumValues() string {
	values := make([]string, 0, len(toPaymentRequestTypeString))
	for _, value := range toPaymentRequestTypeString {
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

func (p PaymentRequestType) IsValidPaymentRequestType() error {
	if p == NonePaymentRequestType {
		return fmt.Errorf("payment request type should be %s", p.GetAllowedEnumValues())
	}
	return nil
}
