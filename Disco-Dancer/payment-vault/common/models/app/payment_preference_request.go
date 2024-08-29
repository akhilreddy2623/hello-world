package app

import "geico.visualstudio.com/Billing/plutus/enums"

// TODO: it will be added after agent and agency use cases are sort out
//AccountId int64
type PaymentPreferenceRequest struct {
	UserId             string                   `json:"UserId" example:"125"`
	ProductType        enums.ProductType        `json:"ProductType" swaggertype:"string" example:"ppa"`
	ProductSubType     enums.ProductSubType     `json:"ProductSubType" swaggertype:"string" example:"auto"`
	PaymentRequestType enums.PaymentRequestType `json:"PaymentRequestType" swaggertype:"string" example:"enterpriserental"`
	ProductIdentifier  string                   `json:"ProductIdentifier" example:"ALL"`
	PayIn              []*PaymentPreferenceDetail
	PayOut             []*PaymentPreferenceDetail
}

type PaymentPreferenceDetail struct {
	PaymentMethodId   int64 `json:"PaymentMethodId" example:"1"`
	Split             int16 `json:"Split" example:"80"`
	AutoPayPreference bool  `json:"AutoPayPreference" example:"true"`
}
