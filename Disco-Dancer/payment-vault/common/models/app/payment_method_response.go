package app

import commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"

type PaymentMethodResponse struct {
	Status          string                         `json:"status,omitempty" example:"ACCEPTED"`
	PaymentMethodId int64                          `json:"paymentmethodid,omitempty" example:"123456789"`
	ErrorDetails    *commonAppModels.ErrorResponse `json:"errordetails,omitempty"`
}

func CreateStorePaymentMethodResponse(paymentmethodId *int64, errorDetails *commonAppModels.ErrorResponse) *PaymentMethodResponse {
	if errorDetails == nil {
		return &PaymentMethodResponse{
			Status:          "ACCEPTED",
			PaymentMethodId: *paymentmethodId,
		}
	} else {
		return &PaymentMethodResponse{
			Status:       "ERRORED",
			ErrorDetails: errorDetails,
		}
	}
}
