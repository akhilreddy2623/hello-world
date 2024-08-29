package handlers

import (
	"context"
	"encoding/json"

	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	inbound_settlement_postprocessor "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/inbound/postprocessors"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var FileProcessingCompletedHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "PaymentExecutor - Received message in paymentplatform.inboundfileprocessor.fileprocessingcompleted topic to process file record: '%s'", message.MessageID)

	fileProcessingCompleted := app.FileProcessingCompleted{}

	err := json.Unmarshal([]byte(*message.Body), &fileProcessingCompleted)
	if err != nil {
		log.Error(ctx, err, "error in unmarshalling fileProcessingCompleted message")
		return err
	}

	switch fileProcessingCompleted.BusinessFileType {
	case "achonetimeack":
		{
			err = inbound_settlement_postprocessor.AchOnetimeAckPostProcessor()
		}
	}
	if err != nil {
		log.Error(ctx, err, "error while post processing application level inbound file post processor for businesstype '%s'", fileProcessingCompleted.BusinessFileType)
		return err
	}
	return nil
}
