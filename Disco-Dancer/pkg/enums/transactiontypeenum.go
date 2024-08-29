package enums

import "strings"

type TransactionType int

const (
	NoneTransactionType TransactionType = iota
	PayIn
	PayOut
)

func (t TransactionType) String() string {
	return [...]string{"", "payin", "payout"}[t]
}

func (t TransactionType) EnumIndex() int {
	return int(t)
}

func GetTransactionTypeEnum(transactionType string) TransactionType {
	switch strings.ToLower(transactionType) {
	case "payin":
		return PayIn
	case "payout":
		return PayOut
	}
	return NoneTransactionType
}
