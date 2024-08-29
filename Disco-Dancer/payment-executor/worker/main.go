package main

import (
	"context"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/database"

	filmapiclient "geico.visualstudio.com/Billing/plutus/filmapi-client"
	inbound_fileprocessor "geico.visualstudio.com/Billing/plutus/inbound-fileprocessor"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	settlement "geico.visualstudio.com/Billing/plutus/payment-executor-common/settlement/outbound"
	"geico.visualstudio.com/Billing/plutus/payment-executor-worker/handlers"
	"github.com/geico-private/pv-bil-frameworks/config"
)

var log = logging.GetLogger("payment-executor-worker")

func main() {

	log.Info(context.Background(), "Starting payment executor worker role")
	configHandler := commonFunctions.GetConfigHandler()
	database.Init(configHandler)

	volatageSideCarAddress := configHandler.GetString("Voltage.Sidecar.Address", "localhost:50051")
	crypto.Init(volatageSideCarAddress)

	url := configHandler.GetString("Filmapi.Url", "https://gze-flmapi-dv1-app.gze-bllase-np1-ase.appserviceenvironment.net/")
	authorization := configHandler.GetString("Filmapi.Authorization", "c3VwcG9ydHVzZXJfbnA6V1RTOFpFQFUsMzZXZ2pl")
	timeout := configHandler.GetInt("ExternalApi.Timeout", 10)
	retryCount := configHandler.GetInt("ExternalApi.Retrycount", 5)
	filmapiclient.Init(url, authorization, timeout, retryCount, nil)

	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")
	messaging.InitKafka(brokers, kafkaConsumerGroupId)

	var executePaymentRequestTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentRequests", "")
	var settlePaymentsTaskRequestsTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	var inboundFileProcessRecordTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.InboundFileProcessRecord", "")
	var inboundFileProcessRecordFeedbackTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.InboundFileProcessRecordFeedback", "")
	var fileProcessingCompletedTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")
	handlers.ExecuteTaskResponseTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")
	handlers.DataHubPaymentEventsTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")
	settlement.Outboundfileprocessingpath = configHandler.GetString("OutboundFileProcessingPath.InsuranceAutoAuctions", "")

	messaging.KafkaSubscribe(executePaymentRequestTopic, handlers.ExecutePaymentRequestHandler)
	messaging.KafkaSubscribe(settlePaymentsTaskRequestsTopic, handlers.PaymentSettlementTaskHandler)
	messaging.KafkaSubscribe(inboundFileProcessRecordTopic, handlers.InboundFileProcessRecordHandler)
	messaging.KafkaSubscribe(fileProcessingCompletedTopic, handlers.FileProcessingCompletedHandler)

	inbound_fileprocessor.Init(inboundFileProcessRecordFeedbackTopic)
	inboundFileProcessing(configHandler, inboundFileProcessRecordTopic)

	keepAlive(context.Background())

}

func inboundFileProcessing(configHandler config.AppConfiguration, inboundFileProcessRecordTopic string) {
	go achOneTimeAckFile(configHandler, inboundFileProcessRecordTopic)
}

func achOneTimeAckFile(configHandler config.AppConfiguration, inboundFileProcessRecordTopic string) {

	var pollingInterval = configHandler.GetInt("InboundFileProcessing.AchOnetimeAck.Polling.Interval", 10)
	var folderLocation = configHandler.GetString("InboundFileProcessing.AchOnetimeAck.Folder.Location", "")
	var archiveLocation = configHandler.GetString("InboundFileProcessing.AchOnetimeAck.Archive.Location", "")

	fileProcessorInput := inbound_fileprocessor.FileProcessorInput{
		PollingDurationInMins:  pollingInterval,
		FolderLocation:         folderLocation,
		BusinessFileType:       "achonetimeack",
		ArchiveFolderLocation:  archiveLocation,
		ProcessRecordTopicName: inboundFileProcessRecordTopic,
	}
	inbound_fileprocessor.Process(fileProcessorInput)
}

func keepAlive(ctx context.Context) {
	for {
		log.Info(ctx, "payment-executor - Keepalive is still running.")
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Minute):
		}
	}
}
