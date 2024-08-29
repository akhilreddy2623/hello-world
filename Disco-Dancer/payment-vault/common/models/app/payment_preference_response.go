package app

import (
	"time"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
)

type PaymentPreferenceResponseList struct {
	Status             string                         `json:"status"`
	PaymentPreferences []*PaymentPreferenceResponse   `json:"PaymentPreferences,omitempty"`
	ErrorDetails       *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}

type PaymentPreferenceResponse struct {
	PaymentMethodType      string
	AccountIdentifier      string
	RoutingNumber          string                         `json:"RoutingNumber,omitempty" example:"023456789"`
	Last4AccountIdentifier string                         `json:"Last4AccountIdentifier" example:"6789"`
	PaymentExtendedData    string                         `json:"PaymentExtendedData" example:"{}"`
	WalletStatus           bool                           `json:"WalletStatus" example:"true"`
	PaymentMethodStatus    string                         `json:"PaymentMethodStatus" example:"ACTIVE"`
	AccountValidationDate  time.Time                      `json:"AccountValidationDate" example:"2021-01-01T00:00:00Z"`
	WalletAccess           bool                           `json:"WalletAccess" example:"true"`
	AutoPayPreference      bool                           `json:"AutoPayPreference" example:"true"`
	Split                  int32                          `json:"Split" example:"80"`
	Status                 string                         `json:"Status" example:"ACCEPTED"`
	ErrorDetails           *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}

type PaymentExtendedData struct {
	FirstName           string `json:"FirstName,omitempty"`
	LastName            string `json:"LastName,omitempty"`
	CheckAccountingType string `json:"CheckAccountingType,omitempty"`
	ZipCode             string `json:"ZipCode,omitempty"`
	CVV                 string `json:"CVV,omitempty"`
	CardExpirationMonth string `json:"CardExpirationMonth,omitempty"`
	CardExpirationYear  string `json:"CardExpirationYear,omitempty"`
}

type StorePaymentPreferenceResponse struct {
	Status              string                         `json:"status,omitempty" example:"ACCEPTED"`
	PaymentPreferenceId int64                          `json:"paymentpreferenceid,omitempty" example:"123456789"`
	ErrorDetails        *commonAppModels.ErrorResponse `json:"errordetails,omitempty"`
}

func CreateStorePaymentPreferenceResponse(paymentPreferenceId *int64, errorDetails *commonAppModels.ErrorResponse) *StorePaymentPreferenceResponse {
	if errorDetails == nil {
		return &StorePaymentPreferenceResponse{
			Status:              "ACCEPTED",
			PaymentPreferenceId: *paymentPreferenceId,
		}
	} else {
		return &StorePaymentPreferenceResponse{
			Status:       "ERRORED",
			ErrorDetails: errorDetails,
		}
	}
}
