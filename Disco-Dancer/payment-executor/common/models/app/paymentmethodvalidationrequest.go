package app

import "geico.visualstudio.com/Billing/plutus/enums"

type PaymentMethodValidationRequest struct {
	UserId        string
	Amount        float32
	FirstName     string
	LastName      string
	AccountNumber string
	RoutingNumber string
	AccountType   enums.ACHAccountType
}
