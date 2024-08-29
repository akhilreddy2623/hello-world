package clients

import (
	"context"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultGrpcAddress = "0.0.0.0:5052"

var log = logging.GetLogger("payment-vault-client")
var configHandler = commonFunctions.GetConfigHandler()

func getExecutorConnection() (*grpc.ClientConn, error) {
	grpcAddress := configHandler.GetString("Grpc.Address.Executor", defaultGrpcAddress)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Error(context.Background(), err, "did not connect to server at address: '%s'", grpcAddress)
		return nil, err
	}

	return conn, nil
}
