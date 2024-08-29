package enums

type PaymentMethodStatus int

const (
	NonePaymentMethodStatus PaymentMethodStatus = iota
	Active
	Restricted
	Fraud
	Invalid
)

func (p PaymentMethodStatus) String() string {
	return [...]string{
		"",
		"active",
		"restricted",
		"fraud",
		"invalid"}[p]
}

func (p PaymentMethodStatus) EnumIndex() int {
	return int(p)
}

func GetPaymentMethodStatusEnum(paymentMethodStatus int) PaymentMethodStatus {
	switch paymentMethodStatus {
	case 1:
		return Active
	case 2:
		return Restricted
	case 3:
		return Fraud
	case 4:
		return Invalid
	}
	return NonePaymentMethodStatus
}
