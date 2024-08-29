package main

import (
	"context"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/data-hub-worker/handlers"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
)

var log = logging.GetLogger("data-hub")

func main() {

	log.Info(context.Background(), "Starting data-hub worker role")

	// TDDO: Following two lines should be uncommented after config service is running as sidecar
	//configContext := configservice.ConfigContext{ApplicationId: "data-hub", Env: "DV1"}
	//configHandler := commonFunctions.GetConfigHandler(configContext)

	configHandler := commonFunctions.GetConfigHandler()

	// TODO: it should be removed later, just for testing get key from config service
	//log.Info(context.Background(), "Task Manager URL: %s", configHandler.GetString("PaymentPlatform.TaskManager.Url", "localhost:30000/task"))

	database.Init(configHandler)

	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")
	messaging.InitKafka(brokers, kafkaConsumerGroupId)

	executePaymentResponseTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.ConsolidatedPayments", "")
	paymentEventsTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")

	messaging.KafkaSubscribe(executePaymentResponseTopic, handlers.ConsolidatedExecutePaymentEventHandler)
	messaging.KafkaSubscribe(paymentEventsTopic, handlers.PaymentEventBalancingHandler)

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
