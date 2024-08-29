package main

import (
	"context"
	"time"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/task-manager-worker/handlers"
)

var log = logging.GetLogger("task-manager-worker")

func main() {
	log.Info(context.Background(), "Starting task manager worker role")

	configHandler := commonFunctions.NewConfigHandler()
	database.Init(configHandler)

	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")

	messaging.InitKafka(brokers, kafkaConsumerGroupId)
	executeTaskResponseTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.ExecuteTaskResponses", "")
	messaging.KafkaSubscribe(executeTaskResponseTopic, handlers.ExecuteTaskResponseHandler)

	handlers.InitTaskSchedules()
	go handlers.CheckAndRunTaskSchedules()

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
