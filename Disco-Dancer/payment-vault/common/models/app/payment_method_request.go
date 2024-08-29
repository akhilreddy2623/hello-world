package app

import "geico.visualstudio.com/Billing/plutus/enums"

type PaymentMethodRequest struct {
	UserID               string                  `json:"UserID" example:"125"`
	PaymentMethodType    enums.PaymentMethodType `json:"PaymentMethodType" swaggertype:"string" example:"ach"`
	FirstName            string                  `json:"FirstName" example:"John"`
	LastName             string                  `json:"LastName" example:"Doe"`
	CallerApp            enums.CallerApp         `json:"CallerApp" swaggertype:"string" example:"mcp"`
	PaymentMethodDetails Paymentmethoddetails
}

type Paymentmethoddetails struct {
	AccountNumber       string               `json:"AccountNumber" example:"123456789"`
	RoutingNumber       string               `json:"RoutingNumber" example:"023456789"`
	ACHAccountType      enums.ACHAccountType `json:"ACHAccountType" swaggertype:"string" example:"checking"`
	CardNumber          string               `json:"CardNumber" example:"123456789"`
	ZipCode             string               `json:"ZipCode" example:"12345"`
	CVV                 string               `json:"CVV" example:"123"`
	NickName            string               `json:"NickName" example:"John's Account"`
	CardExpirationMonth string               `json:"CardExpirationMonth" example:"12"`
	CardExpirationYear  string               `json:"CardExpirationYear" example:"2022"`
}

func (p *PaymentMethodRequest) IsACHPaymentMethodType() bool {
	return p.PaymentMethodType == enums.ACH
}

func (p *PaymentMethodRequest) IsCardPaymentMethodType() bool {
	return p.PaymentMethodType == enums.Card
}
