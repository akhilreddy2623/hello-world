package app

import (
	"encoding/json"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/models"
)

type PaymentRequest struct {
	TenantId                      int64  `json:"tenantid" example:"101"`
	TenantRequestId               int64  `json:"tenantRequestId" example:"1234567"`
	UserId                        string `json:"userId" example:"125"`
	AccountId                     string `json:"accountId" example:"PQR123"`
	ProductIdentifier             string `json:"productIdentifier" example:"ALL"`
	PaymentFrequency              string `json:"paymentFrequency" example:"onetime"`
	PaymentExtractionSchedule     []models.PaymentExtractionSchedule
	TransactionType               string                   `json:"transactionType" example:"payout"`
	CallerApp                     string                   `json:"callerApp" example:"mcp"`
	Metadata                      json.RawMessage          `json:"metadata" swaggertype:"string" example:"{}"`
	PaymentRequestType            string                   `json:"paymentRequestType" example:"iaa"`
	PaymentExtractionScheduleJson json.RawMessage          `swaggerignore:"true"`
	PaymentFrequencyEnum          enums.PaymentFrequency   `swaggerignore:"true"`
	TransactionTypeEnum           enums.TransactionType    `swaggerignore:"true"`
	PaymentRequestTypeEnum        enums.PaymentRequestType `swaggerignore:"true"`
	CallerAppEnum                 enums.CallerApp          `swaggerignore:"true"`
}

type PaymentResponse struct {
	TenantRequestId int64                          `json:"tenantRequestId" example:"123456789"`
	Status          string                         `json:"status" example:"accepted"`
	ErrorDetails    *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}
