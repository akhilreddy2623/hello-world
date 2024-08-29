package integrationtest

import (
	"encoding/json"
	"regexp"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

var (
	messageToPublish string
)

func PublishMessageToTopic(topicName string, testConfig config.AppConfiguration, paymentMethodType string, paymentRequestType string, paymentFrequency string) {
	executetaskrequests := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	fileprocessingcompleted := testConfig.GetString("PaymentPlatform.Kafka.Topics.InboundFileFileProcessingCompleted", "")
	switch topicName {
	case executetaskrequests:
		messageToPublish = getMessageDetails(topicName, paymentMethodType, paymentRequestType, paymentFrequency)
	case fileprocessingcompleted:
		messageToPublish = getMessageDetails(topicName, "", "", "")
	}
	messaging.KafkaPublish(topicName, messageToPublish)
}

func getMessageDetails(topicName string, paymentMethodType string, paymentRequestType string, paymentFrequency string) string {
	var message interface{}
	date, _ := time.Parse(commonMessagingModels.TimeFormat, twoDaysAgo)

	switch topicName {
	case "paymentplatform.internal.executetaskrequests":
		message = commonMessagingModels.ExecuteTaskRequest{
			Version:         1,
			Component:       "executor",
			TaskName:        "settleachpayments",
			TaskDate:        commonMessagingModels.JsonDate(date),
			TaskExecutionId: 1,
			ExecutionParameters: commonMessagingModels.ExecutionParameters{
				PaymentMethodType:  paymentMethodType,
				PaymentRequestType: paymentRequestType,
				PaymentFrequency:   paymentFrequency,
			},
		}
	case "paymentplatform.inboundfileprocessor.fileprocessingcompleted":
		message = app.FileProcessingCompleted{
			BusinessFileType: "achonetimeack",
		}

	}
	messageJson, _ := json.Marshal(message)
	return string(messageJson)
}

func GetJsonDate(message *kafkamessaging.Message) string {
	// Preprocess the message body to remove the time component from the paymentDate field.
	re := regexp.MustCompile(`("PaymentDate"\s*:\s*")([^T]+)T[^"]*(")`)
	modifiedBody := re.ReplaceAllString(*message.Body, `$1$2$3`)

	return modifiedBody
}
