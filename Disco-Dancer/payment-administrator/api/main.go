package main

import (
	"context"
	"fmt"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-api/handlers"
)

// @title           GEICO Payment Platform Administrator API
// @version         1.0
// @description     Responsible for handling incoming payment requests and managing payment schedules.
// @contact.name   	Payment Platform Dev Team
// @contact.email  	GPP@geico.com

type Config struct{}

var log = logging.GetLogger("payment-administrator-api")

func main() {

	configHandler := commonFunctions.GetConfigHandler()
	webPort := configHandler.GetString("PaymentPlatform.WebServer.Port", "")

	//getting list of valid tenantIds from config
	// 101 - claims; 10,11 - integration testing
	handlers.ValidTenantIds = configHandler.GetList("PaymentPlatform.MakePayment.ValidTenantIds")
	handlers.MaxPaymentAmountIaa = configHandler.GetInt("PaymentPlatform.MakePayment.MaxPaymentAmount.InsuranceAutoAuctions", 0)

	database.Init(configHandler)

	app := Config{}

	brokers := configHandler.GetList("PaymentPlatform.Kafka.Brokers")
	kafkaConsumerGroupId := configHandler.GetString("PaymentPlatform.Kafka.ConsumerGroupId", "")
	messaging.InitKafka(brokers, kafkaConsumerGroupId)

	log.Info(context.Background(), "starting payment administrator service on port %s", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Error(context.Background(), err, "error occurred when starting the payment administrator service")
	}

}
