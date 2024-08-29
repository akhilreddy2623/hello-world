package messaging

import (
	"encoding/json"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
)

type ExecutePaymentResponse struct {
	Version                int                            `json:"version"`
	TenantId               int64                          `json:"tenantId,omitempty"`
	TenantRequestId        int64                          `json:"tenantRequestId,omitempty"`
	PaymentId              int64                          `json:"paymentId,omitempty"`
	SettlementIdentifier   string                         `json:"settlementIdentifier"`
	Status                 enums.PaymentStatus            `json:"status"`
	Amount                 float32                        `json:"amount"`
	PaymentDate            JsonDate                       `json:"paymentDate"`
	Last4AccountIdentifier string                         `json:"last4AccountIdentifier"`
	Metadata               json.RawMessage                `json:"metadata,omitempty"`
	ErrorDetails           *commonAppModels.ErrorResponse `json:"errorDetails,omitempty"`
}
