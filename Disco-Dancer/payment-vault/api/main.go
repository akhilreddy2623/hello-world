package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-vault-api/internal"
	proto "geico.visualstudio.com/Billing/plutus/proto/paymentpreference"
	"github.com/geico-private/pv-bil-frameworks/config"

	// rkboot "github.com/rookie-ninja/rk-boot"
	// rkgrpc "github.com/rookie-ninja/rk-grpc/boot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

// @title           GEICO Payment Platform Vault API
// @version         1.0
// @description     Responsible for interacting with GEICO Payment Platform Vault.
// @contact.name   Payment Platform Dev Team
// @contact.email  GPP@geico.com

const defaultWebPort = "30000"
const defaultGrpcAddress = "0.0.0.0:5051"

var (
	log        = logging.GetLogger("payment-vault-api")
	grpcSystem = "" // empty string represents the health of the entire grpc system
)

type Config struct{}

func main() {
	configHandler := commonFunctions.GetConfigHandler()
	webPort := configHandler.GetString("WebServer.Port", defaultWebPort)
	volatageSideCarAddress := configHandler.GetString("Voltage.Sidecar.Address", "localhost:50051")

	database.Init(configHandler)
	crypto.Init(volatageSideCarAddress)

	app := Config{}

	log.Info(context.Background(), "starting payment vault service on port %s", webPort)

	go grpcMain(configHandler)

	// go grpcSwagger()

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Error(context.Background(), err, "error occurred when starting the payment vault service")
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
	healthcheck := health.NewServer()
	healthcheck.SetServingStatus(grpcSystem, healthgrpc.HealthCheckResponse_SERVING)
	healthgrpc.RegisterHealthServer(server, healthcheck)
	proto.RegisterPaymentPreferenceServiceServer(server, &internal.Server{})
	if err = server.Serve(lis); err != nil {
		log.Error(context.Background(), err, "grpc failed to serve on '%s'", grpcAddress)
	}

}

// func grpcSwagger() {
// 	boot := rkboot.NewBoot()
// 	grpcEntry := boot.GetEntry("paymentvaultgrpc").(*rkgrpc.GrpcEntry)
// 	grpcEntry.AddRegFuncGrpc(func(server *grpc.Server) { proto.RegisterPaymentReferenceServiceServer(server, &internal.Server{}) })

// 	grpcEntry.AddRegFuncGw(proto.RegisterPaymentReferenceServiceHandlerFromEndpoint)

// 	boot.Bootstrap(context.Background())

// 	// Wait for shutdown sig
// 	boot.WaitForShutdownSig(context.Background())
// }
