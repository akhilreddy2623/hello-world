package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-administrator-common/repository"
	"geico.visualstudio.com/Billing/plutus/validations"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	appmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/app"
	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
)

const (
	invalidJson                = "invalidJson"
	invalidField               = "invalidField"
	internalServerError        = "internalServerError"
	datanotfound               = "data not found"
	invalidJsonResponseMessage = "invalid response, unable to encode the json payload"
)

var errorMessageAndStatusCodeMapping = map[int]string{
	http.StatusBadRequest: "unable to process payment, as previous payment request with same tenantid and tenantRequestId already exist",
	http.StatusNotFound:   "unable to get payment preferences from payment vault",
}

var handlerLog = logging.GetLogger("payment-administrator-handler")
var ValidTenantIds []string
var MaxPaymentAmountIaa int

//go:generate mockery --name PaymentHandlerInterface
type PaymentHandlerInterface interface {
	PublishPaymentBalancingEventMessage(
		payments []dbmodels.Payment,
		incomingPaymentRequest dbmodels.IncomingPaymentRequest) error
}

type PaymentHandler struct {
}

// MakePaymentHandler is the HTTP handler function for making payment request to GEICO Payment Platform
//
//	@Summary		Make Payment request to GEICO Payment Platform
//	@Description	This API accepts a request to make a payment through GEICO Payment Platform. Onboarding teams should onboard as a Tenant and setup a payment request type. Drop email to GPP@geico.com for more details
//	@Tags			Administrator
//	@Accept			json
//	@Produce		json
//	@Param			PaymentRequest	body		appmodels.PaymentRequest	true	"PaymentRequest object that needs to be processed"
//	@Response		default			{object}	appmodels.PaymentResponse	"Status of payment request, errorDetails will be only returned in case of 4XX or 5XX error"
//	@Router			/payment [post]
func (PaymentHandler) MakePaymentHandler(w http.ResponseWriter, r *http.Request) {
	var paymentRequest appmodels.PaymentRequest
	var err error
	jsonErr := json.NewDecoder(r.Body).Decode(&paymentRequest)
	paymentRequest.PaymentExtractionScheduleJson, err = json.Marshal(paymentRequest.PaymentExtractionSchedule)

	if jsonErr != nil || err != nil {
		handlerLog.Error(context.Background(), jsonErr, invalidJson)
		writeResponse(
			w,
			paymentRequest.TenantRequestId,
			&commonAppModels.ErrorResponse{
				Type:       invalidJson,
				Message:    jsonErr.Error(),
				StatusCode: http.StatusBadRequest})
		return
	}

	var paymentRepository repository.PaymentRepositoryInterface = repository.PaymentRepository{}
	var paymentHandler PaymentHandlerInterface = PaymentHandler{}

	errorResponse := makePayment(&paymentRequest, paymentRepository, paymentHandler)

	writeResponse(w, paymentRequest.TenantRequestId, errorResponse)
}

func makePayment(paymentRequest *appmodels.PaymentRequest, paymentRepository repository.PaymentRepositoryInterface, paymentHandler PaymentHandlerInterface) *commonAppModels.ErrorResponse {
	var errorResponse *commonAppModels.ErrorResponse = nil

	populateEnums(paymentRequest)

	errorResponse = validateMakePaymentRequest(*paymentRequest)
	if errorResponse != nil {
		return errorResponse
	}

	errorResponse = validatePaymentAmountAndDate(*paymentRequest)
	if errorResponse != nil {
		return errorResponse
	}

	incomingPaymentRequest := getPaymentModel(paymentRequest)

	payments, err := paymentRepository.MakePayment(&incomingPaymentRequest)

	if err != nil {
		var found bool
		for statusCode, errorMessage := range errorMessageAndStatusCodeMapping {
			if strings.EqualFold(err.Error(), errorMessage) {
				errorResponse = &commonAppModels.ErrorResponse{
					Type:       invalidField,
					Message:    err.Error(),
					StatusCode: statusCode}
				found = true
				break
			}
		}
		if !found {
			errorResponse = &commonAppModels.ErrorResponse{
				Type:       internalServerError,
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError}
		}
	} else {
		paymentHandler.PublishPaymentBalancingEventMessage(payments, incomingPaymentRequest)
	}

	return errorResponse
}

