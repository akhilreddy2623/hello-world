package db

import "geico.visualstudio.com/Billing/plutus/enums"

type StorePaymentPreference struct {
	PreferenceId  int64
	ProductDetail ProductDetail
	PayIn         []*PaymentPreferenceExtendedData `json:",omitempty"`
	PayOut        []*PaymentPreferenceExtendedData `json:",omitempty"`
}

type PaymentPreferenceExtendedData struct {
	PaymentMethodId   int64 `json:"PaymentMethodId"`
	Split             int16 `json:"Split"`
	AutoPayPreference bool  `json:"AutoPayPreference"`
}

type ProductDetail struct {
	ProductDetailId    int64
	UserId             string
	PaymentRequestType enums.PaymentRequestType
	ProductType        enums.ProductType
	ProductSubType     enums.ProductSubType
	ProductIdentifier  string
}
