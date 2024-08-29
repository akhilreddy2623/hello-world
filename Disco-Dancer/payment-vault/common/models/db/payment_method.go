package db

import "geico.visualstudio.com/Billing/plutus/enums"

type PaymentMethod struct {
	PaymentMethodId            int64
	UserID                     string
	CallerApp                  enums.CallerApp
	PaymentMethodType          enums.PaymentMethodType
	AccountIdentifier          string
	Last4AccountIdentifier     string
	EncryptedAccountIdentifier string
	RoutingNumber              string
	NickName                   string
	PaymentExtendedData        PaymentExtendedData
}

type PaymentExtendedData struct {
	FirstName           string               `json:"FirstName,omitempty"`
	LastName            string               `json:"LastName,omitempty"`
	ACHAccountType      enums.ACHAccountType `json:"ACHAccountType,omitempty"`
	ZipCode             string               `json:"ZipCode,omitempty"`
	CVV                 string               `json:"CVV,omitempty"`
	CardExpirationMonth string               `json:"CardExpirationMonth,omitempty"`
	CardExpirationYear  string               `json:"CardExpirationYear,omitempty"`
}
