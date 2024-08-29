package main

import (
	"context"
	"fmt"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/config-manager-api/internal"
	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/geico-private/pv-bil-frameworks/logging"
)

var log = logging.GetLogger("config-manager-api")

type Config struct{}

func main() {

	// Initiate configuration setup
	configHandler := commonFunctions.GetConfigHandler()

	webPort := configHandler.GetString("PaymentPlatform.WebServer.Port", "")

	database.Init(configHandler)
	go internal.InitiateConfigManagerServer(configHandler)
	app := Config{}

	log.Info(context.Background(), "starting config manager service on port %s", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Error(context.Background(), err, "error occurred when starting the config manager service")
	}
}
