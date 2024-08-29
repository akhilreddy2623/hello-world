package clients

import (
	"context"
	"crypto/tls"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const defaultGrpcAddress = "0.0.0.0:9092"

var log = logging.GetLogger("payment-administrator-client")

func getPaymentVaultConnection() (*grpc.ClientConn, error) {
	configHandler := commonFunctions.GetConfigHandler()
	grpcAddress := configHandler.GetString("PaymentPlatform.Grpc.Address.Vault", "")
	log.Info(context.Background(), "Connecting to payment vault at address: '%s'", grpcAddress)

	var conn *grpc.ClientConn
	var err error

	if grpcAddress == defaultGrpcAddress {
		conn, err = grpc.Dial(grpcAddress, grpc.WithInsecure())
	} else {
		cred := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})
		conn, err = grpc.Dial(grpcAddress, grpc.WithTransportCredentials(cred))
	}

	if err != nil {
		log.Error(context.Background(), err, "did not connect to server at address: '%s'", grpcAddress)
		return nil, err
	}

	return conn, nil
}
