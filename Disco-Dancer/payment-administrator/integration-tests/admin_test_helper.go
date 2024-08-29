package integrationtest

import (
	"encoding/json"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"github.com/geico-private/pv-bil-frameworks/config"
)

var (
	messageToPublish string
)

func PublishMessageToTopic(topicName string, testConfig config.AppConfiguration, paymentMethodType string, paymentRequestType string, paymentFrequency string) {

	executePaymentResponseTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")
	executeTaskRequestsTopic := testConfig.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")

	switch topicName {
	case executePaymentResponseTopic:
		messageToPublish = getMessageDetails(topicName, "", "", "")
	case executeTaskRequestsTopic:
		messageToPublish = getMessageDetails(topicName, paymentMethodType, paymentRequestType, paymentFrequency)
	}

	messaging.KafkaPublish(topicName, messageToPublish)
}

func getMessageDetails(topicName string, paymentMethodType string, paymentRequestType string, paymentFrequency string) string {
	var message interface{}
	date, _ := time.Parse(commonMessagingModels.TimeFormat, "2024-04-16")

	switch topicName {
	case "paymentplatform.internal.executepaymentresponses":
		message = commonMessagingModels.ExecutePaymentResponse{
			Version:                1,
			PaymentId:              1,
			Status:                 enums.Completed,
			SettlementIdentifier:   "1234567890",
			PaymentDate:            commonMessagingModels.JsonDate(date),
			Amount:                 25.16,
			Last4AccountIdentifier: "7890",
		}

	case "paymentplatform.internal.executetaskrequests":

		message = commonMessagingModels.ExecuteTaskRequest{
			Version:         1,
			Component:       "administrator",
			TaskName:        "processPayments",
			TaskDate:        commonMessagingModels.JsonDate(date),
			TaskExecutionId: 1,
			ExecutionParameters: commonMessagingModels.ExecutionParameters{
				PaymentMethodType:  paymentMethodType,
				PaymentRequestType: paymentRequestType,
				PaymentFrequency:   paymentFrequency,
			},
		}
	}

	messageJson, _ := json.Marshal(message)
	return string(messageJson)
}
