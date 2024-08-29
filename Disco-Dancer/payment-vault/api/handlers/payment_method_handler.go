package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/repository"
	"geico.visualstudio.com/Billing/plutus/validations"
)

var handlerlog = logging.GetLogger("payment-vault-handler")

type validationFunction func(string, string) error

type validation struct {
	value        string
	name         string
	validateFunc validationFunction
}

const (
	errorInvalidJsonRequestMessage  = "invalid request, unable to decode the json payload"
	errorInvalidFieldMessage        = "invalid request, json field value is invalid or missing"
	errorInternalServerErrorMessage = "internal server error, please contact payment platform dev team"
	PaymentMethodValidationError    = "PaymentMethodValidationError"
	InternalServerError             = "InternalServerError"
	PaymentMethodExistError         = "PaymentMethodExistError"
	UserIdNotFoundError             = "UserIdNotFoundError"
)

type PaymentMethodHandler struct {
}

// StorePaymenMethod is the HTTP handler function for storing payment methods for users.

// @Summary		    Add Payment method.
// @Description	    This API adds a payment methods, with account related details, for a user, in GEICO Payment Platform.
// @Tags			PaymentVault
// @Accept			json
// @Produce		    json
// @Param			PaymentMethodRequest	body	app.PaymentMethodRequest	true	"Payment Preference method request body that needs to be added"
// @Response		default						{object} app.PaymentMethodResponse 		"Status of payment method request, errorDetails will be only displayed in case of 4XX or 5XX error"
// @Router			/paymentmethod [post]
func (p PaymentMethodHandler) StorePaymentMethod(w http.ResponseWriter, r *http.Request) {
	var request app.PaymentMethodRequest
	repository := repository.PaymentMethodRepository{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handlerlog.Error(context.Background(), err, errorInvalidJsonRequestMessage)
		errorResponse := createValidationErrorResponse(err.Error(), PaymentMethodValidationError, http.StatusBadRequest)
		WriteResponse(w, app.CreateStorePaymentMethodResponse(nil, errorResponse), http.StatusBadRequest)
		return
	}

	paymentMethodId, err := storePaymentMethodDetails(&request, &repository)
	statusCode := http.StatusOK
	if err != nil {
		statusCode = err.StatusCode
	}
	WriteResponse(w, app.CreateStorePaymentMethodResponse(paymentMethodId, err), statusCode)
}

func validatePaymentMethodRequest(request *app.PaymentMethodRequest) *commonAppModels.ErrorResponse {
	fields := map[string]string{
		"UserID": request.UserID,
	}

	if err := validations.CheckEmptyFields(fields); err != nil {
		return createValidationErrorResponse(err.Error(), PaymentMethodValidationError, http.StatusBadRequest)
	}

	if request.CallerApp == enums.NoneCallerApp {
		return createValidationErrorResponse("callerapp should be valid", PaymentMethodValidationError, http.StatusBadRequest)
	}

	var err *commonAppModels.ErrorResponse

	switch {
	case request.IsCardPaymentMethodType():
		err = validateCardDetails(request)
	case request.IsACHPaymentMethodType():
		err = validateACHDetails(request)
	default:
		err = createValidationErrorResponse("paymentmethodtype should be either ach or card", PaymentMethodValidationError, http.StatusBadRequest)
	}

	return err
}

func storePaymentMethodDetails(request *app.PaymentMethodRequest, repository repository.PaymentMethodRepositoryInterface) (*int64, *commonAppModels.ErrorResponse) {

	if err := validatePaymentMethodRequest(request); err != nil {
		return nil, err
	}

	model := convertToPaymentMethodDBModel(request)

	if err := repository.StorePaymentMethod(&model); err != nil {
		return nil, createErrorResponse(err.Error())
	}

	return &model.PaymentMethodId, nil
}

