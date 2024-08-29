package enums

import "strings"

type PaymentMethodValidationStatus int

const (
	NonePaymentMethodValidationStatus PaymentMethodValidationStatus = iota
	ApprovedPaymentMethodValidationStatus
	DeclinedPaymentMethodValidationStatus
)

func (p PaymentMethodValidationStatus) String() string {
	return [...]string{"", "approved", "declined"}[p]
}

func (p PaymentMethodValidationStatus) EnumIndex() int {
	return int(p)
}

func GetPaymentMethodValidationStatusEnum(paymentMethodValidationStatus string) PaymentMethodValidationStatus {
	switch strings.ToLower(paymentMethodValidationStatus) {
	case "approved":
		return ApprovedPaymentMethodValidationStatus
	case "declined":
		return DeclinedPaymentMethodValidationStatus
	}
	return NonePaymentMethodValidationStatus
}

func GetPaymentMethodValidationStatusEnumByIndex(paymentMethodValidationStatus int) PaymentMethodValidationStatus {
	switch paymentMethodValidationStatus {
	case 1:
		return ApprovedPaymentMethodValidationStatus
	case 2:
		return DeclinedPaymentMethodValidationStatus
	}
	return NonePaymentMethodValidationStatus
}
