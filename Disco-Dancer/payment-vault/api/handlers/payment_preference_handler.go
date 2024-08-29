package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	appmodels "geico.visualstudio.com/Billing/plutus/payment-vault-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository"
	"geico.visualstudio.com/Billing/plutus/validations"
)

const (
	invalidUserId                    = "invalidUserId"
	invalidField                     = "invalidField"
	invalidJson                      = "invalidJson"
	internalServerError              = "internalServerError"
	noDataFound                      = "No Data Found"
	invalidJsonResponseMessage       = "invalid response, unable to encode the json payload"
	paymentPreferenceValidationError = "paymentpreferencevalidationerror"
)

var handlerLog = logging.GetLogger("payment-vault-handler")

type PaymentPreferenceHandler struct {
}

// GetPaymentPrefrenceHandler is the HTTP handler function for getting payment preferences from GEICO Payment Platform.

// @Summary		    Get Payment Preference.
// @Description	    This API returns payment preference from GEICO Payment Platform, for a user, based on product, transactionType and Payment Request Type.
// @Tags			PaymentVault
// @Param			userId	query	string	true	"UserId from identity"
// @Param			transactionType	query	string	true	"TransactionType (eg.PayIn,PayOut)"
// @Param			paymentRequestType	query	string	true	"PaymentRequestType(eg.CustomerChoice,insuranceautoauctions)"
// @Param			productIdentifier	query	string	true	"Product identifier eg.Policy Number"
// @Produce		    application/json
// @Response		default {object} appmodels.PaymentPreferenceResponseList "Status of get payment preference request, errorDetails will be only displayed in case of 4XX or 5XX error"
// @Router			/paymentpreference [get]
func (PaymentPreferenceHandler) GetPaymentPreferenceHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters
	queryParams := r.URL.Query()
	userId := queryParams.Get("userId")
	transactionType := queryParams.Get("transactionType")
	paymentrequestType := queryParams.Get("paymentRequestType")
	productIdentifier := queryParams.Get("productIdentifier")

	var paymentPreferenceRepository repository.PaymentPreferenceRepositoryInterface = repository.PaymentPreferenceRepository{}

	paymentPreference, errorResponse := getPaymentPreference(userId, transactionType, paymentrequestType, productIdentifier, paymentPreferenceRepository)

	writeResponse(w, paymentPreference, errorResponse)
}

// StorePaymentPreference is the HTTP handler function for storing payment preferences of users for GEICO Payment Platform

// @Summary		    Add Payment Preference.
// @Description	    This API adds payment preference for payin and payout, for a user in GEICO Payment Platform.
// @Tags			PaymentVault
// @Accept			json
// @Produce		    json
// @Param			PaymentPreferenceRequest	body	appmodels.PaymentPreferenceRequest	true	"Payment Preference Request body that needs to be added"
// @Response		default						{object} appmodels.StorePaymentPreferenceResponse 	"Status of payment preference request, errorDetails will be only displayed in case of 4XX or 5XX error"
// @Router			/paymentpreference [post]
func (p PaymentPreferenceHandler) StorePaymentPreference(w http.ResponseWriter, r *http.Request) {
	var repository repository.PaymentPreferenceRepositoryInterface = repository.PaymentPreferenceRepository{}
	var request appmodels.PaymentPreferenceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handlerlog.Error(context.Background(), err, errorInvalidJsonRequestMessage)
		errorResponse := createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
		WriteResponse(w, appmodels.CreateStorePaymentPreferenceResponse(nil, errorResponse), http.StatusBadRequest)
		return
	}

	paymentPreferenceId, err := p.storePaymentPreferenceDetails(&request, repository)
	statusCode := http.StatusOK
	if err != nil {
		statusCode = err.StatusCode
	}
	WriteResponse(w, appmodels.CreateStorePaymentPreferenceResponse(paymentPreferenceId, err), statusCode)

}

func getPaymentPreference(
	userId string,
	transactionType string,
	paymentRequestType string,
	productIdentifier string,
	paymentPreferenceRepository repository.PaymentPreferenceRepositoryInterface) (
	[]repository.PaymentPreference,
	*commonAppModels.ErrorResponse) {

	// Validate the input parameters
	var errorResponse *commonAppModels.ErrorResponse = nil
	errorResponse = validateInputParameters(userId, transactionType, paymentRequestType, productIdentifier)

	if errorResponse != nil {
		return nil, errorResponse
	}

	transactionTypeEnum := enums.GetTransactionTypeEnum(transactionType)
	paymentRequestTypeEnum := enums.GetPaymentRequestTypeEnum(paymentRequestType)

	// Get Payment Preferences from the repository
	paymentPreference, err := paymentPreferenceRepository.GetPaymentPreference(userId, productIdentifier, transactionTypeEnum, paymentRequestTypeEnum)

	if err != nil {
		errorResponse = &commonAppModels.ErrorResponse{
			Type:       internalServerError,
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError}
		return nil, errorResponse
	}

	errorResponse = func() *commonAppModels.ErrorResponse {
		var errorResponse *commonAppModels.ErrorResponse = errorResponse
		if len(paymentPreference) == 0 {
			errorResponse = &commonAppModels.ErrorResponse{Type: noDataFound, Message: fmt.Sprintf("Payment Preference not found for UserId: %s.", userId), StatusCode: http.StatusNotFound}
		}
		return errorResponse
	}()

	return paymentPreference, errorResponse
}

