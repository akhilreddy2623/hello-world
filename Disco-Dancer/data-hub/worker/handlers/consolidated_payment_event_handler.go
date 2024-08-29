package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/api"
	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/data-hub-common/models"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"github.com/geico-private/pv-bil-frameworks/kafkamessaging"
)

// TODO: Update all these constants
const (
	WorkdayRequstUrlDefault  = "http://localhost:30000/workday"
	WordayTimeoutDefault     = 30
	WorkdayRetryCountDefault = 5
)

var log = logging.GetLogger("data-hub-handlers")

// TDDO: Following line should be uncommented after config service is running as sidecar
// var configHandler = commonFunctions.GetConfigHandler(configservice.ConfigContext{ApplicationId: "data-hub", Env: "DV1"})
var configHandler = commonFunctions.GetConfigHandler()

var ConsolidatedExecutePaymentEventHandler = func(ctx context.Context, message *kafkamessaging.Message) error {
	log.Info(ctx, "Received message in paymentPlatform.executePaymentResponse topic: '%s'", *message.Body)

	consolidatedPaymentResponse := commonMessagingModels.ConsolidatedExecutePaymentResponse{}
	if err := json.Unmarshal([]byte(*message.Body), &consolidatedPaymentResponse); err != nil {
		log.Error(context.Background(), err, "unable to unmarshal executePaymentResponse")
		return err
	}

	vendorType, providerName := getVendorData(consolidatedPaymentResponse.PaymentRequestType)

	workdayDataRequest := models.WorkdayDataRequest{
		VendorType:             vendorType,
		PaymentReference:       consolidatedPaymentResponse.SettlementIdentifier,
		ProviderName:           providerName,
		PaymentDate:            consolidatedPaymentResponse.PaymentDate,
		Amount:                 consolidatedPaymentResponse.Amount,
		Last4AccountIdentifier: consolidatedPaymentResponse.Last4AccountIdentifier,
		Status:                 consolidatedPaymentResponse.Status,
	}

	return callWorkDayAPI(workdayDataRequest)
}

func callWorkDayAPI(consolidatedPaymentResponse models.WorkdayDataRequest) error {

	consolidatedPaymentResponseJson, err := json.Marshal(consolidatedPaymentResponse)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal consolidatedPaymentResponse")
		return err
	}

	workdayRequstUrl := configHandler.GetString("PaymentPlatform.Workday.Url", WorkdayRequstUrlDefault)
	workdayRetryCount := configHandler.GetInt("PaymentPlatform.Workday.RetryCount", WorkdayRetryCountDefault)
	workdayTimeout := configHandler.GetInt("PaymentPlatform.Workday.Timeout", WordayTimeoutDefault)

	timeoutDuration := time.Duration(workdayTimeout * int(time.Second))

	apiRequest := commonAppModels.APIRequest{
		Type:          enums.Workday,
		Url:           workdayRequstUrl,
		Request:       consolidatedPaymentResponseJson,
		Authorization: "",
		Timeout:       timeoutDuration,
	}

	apiResponse, err := api.RetryApiCall(api.PostRestApiCall, apiRequest, workdayRetryCount)
	if err != nil {
		log.Error(context.Background(), err, "call failed to %s", workdayRequstUrl)
		return err
	}

	// TODO: Handle actual response from workday
	bodyString := string(apiResponse.Response)

	log.Info(context.Background(), "Response From Workday API: %s", bodyString)

	return err
}

func getVendorData(paymentRequestType string) (string, string) {
	switch strings.ToLower(paymentRequestType) {
	case "insuranceautoauctions":
		return "IAA", "Insurance Auto Auctions Inc"
	}

	return "", ""
}