func getPaymentModel(paymentRequest *appmodels.PaymentRequest) dbmodels.IncomingPaymentRequest {
	return dbmodels.IncomingPaymentRequest{
		TenantId:                      paymentRequest.TenantId,
		TenantRequestId:               paymentRequest.TenantRequestId,
		UserId:                        paymentRequest.UserId,
		AccountId:                     paymentRequest.AccountId,
		ProductIdentifier:             paymentRequest.ProductIdentifier,
		PaymentFrequencyEnum:          paymentRequest.PaymentFrequencyEnum,
		PaymentExtractionSchedule:     paymentRequest.PaymentExtractionSchedule,
		PaymentExtractionScheduleJson: paymentRequest.PaymentExtractionScheduleJson,
		TransactionTypeEnum:           paymentRequest.TransactionTypeEnum,
		CallerAppEnum:                 paymentRequest.CallerAppEnum,
		PaymentRequestTypeEnum:        paymentRequest.PaymentRequestTypeEnum,
		Metadata:                      paymentRequest.Metadata,
	}
}

func validateMakePaymentRequest(paymentRequest appmodels.PaymentRequest) *commonAppModels.ErrorResponse {
	var errorResponse *commonAppModels.ErrorResponse = nil
	tenantIds := make([]int, 0)
	for _, tenantId := range ValidTenantIds {
		if n, err := strconv.Atoi(tenantId); err == nil {
			tenantIds = append(tenantIds, n)
		}
	}
	if paymentRequest.TenantId < 1 {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "tenantId should be a valid positive big int",
			StatusCode: http.StatusBadRequest}
	} else if !validations.CheckIntBelongstoList(int(paymentRequest.TenantId), tenantIds) {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "unknown tenantId, please provide a valid tenantId",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.TenantRequestId < 1 {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "tenantRequestId should be a valid positive big int",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.UserId == "" {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "useId should be a non empty string",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.ProductIdentifier == "" {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "productIdentifier should be a non empty string",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.PaymentFrequencyEnum == enums.NonePaymentFrequency {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "paymentFrequency should be onetime or recurring",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.TransactionTypeEnum == enums.NoneTransactionType {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "transactionType should be payin or payout",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.CallerAppEnum == enums.NoneCallerApp {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "caller app should be atlas or mcp",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.PaymentRequestTypeEnum == enums.NonePaymentRequestType {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "unknown paymentRequestType. Please provide a valid value.",
			StatusCode: http.StatusBadRequest}
	} else if len(paymentRequest.PaymentExtractionSchedule) < 1 {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "empty payment extraction schedule",
			StatusCode: http.StatusBadRequest}
	} else if len(paymentRequest.PaymentExtractionSchedule) > 1 && paymentRequest.PaymentFrequencyEnum == enums.OneTime {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "invalid payment extraction schedule for onetime paymentFrequency",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.PaymentRequestTypeEnum == enums.CustomerChoice && paymentRequest.CallerAppEnum != enums.ATLAS {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "for customerchoice payment request, atlas should be the caller app",
			StatusCode: http.StatusBadRequest}
	} else if paymentRequest.PaymentRequestTypeEnum == enums.InsuranceAutoAuctions && paymentRequest.CallerAppEnum != enums.MCP {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "for insuranceautoauctions payment request, mcp should be the caller app",
			StatusCode: http.StatusBadRequest}
	}

	return errorResponse
}

// Validating if all the payment amounts in the payment request is less(or equal) to the max allowed payment amount and is valid positive integer
func validatePaymentAmountAndDate(paymentRequest appmodels.PaymentRequest) *commonAppModels.ErrorResponse {
	var errorResponse *commonAppModels.ErrorResponse = nil

	currentDate := getCurrentESTDate()

	for _, paymentExtractionSchedule := range paymentRequest.PaymentExtractionSchedule {

		//Payment extraction date earlier than today's date not allowed
		if time.Time(paymentExtractionSchedule.Date).Before(currentDate) {
			errorResponse = &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "payment date cannot be earlier than today's date",
				StatusCode: http.StatusBadRequest}
			break
		} else if paymentExtractionSchedule.Amount <= 0.0 {
			errorResponse = &commonAppModels.ErrorResponse{
				Type:       invalidField,
				Message:    "provide a valid payment amount greater than 0",
				StatusCode: http.StatusBadRequest}
			break
		} else if paymentRequest.PaymentRequestTypeEnum == enums.InsuranceAutoAuctions {
			if paymentExtractionSchedule.Amount > float32(MaxPaymentAmountIaa) {
				errorResponse = &commonAppModels.ErrorResponse{
					Type:       invalidField,
					Message:    fmt.Sprintf("payment amount exceeds the maximum allowed limit of $%d for InsuranceAutoAuctions", MaxPaymentAmountIaa),
					StatusCode: http.StatusBadRequest}
				break
			}
		}
		// TODO: Add the validation for other payment request types
	}

	return errorResponse
}

func writeResponse(w http.ResponseWriter, tenantRequestId int64, errorResponse *commonAppModels.ErrorResponse) {
	var status string = enums.Accepted.String()
	if errorResponse != nil {
		status = enums.Errored.String()
	}
	paymentResponse := appmodels.PaymentResponse{
		TenantRequestId: tenantRequestId,
		Status:          status,
		ErrorDetails:    errorResponse,
	}
	response, err := json.MarshalIndent(paymentResponse, "", "\t")
	if err != nil {
		handlerLog.Error(context.Background(), err, invalidJsonResponseMessage)
	}

	var statusCode int = 0
	if errorResponse != nil {
		statusCode = errorResponse.StatusCode
	} else {
		statusCode = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func populateEnums(paymentRequest *appmodels.PaymentRequest) {
	paymentRequest.PaymentFrequencyEnum = enums.GetPaymentFrequecyEnum(paymentRequest.PaymentFrequency)
	paymentRequest.TransactionTypeEnum = enums.GetTransactionTypeEnum(paymentRequest.TransactionType)
	paymentRequest.PaymentRequestTypeEnum = enums.GetPaymentRequestTypeEnum(paymentRequest.PaymentRequestType)
	paymentRequest.CallerAppEnum = enums.GetCallerAppEnum(paymentRequest.CallerApp)
}

func (PaymentHandler) PublishPaymentBalancingEventMessage(payments []dbmodels.Payment, incomingPaymentRequest dbmodels.IncomingPaymentRequest) error {

	configHandler := commonFunctions.GetConfigHandler()

	paymentEventsTopic := configHandler.GetString("PaymentPlatform.Kafka.Topics.PaymentEvents", "")

	for _, payment := range payments {
		paymentEvent := commonMessagingModels.PaymentEvent{
			Version:            1,
			PaymentId:          payment.PaymentId,
			Amount:             payment.Amount,
			PaymentDate:        payment.PaymentDate,
			PaymentRequestType: incomingPaymentRequest.PaymentRequestTypeEnum,
			EventType:          enums.SavedInAdminstrator,
			EventDateTime:      time.Now(),
			PaymentMethodType:  incomingPaymentRequest.PaymentMethodType,
		}
		paymentEventJson, _ := json.Marshal(paymentEvent)

		err := messaging.KafkaPublish(paymentEventsTopic, string(paymentEventJson))
		if err != nil {
			handlerLog.Error(context.Background(), err, "error publishing payment balancing event for paymentId: %s", payment.PaymentId)
			return err
		}
	}

	return nil
}

// Get the the current date only in EST zone
func getCurrentESTDate() time.Time {
	timeZone, err := time.LoadLocation("America/New_York")

	if err != nil {
		handlerLog.Error(context.Background(), err, "error loading timezone")
	}

	t := time.Now().In(timeZone)
	currentDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return currentDate
}
