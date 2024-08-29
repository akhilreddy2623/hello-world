package db

import (
	"encoding/json"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/models"
)

type IncomingPaymentRequest struct {
	RequestId                     int64
	TenantId                      int64
	TenantRequestId               int64
	UserId                        string
	AccountId                     string
	ProductIdentifier             string
	PaymentExtractionSchedule     []models.PaymentExtractionSchedule
	Metadata                      json.RawMessage
	PaymentExtractionScheduleJson json.RawMessage
	PaymentFrequencyEnum          enums.PaymentFrequency
	TransactionTypeEnum           enums.TransactionType
	PaymentRequestTypeEnum        enums.PaymentRequestType
	CallerAppEnum                 enums.CallerApp
	PaymentMethodType             enums.PaymentMethodType
}

type Payment struct {
	PaymentId            int64
	RequestId            int64
	TenantRequestId      int64
	UserId               string
	AccountId            string
	ProductIdentifier    string
	PaymentFrequencyEnum enums.PaymentFrequency
	PaymentDate          time.Time
	Amount               float32
	Status               int
}