// ConvertToPaymentPreferenceResponseModel maps a slice of repository.PaymentPreference to a slice of appmodels.PaymentPreferenceResponse.
// It converts each payment preference in the input slice to the corresponding response model and returns the resulting slice.
func ConvertToPaymentPreferenceResponseModel(paymentPreference []repository.PaymentPreference) []*appmodels.PaymentPreferenceResponse {

	var paymentPreferenceList []*appmodels.PaymentPreferenceResponse

	for _, paymentPref := range paymentPreference {
		paymentPreference := appmodels.PaymentPreferenceResponse{
			PaymentMethodType:      enums.GetPaymentMethodTypeEnum(int(paymentPref.PaymentMethodType)).String(),
			AccountIdentifier:      paymentPref.AccountIdentifier,
			RoutingNumber:          paymentPref.RoutingNumber,
			PaymentExtendedData:    paymentPref.PaymentExtendedData,
			WalletStatus:           paymentPref.WalletStatus,
			PaymentMethodStatus:    enums.GetPaymentMethodStatusEnum(int(paymentPref.PaymentMethodStatus)).String(),
			AccountValidationDate:  paymentPref.AccountValidationDate,
			WalletAccess:           paymentPref.WalletAccess,
			AutoPayPreference:      paymentPref.AutoPayPreference,
			Split:                  paymentPref.Split,
			Last4AccountIdentifier: paymentPref.Last4AccountIdentifier,
		}
		paymentPreferenceList = append(paymentPreferenceList, &paymentPreference)
	}

	return paymentPreferenceList
}

// validateInputParameters validates the input parameters for the payment preference handler.
// It checks if the userId is alphanumeric, if the transactionType is a valid enum value,
// and if the paymentRequestType is a valid enum value. If any of the validations fail,
// it returns an ErrorResponse with the appropriate error message and status code.
func validateInputParameters(
	userId string,
	transactionType string,
	paymentRequestType string,
	productIdentifier string) *commonAppModels.ErrorResponse {

	var validationErrorResponse *commonAppModels.ErrorResponse = nil

	if !validations.IsAlphanumeric(userId) {
		validationErrorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidUserId,
			Message:    "UserId can only contain alphabets and numbers.",
			StatusCode: http.StatusBadRequest}
	} else if enums.GetTransactionTypeEnum(transactionType) == enums.NoneTransactionType {
		validationErrorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "transactionType should be a valid enum value",
			StatusCode: http.StatusBadRequest}
	} else if enums.GetPaymentRequestTypeEnum(paymentRequestType) == enums.NonePaymentRequestType {
		validationErrorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "paymentRequestType should be a valid enum value",
			StatusCode: http.StatusBadRequest}
	} else if !(len(strings.TrimSpace(productIdentifier)) == 0) && !validations.IsAlphanumeric(productIdentifier) {
		validationErrorResponse = &commonAppModels.ErrorResponse{
			Type:       invalidField,
			Message:    "ProductIdentifier can only contain alphabets and numbers.",
			StatusCode: http.StatusBadRequest}
	}

	return validationErrorResponse
}

