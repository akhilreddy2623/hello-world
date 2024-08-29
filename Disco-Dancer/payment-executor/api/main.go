package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-api/internal"
	proto "geico.visualstudio.com/Billing/plutus/proto/paymentmethodvalidation"
	"github.com/geico-private/pv-bil-frameworks/config"
	"google.golang.org/grpc"
)

const defaultWebPort = "30000"
const defaultGrpcAddress = "0.0.0.0:5051"

var log = logging.GetLogger("payment-executor-api")

type Config struct{}

func main() {

	configHandler := commonFunctions.GetConfigHandler()
	webPort := configHandler.GetString("WebServer.Port", defaultWebPort)
	database.Init(configHandler)

	app := Config{}

	log.Info(context.Background(), "starting payment executor service on port %s", webPort)

	go grpcMain(configHandler)
	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Error(context.Background(), err, "error occurred when starting the payment executor service")
	}
}

func grpcMain(configHandler config.AppConfiguration) {

	grpcAddress := configHandler.GetString("Grpc.Address", defaultGrpcAddress)
	lis, err := net.Listen("tcp", grpcAddress)

	if err != nil {
		log.Error(context.Background(), err, "grpc failed to listen on '%s'", grpcAddress)
	}

	log.Info(context.Background(), "grpc server listening on %s", grpcAddress)
	server := grpc.NewServer()
	proto.RegisterPaymentMethodValidationServiceServer(server, &internal.PaymentMethodValidationServer{})

	if err = server.Serve(lis); err != nil {
		log.Error(context.Background(), err, "grpc failed to serve on '%s'", grpcAddress)
	}

}
