package models

import "time"

type WorkdayDataRequest struct {
	VendorType             string
	PaymentReference       string
	ProviderName           string
	PaymentDate            time.Time
	Amount                 float32
	Last4AccountIdentifier string
	Status                 string
}
