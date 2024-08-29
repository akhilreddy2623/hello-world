package main

import (
	"context"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-worker/handlers"
)

var log = logging.GetLogger("payment-administrator-worker")

func main() {

	log.Info(context.Background(), "Starting payment administrator worker role")

	configHandler := commonFunctions.GetConfigHandler()
	database.Init(configHandler)

	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")
	messaging.InitKafka(brokers, kafkaConsumerGroupId)

	executeTaskRequestTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskRequests", "")
	executePaymentResponseTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentResponses", "")
	handlers.ExecutePaymentRequestTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecutePaymentRequests", "")
	handlers.ExecuteTaskResponseTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")
	handlers.PaymentResponseTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentResponses", "")
	handlers.DataHubPaymentEventsTopic = configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")

	messaging.KafkaSubscribe(executeTaskRequestTopic, handlers.ProcessTaskRequestHandler)
	messaging.KafkaSubscribe(executePaymentResponseTopic, handlers.ExecutePaymentResponseHandler)

	keepAlive(context.Background())
}

func keepAlive(ctx context.Context) {
	for {
		log.Info(ctx, "Keepalive is still running.")
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Minute):
		}
	}
}
