package handlers

import (
	"context"
	"encoding/json"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	inbound_settlement "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/inbound"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var InboundFileProcessRecordHandler = func(ctx context.Context, message *kafkamessaging.Message) error {

	log.Info(ctx, "PaymentExecutor - Received message in paymentplatform.inboundfileprocessor.processrecord topic to process file record: '%s'", message.MessageID)

	fileRecord := app.FileRecord{}

	err := json.Unmarshal([]byte(*message.Body), &fileRecord)
	if err != nil {
		log.Error(ctx, err, "error in unmarshalling fileRecord message")
		return err
	}
	isError := false
	switch fileRecord.BusinessFileType {
	case "achonetimeack":
		{
			err = inbound_settlement.AchOnetimeAckProcessor(fileRecord)
		}
	}

	if err != nil {
		log.Error(ctx, err, "error while processing application level inbound file processor for businesstype '%s'", fileRecord.BusinessFileType)
		isError = true
	}

	configHandler := commonFunctions.GetConfigHandler()
	var recordFeedbackTopicName = configHandler.GetString("PaymentPlatform.Kafka.Topics.InboundFileProcessRecordFeedback", "")

	err = messaging.KafkaPublish(recordFeedbackTopicName, getFileRecordFeedback(fileRecord, isError))
	if err != nil {
		log.Error(ctx, err, "error in publishing record feedback in '%s'' topic", recordFeedbackTopicName)
		return err
	}
	return nil
}

func getFileRecordFeedback(fileRecord app.FileRecord, isError bool) string {

	fileRecordFeedback := app.FileRecordFeedback{
		FileId:                fileRecord.FileId,
		RecordId:              fileRecord.RecordId,
		BusinessFileType:      fileRecord.BusinessFileType,
		TotalRecordCount:      fileRecord.TotalRecordCount,
		FilePath:              fileRecord.FilePath,
		ArchiveFolderLocation: fileRecord.ArchiveFolderLocation,
		IsError:               isError,
	}
	recordJson, err := json.Marshal(fileRecordFeedback)
	if err != nil {
		log.Error(context.Background(), err, "error in marshalling file record")
	}
	return string(recordJson)
}
