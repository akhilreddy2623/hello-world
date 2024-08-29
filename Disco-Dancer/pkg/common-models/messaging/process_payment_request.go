package messaging

type ProcessPaymentRequest struct {
	TenantId               int64
	PaymentId              int64
	PaymentFrequency       string
	TransactionType        string
	PaymentRequestType     string
	PaymentMethodType      string
	PaymentExtendedData    string
	AccountIdentifier      string
	Last4AccountIdentifier string
	RoutingNumber          string
	Amount                 float32
	PaymentDate            JsonDate
}
