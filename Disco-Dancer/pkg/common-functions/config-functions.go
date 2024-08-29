package commonFunctions

import (
	"context"
	"sync"
	"time"

	"geico.visualstudio.com/Billing/plutus/logging"
	"github.com/geico-private/pv-bil-frameworks/config"
	"github.com/geico-private/pv-bil-frameworks/config/configservice"
)

var once sync.Once
var configInstance config.AppConfiguration
var log = logging.GetLogger("config-functions")

// TODO : if it is going to be side car, this will work all the time
var grpcConfigManagerEndpoint = "localhost:5051"

func GetConfigHandler(configContext ...configservice.ConfigContext) config.AppConfiguration {

	once.Do(func() {
		config.Init()
		// if application is provided config context , then get config value from config service
		if len(configContext) > 0 {

			customProviders, err := configservice.NewConfigServiceProvider(
				configservice.WithDefaultContext(configContext[0]),
				configservice.WithConfigServiceTransport(configservice.GrpcConfigServiceTransport{GrpcEndpoint: grpcConfigManagerEndpoint}),
				configservice.WithCacheRefreshInterval(30*time.Minute),
			)
			if err != nil {
				log.Error(context.Background(), err, "Error creating custom config provider or Unable to connect to config service.")
				// Do not start Application if config service is not available
				panic(err)
			}

			configInstance = config.NewConfigBuilder().
				AddDefaults().
				AddCustomProvider(customProviders).
				Build()
		} else {
			configInstance = config.NewConfigBuilder().
				AddDefaults().
				Build()
		}
	})

	return configInstance
}

// This method is used for resetting the config handler for the Integration Test
func SetConfigHandler(c config.AppConfiguration) {
	configInstance = c
}

// this method is now depricated and should not be used, its here just to make sure non migrated code does not break
// This will be removed once all the services have been migrated to new the function GetConfigHandler()
func NewConfigHandler() config.AppConfiguration {
	config.Init()
	config := config.
		NewConfigBuilder().
		AddDefaults().
		Build()
	return config
}
