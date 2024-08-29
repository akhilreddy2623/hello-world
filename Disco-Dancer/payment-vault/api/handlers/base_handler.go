package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/logging"
)

var log = logging.GetLogger("basehandler")

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type Basehandler struct {
}

// Adding Couple of methods to test encrypt and ecrypt during Deployment
// TODO : Remove temp code

// Encrypt is the HTTP handler function for encrypting a value in cipher text.

// @Summary		    Get encrypted value.
// @Description	    This API encrypts the value passed in the query parameter.
// @Tags			PaymentVault
// @Param			value	query	string	true	"Value to be encrypted"
// @Produce		    application/json
// @Response		default {string} string "Status of the request, errorDetails will be only displayed in case of 4XX or 5XX error"
// @Router			/encrypt [get]
func (Basehandler) Encrypt(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("value")
	log.Info(context.Background(), "Strating Encryption for value %s", value)
	encryptedValue, err := crypto.EncryptPaymentInfo(value)
	var payload jsonResponse

	if err != nil {
		payload = jsonResponse{
			Error:   true,
			Message: err.Error(),
		}
	} else {

		payload = jsonResponse{
			Error:   false,
			Message: *encryptedValue,
		}
	}
	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)
	log.Info(context.Background(), "Completed Encryption for value %s and encrypted value is %s", value, *encryptedValue)
}

// Decrypt is the HTTP handler function for decrypting a cipher text.

// @Summary		    Get decrypted value.
// @Description	    This API decrypts the value passed in the query parameter.
// @Tags			PaymentVault
// @Param			value	query	string	true	"Value to be decrypted"
// @Produce		    application/json
// @Response		default {string} string "Status of the request, errorDetails will be only displayed in case of 4XX or 5XX error"
// @Router			/decrypt [get]
func (Basehandler) Decrypt(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("value")
	log.Info(context.Background(), "Strating Decryption for value %s", value)
	decryptedValue, err := crypto.DecryptPaymentInfo(value)

	var payload jsonResponse

	if err != nil {
		payload = jsonResponse{
			Error:   true,
			Message: err.Error(),
		}
	} else {

		payload = jsonResponse{
			Error:   false,
			Message: *decryptedValue,
		}
	}

	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)
	log.Info(context.Background(), "Completed Decryption for value %s and decrypted value is %s", value, *decryptedValue)
}

func (Basehandler) Workday(w http.ResponseWriter, r *http.Request) {

	requestBytes, _ := io.ReadAll(r.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(requestBytes)
}

func WriteResponse(w http.ResponseWriter, model any, statusCode int) {

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(model)
	if err != nil {
		return
	}

}

func ValidateFields(validations []validation, typeName string) *commonAppModels.ErrorResponse {
	for _, v := range validations {
		err := v.validateFunc(v.value, v.name)
		if err != nil {
			return createValidationErrorResponse(err.Error(), typeName, http.StatusBadRequest)
		}
	}
	return nil
}