func convertToPaymentMethodDBModel(request *app.PaymentMethodRequest) db.PaymentMethod {
	var paymentMethod db.PaymentMethod

	last4OfAccountNumber := request.PaymentMethodDetails.AccountNumber[len(request.PaymentMethodDetails.AccountNumber)-4:]

	if request.IsACHPaymentMethodType() {

		paymentMethod = db.PaymentMethod{
			UserID:                 request.UserID,
			CallerApp:              request.CallerApp,
			PaymentMethodType:      request.PaymentMethodType,
			NickName:               request.PaymentMethodDetails.NickName,
			AccountIdentifier:      request.PaymentMethodDetails.AccountNumber,
			RoutingNumber:          request.PaymentMethodDetails.RoutingNumber,
			Last4AccountIdentifier: last4OfAccountNumber,
			PaymentExtendedData: db.PaymentExtendedData{

				FirstName: request.FirstName,
				LastName:  request.LastName,
				ACHAccountType: func() enums.ACHAccountType {
					if request.PaymentMethodDetails.ACHAccountType == enums.NoneACHAccountType {
						return enums.Checking
					}
					return request.PaymentMethodDetails.ACHAccountType
				}(),
			},
		}

	} else if request.IsCardPaymentMethodType() {
		paymentMethod = db.PaymentMethod{
			UserID:                 request.UserID,
			CallerApp:              request.CallerApp,
			PaymentMethodType:      request.PaymentMethodType,
			NickName:               request.PaymentMethodDetails.NickName,
			AccountIdentifier:      request.PaymentMethodDetails.CardNumber,
			RoutingNumber:          request.PaymentMethodDetails.RoutingNumber,
			Last4AccountIdentifier: last4OfAccountNumber,
			PaymentExtendedData: db.PaymentExtendedData{

				FirstName:           request.FirstName,
				LastName:            request.LastName,
				ZipCode:             request.PaymentMethodDetails.ZipCode,
				CardExpirationMonth: request.PaymentMethodDetails.CardExpirationMonth,
				CardExpirationYear:  request.PaymentMethodDetails.CardExpirationMonth,
			},
		}
	}
	return paymentMethod
}
func validateFields(validations []validation) *commonAppModels.ErrorResponse {
	for _, v := range validations {
		err := v.validateFunc(v.value, v.name)
		if err != nil {
			return createValidationErrorResponse(err.Error(), PaymentMethodValidationError, http.StatusBadRequest)
		}
	}
	return nil
}

func validateACHDetails(request *app.PaymentMethodRequest) *commonAppModels.ErrorResponse {
	validations := []validation{
		{request.PaymentMethodDetails.AccountNumber, "AccountNumber", validations.IsValidAccountNumber},
		{request.PaymentMethodDetails.RoutingNumber, "RoutingNumber", validations.IsValidRoutingNumber},
	}

	return validateFields(validations)
}

func validateCardDetails(request *app.PaymentMethodRequest) *commonAppModels.ErrorResponse {
	validations := []validation{
		{request.PaymentMethodDetails.CardNumber, "CardNumber", validations.IsValidCardNumber},
		{request.PaymentMethodDetails.CardExpirationMonth, "CardExpirationMonth", validations.IsValidMonth},
		{request.PaymentMethodDetails.CardExpirationYear, "CardExpirationYear", validations.IsValidYear},
		{request.PaymentMethodDetails.CVV, "CVV", validations.IsValidCVV},
		{request.PaymentMethodDetails.ZipCode, "ZipCode", validations.IsValidZIP},
	}

	return validateFields(validations)
}

func createValidationErrorResponse(errorMessage string, typeName string, statusCode int) *commonAppModels.ErrorResponse {

	return &commonAppModels.ErrorResponse{
		Type:       typeName,
		Message:    errorMessage,
		StatusCode: statusCode,
	}
}

func createErrorResponse(errorMessage string) *commonAppModels.ErrorResponse {

	var statusCode int
	var typeName string

	switch errorMessage {

	case "Payment Method Already Exists":
		statusCode = http.StatusBadRequest
		typeName = PaymentMethodExistError
	case "cannot add payment method as payment method validation was declined":
		statusCode = http.StatusBadRequest
		typeName = PaymentMethodValidationError
	case "User ID does not exist":
		statusCode = http.StatusBadRequest
		typeName = UserIdNotFoundError
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
