package db

import "geico.visualstudio.com/Billing/plutus/enums"

type PaymentMethodValidation struct {
	BankAccountNumber          string
	EncryptedBankAccountNumber string
	RoutingNumber              string
	FirstName                  string
	LastName                   string
	AccountType                enums.ACHAccountType
	UserID                     string
	CallerApp                  enums.CallerApp
	Amount                     float32
}
