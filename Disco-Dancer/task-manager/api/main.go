package main

import (
	"context"
	"fmt"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
)

const defaultWebPort = "30000"

var log = logging.GetLogger("task-manager-api")

type Config struct{}

var configHandler = commonFunctions.GetConfigHandler()

func main() {

	webPort := configHandler.GetString("WebServer.Port", defaultWebPort)
	database.Init(configHandler)
	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")
	messaging.InitKafka(brokers, kafkaConsumerGroupId)
	app := Config{}

	log.Info(context.Background(), "starting task manager service on port %s", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Error(context.Background(), err, "error occurred when starting the task manager service")
	}
}
