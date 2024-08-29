package external

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/api"
	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
)

const (
	ActionVerify              = "verify"
	Org_Id                    = "org_347080"
	Loc_Id                    = "loc_129706"
	ForteUrlDefault           = ""
	ForteAuthorizationDefault = ""
	ForteTimeoutDefault       = 30
	ForteRetryCountDefault    = 5
)

var log = logging.GetLogger("payment-executor-external")

var configHandler = commonFunctions.GetConfigHandler()

//go:generate mockery --name ValidationApiInterface
type ValidationApiInterface interface {
	ForteValidation(forteRequest app.ForteRequest) (*commonAppModels.ForteResponse, error)
}

type ForteApi struct {
}

func (ForteApi) ForteValidation(forteRequest app.ForteRequest) (*commonAppModels.ForteResponse, error) {
	forteUrl := getForteUrl()
	forteAuthorization := configHandler.GetString("PaymentPlatform.Forte.Authorization", ForteAuthorizationDefault)
	forteTimeOutInSec := configHandler.GetInt("PaymentPlatform.ExternalApi.Timeout", ForteTimeoutDefault)
	forteRetryCount := configHandler.GetInt("PaymentPlatform.ExternalApi.Retrycount", ForteRetryCountDefault)

	forteReq, err := json.Marshal(forteRequest)
	if err != nil {
		log.Error(context.Background(), err, "error in marshalling forte request")
		return nil, err
	}
	timeoutDuration := time.Duration(forteTimeOutInSec * int(time.Second))

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.Forte,
		Url:           forteUrl,
		Request:       forteReq,
		Authorization: fmt.Sprintf("Basic %s", forteAuthorization),
		Header:        Org_Id,
		Timeout:       timeoutDuration,
	}

	apiResponse, err := api.RetryApiCall(api.PostRestApiCall, apiRequest, forteRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "forte call is not succesful despite '%d' attempts", forteRetryCount)
		return nil, err
	}
	forteValidationResponse := commonAppModels.ForteResponse{}
	err = json.Unmarshal(apiResponse.Response, &forteValidationResponse)
	if err != nil {
		log.Error(context.Background(), err, "error in umarshalling forte response")
		return nil, err
	}
	return &forteValidationResponse, nil
}

func getForteUrl() string {
	forteUrl := configHandler.GetString("PaymentPlatform.Forte.Url", ForteUrlDefault)
	forteUrl = strings.Replace(forteUrl, "org_id", Org_Id, -1)
	forteUrl = strings.Replace(forteUrl, "loc_id", Loc_Id, -1)
	return forteUrl
}
