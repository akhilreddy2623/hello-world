package app

import (
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
)

type PaymentMethodValidationResponse struct {
	Status             enums.PaymentMethodValidationStatus
	ValidationResponse commonAppModels.ForteResponse
}
