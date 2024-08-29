package internal

import (
	"context"
	"errors"
	"fmt"
	"net"

	"geico.visualstudio.com/Billing/plutus/config-manager-common/repository"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/config/configservice/proto"
	"github.com/geico-private/pv-bil-frameworks/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Configuration struct {
}

var log = logging.GetLogger("config-manager-api-internal")

func InitiateConfigManagerServer(configHandler config.AppConfiguration) {
	grpcAddress := configHandler.GetString("PaymentPlatform.Grpc.Address", "")
	lis, err := net.Listen("tcp", grpcAddress)

	if err != nil {
		log.Error(context.Background(), err, "grpc failed to listen on '%s'", grpcAddress)
	}

	log.Info(context.Background(), "grpc server listening on %s", grpcAddress)
	server := grpc.NewServer()

	proto.RegisterConfigServiceServer(server, &Server{})

	if err = server.Serve(lis); err != nil {
		log.Error(context.Background(), err, "grpc failed to serve on '%s'", grpcAddress)
	}
}

func (*Server) GetConfig(ctx context.Context, request *proto.ConfigRequest) (*proto.ConfigResponse, error) {
	var applicationName = request.ConfigContext.Application
	var env = request.ConfigContext.Env

	var configManagerRepository repository.ConfigRepositoryInterface = &repository.ConfigRepository{}
	return getConfigDetails(applicationName, env, configManagerRepository)

}
func getConfigDetails(applicationName string, env string, configManagerRepository repository.ConfigRepositoryInterface) (*proto.ConfigResponse, error) {

	configDetails, err := configManagerRepository.GetConfig(applicationName, env)
	if err != nil {
		log.Error(context.Background(), err, "error occured while getting configuration details for application %s", applicationName)
		return nil, status.Errorf(
			codes.Unknown,
			fmt.Sprintf("error getting configuration details for application %s", applicationName))
	}

	if len(*configDetails) == 0 {
		log.Error(context.Background(), errors.New("error occurred"), "no records found %s", applicationName)
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("no records found for %s", applicationName))
	}

	var configs []*proto.Config

	for _, record := range *configDetails {
		configs = append(configs, &proto.Config{
			Key:         record.Key,
			Value:       record.Value,
			Env:         record.Environment,
			Application: record.Application,
			Tenant:      record.Tenant,
			Vendor:      record.Vendor,
			Product:     record.Product,
		})
	}

	return &proto.ConfigResponse{ConfigDetails: configs}, nil
}
