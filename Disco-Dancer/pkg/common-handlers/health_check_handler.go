package commonHandelers

import (
	"context"
	"encoding/json"
	"net/http"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	"geico.visualstudio.com/Billing/plutus/logging"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonRepositories "geico.visualstudio.com/Billing/plutus/common-repositories"
)

var log = logging.GetLogger("health-check-handler")

type HealthCheckHandler struct {
}

// HealthCheckHandler is the HTTP handler function for getting status of the application

// @Summary		    HealthCheck of the application
// @Description	    This API returns the Health Check Status of the application
// @Tags			HealthCheck
// @Produce		    application/json
// @Response		default	{object}	commonAppModels.HealthCheckResponse "Health Check Results"
// @Router			/health/ready [get]
func (HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	databasePingError := getDatabaseStatus()
	databaseStatus := "Up"
	var errorResponse commonAppModels.ErrorResponse
	configHandler := commonFunctions.GetConfigHandler()

	if databasePingError != nil {
		databaseStatus = "down"

		errorResponse = commonAppModels.ErrorResponse{
			Type:       "DatabasePingError",
			Message:    databasePingError.Error(),
			StatusCode: http.StatusInternalServerError}
	}

	healthCheckStatus := commonAppModels.HealthCheckResponse{
		Error:             databasePingError != nil,
		Message:           "I am alive !!",
		DatabaseStatus:    databaseStatus,
		ConfigFileName:    configHandler.GetString("FileName", "Not Available"),
		SeceretsAvailable: configHandler.GetString("PaymentPlatform.Testing.availability", "No"),
		ErrorResponse:     errorResponse,
	}

	out, _ := json.MarshalIndent(healthCheckStatus, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)
}

func getDatabaseStatus() error {
	var baseRepository commonRepositories.BaseRepositoryInterface = commonRepositories.Baserepository{}
	err := baseRepository.PingDatabase()
	if err != nil {
		log.Error(context.Background(), err, "Unable to ping database.")
	}
	return err
}
