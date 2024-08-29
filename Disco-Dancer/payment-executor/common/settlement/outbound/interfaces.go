package settlement

import (
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
)

type DataReaderInterface interface {
	ReadData(time.Time, db.ExecutionParameters) ([]db.SettlementPayment, error)
}

type DataMapperInterface interface {
	MapData([]db.SettlementPayment) ([]app.SettlementPayment, error)
}

type FileGeneratorInterface interface {
	GenerateFile([]app.SettlementPayment) (string, error)
}

type FileUploaderInterface interface {
	UploadFile(enums.PaymentRequestType, string) error
}

type FilePostProcessorInterface interface {
	FilePostProcessor([]db.SettlementPayment) error
}

type FilePreProcessorInterface interface {
	FilePreProcessor([]db.SettlementPayment) ([]db.SettlementPayment, error)
}
