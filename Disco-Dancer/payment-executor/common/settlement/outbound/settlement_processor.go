package settlement

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
)

var dataReader DataReaderInterface
var dataMapper DataMapperInterface
var fileGenerator FileGeneratorInterface
var fileUploader FileUploaderInterface
var filePostProcessor FilePostProcessorInterface
var settlementPreProcessor SettlementPreProcessor

var log = logging.GetLogger("payment-executor-settlement")

func ProcessOutbound(date time.Time, executionParameters db.ExecutionParameters) error {

	populateInterfaces()

	data, err := dataReader.ReadData(date, executionParameters)
	if err != nil {
		return err
	}

	// If no data found then file processing should exit here

	if len(data) == 0 {
		log.Info(context.Background(), "Payment Executor - No data found to create outbound file. Exiting...")
	} else {

		data, err := settlementPreProcessor.FilePreProcessor(data)
		if err != nil {
			return err
		}

		mappedData, err := dataMapper.MapData(data)
		if err != nil {
			return err
		}

		fileStrSlice, err := fileGenerator.GenerateFile(mappedData)
		if err != nil {
			return err
		}

		err = fileUploader.UploadFile(executionParameters.PaymentRequestType, fileStrSlice)
		if err != nil {
			return err
		}

		err = filePostProcessor.FilePostProcessor(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func populateInterfaces() {

	dataReader = DataReader{}
	dataMapper = DataMapper{}
	fileGenerator = FileGenerator{}
	fileUploader = FileUploader{}
	filePostProcessor = FilePostProcessor{}
	settlementPreProcessor = SettlementPreProcessor{}
}