// writeResponse writes the HTTP response with the provided payment preferences and error response.
// It sets the appropriate status code based on the presence of an error response.
// The response is encoded as JSON and sent in the HTTP response writer.
// TODO - Use the basehandler method when its ready
func writeResponse(w http.ResponseWriter, paymentPreference []repository.PaymentPreference, errorResponse *commonAppModels.ErrorResponse) {
	var status string = enums.Accepted.String()

	if errorResponse != nil {
		status = enums.Errored.String()
	}

	responseModel := ConvertToPaymentPreferenceResponseModel(paymentPreference)

	paymentResponse := appmodels.PaymentPreferenceResponseList{
		PaymentPreferences: responseModel,
		Status:             status,
		ErrorDetails:       errorResponse,
	}

	var statusCode int
	if errorResponse != nil {
		statusCode = errorResponse.StatusCode
	} else {
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(paymentResponse)

	if err != nil {
		handlerLog.Error(context.Background(), err, invalidJson)
		return
	}
}

func (p PaymentPreferenceHandler) storePaymentPreferenceDetails(request *appmodels.PaymentPreferenceRequest, repository repository.PaymentPreferenceRepositoryInterface) (*int64, *commonAppModels.ErrorResponse) {

	if err := p.validatePaymentPreferenceRequest(request); err != nil {
		return nil, err
	}

	model := p.convertToPaymentPreferenceDbModel(request)

	err := repository.StorePaymentPreference(model)
	if err != nil {
		return nil, p.createErrorResponse(err.Error())
	}

	return &model.PreferenceId, nil

}

func (p PaymentPreferenceHandler) convertToPaymentPreferenceDbModel(request *appmodels.PaymentPreferenceRequest) *db.StorePaymentPreference {
	return &db.StorePaymentPreference{
		ProductDetail: db.ProductDetail{
			UserId:             request.UserId,
			ProductType:        request.ProductType,
			ProductSubType:     request.ProductSubType,
			PaymentRequestType: request.PaymentRequestType,
			ProductIdentifier:  request.ProductIdentifier,
		},
		PayIn:  p.convertToPaymentDetailsDbModel(request.PayIn),
		PayOut: p.convertToPaymentDetailsDbModel(request.PayOut),
	}
}

func (PaymentPreferenceHandler) convertToPaymentDetailsDbModel(details []*appmodels.PaymentPreferenceDetail) []*db.PaymentPreferenceExtendedData {
	if details == nil {
		return nil
	}
	dbExtendedData := make([]*db.PaymentPreferenceExtendedData, len(details))
	for i, data := range details {
		dbExtendedData[i] = &db.PaymentPreferenceExtendedData{
			PaymentMethodId:   data.PaymentMethodId,
			Split:             data.Split,
			AutoPayPreference: data.AutoPayPreference,
		}
	}
	return dbExtendedData
}

func (p PaymentPreferenceHandler) validatePaymentPreferenceRequest(request *appmodels.PaymentPreferenceRequest) *commonAppModels.ErrorResponse {
	validationsList := []validation{
		{request.UserId, "UserId", validations.IsValidAlphanumeric},
		{request.ProductIdentifier, "ProductIdentifier", validations.IsValidProductIdentifier},
	}

	if err := ValidateFields(validationsList, paymentPreferenceValidationError); err != nil {
		return err
	}
	if err := p.validateEnumsFields(request); err != nil {
		return err
	}

	if request.PayIn == nil && request.PayOut == nil {
		return createValidationErrorResponse("either PayIn or PayOut should have a value", paymentPreferenceValidationError, http.StatusBadRequest)
	}

	if request.PayIn != nil {
		if err := p.validatePaymentPreferenceDetail(request.PayIn, "PayIn"); err != nil {
			return createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
		}

	}
	if request.PayOut != nil {
		if err := p.validatePaymentPreferenceDetail(request.PayOut, "PayOut"); err != nil {
			return createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
		}
	}

	return nil
}

func (p PaymentPreferenceHandler) validateEnumsFields(request *appmodels.PaymentPreferenceRequest) *commonAppModels.ErrorResponse {
	if err := request.ProductType.IsValidProductType(); err != nil {
		return createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
	}

	if err := request.ProductSubType.IsValidProductSubType(); err != nil {
		return createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
	}

	if err := request.PaymentRequestType.IsValidPaymentRequestType(); err != nil {
		return createValidationErrorResponse(err.Error(), paymentPreferenceValidationError, http.StatusBadRequest)
	}

	return nil
}

func (p PaymentPreferenceHandler) validatePaymentPreferenceDetail(request []*appmodels.PaymentPreferenceDetail, transactionType string) error {
	var percentage []int16
	seen := make(map[int64]bool)

	for _, data := range request {
		err := validations.IsGreaterThanZero(data.PaymentMethodId, "PaymentMethodId")
		if err != nil {
			return err
		}

		// Payment Method Id should not be repeated in PayIn or PayOut
		if seen[data.PaymentMethodId] {
			return fmt.Errorf("PaymentMethodId %d is repeated in %s", data.PaymentMethodId, transactionType)
		}

		seen[data.PaymentMethodId] = true

		if data.Split != 0 {
			percentage = append(percentage, data.Split)
		}

	}

	// Split percentage should be 100
	var total int16
	for _, value := range percentage {
		total += value
	}

	if total != 100 {
		return fmt.Errorf("split total percentage is not 100 in %s, got %d", transactionType, total)
	}

	return nil
}

func (p PaymentPreferenceHandler) createErrorResponse(errorMessage string) *commonAppModels.ErrorResponse {

	var statusCode int
	var typeName string

	switch {
	case strings.Contains(errorMessage, "not found"):
		statusCode = http.StatusNotFound
		typeName = "PaymentMethodNotFound"
	case strings.Contains(errorMessage, "is not active"):
		statusCode = http.StatusBadRequest
		typeName = "PaymentMethodNotActive"
	case errorMessage == "no payment method details found for given user":
		statusCode = http.StatusBadRequest
		typeName = "PaymentMethodNotFound"
	default:
		statusCode = http.StatusInternalServerError
		typeName = InternalServerError
	}

	return &commonAppModels.ErrorResponse{
		Type:       typeName,
		Message:    errorMessage,
		StatusCode: statusCode,
	}
}
