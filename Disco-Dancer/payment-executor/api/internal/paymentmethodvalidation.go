package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/logging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/external"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	proto "geico.visualstudio.com/Billing/plutus/proto/paymentmethodvalidation"
	"geico.visualstudio.com/Billing/plutus/validations"
)

var log = logging.GetLogger("payment-executor-internal")

func (s *PaymentMethodValidationServer) ValidatePaymentMethod(ctx context.Context, request *proto.PaymentMethodValidationRequest) (*proto.PaymentMethodValidationResponse, error) {
	userId := request.GetUserId()
	amount := request.GetAmount()
	firstName := request.GetFirstName()
	lastName := request.GetLastName()
	accountNumber := request.GetAccountNumber()
	routingNumber := request.GetRoutingNumber()
	accountType := request.GetAccountType()
	paymentMethodValidationRequest := getPaymentMethodValidationRequest(userId, amount, firstName, lastName, accountNumber, routingNumber, accountType)

	var forteApi external.ValidationApiInterface = external.ForteApi{}
	paymentMethodValidationResponse, err := validatePaymentMethod(paymentMethodValidationRequest, forteApi)
	if err != nil {
		return &proto.PaymentMethodValidationResponse{}, err
	}

	validationresponse, err := json.Marshal(paymentMethodValidationResponse.ValidationResponse)
	if err != nil {
		log.Error(context.Background(), err, "error in masrshalling validation response")
		return &proto.PaymentMethodValidationResponse{}, err
	}
	return &proto.PaymentMethodValidationResponse{
		Status:             paymentMethodValidationResponse.Status.String(),
		ValidationResponse: string(validationresponse),
	}, nil
}

func validatePaymentMethod(paymentMethodValidationRequest app.PaymentMethodValidationRequest, forteApi external.ValidationApiInterface) (*app.PaymentMethodValidationResponse, error) {
	err := validatePaymentMethodValidationRequest(paymentMethodValidationRequest)
	if err != nil {
		return nil, err
	}

	forteRequest := getForteRequest(paymentMethodValidationRequest)
	forteResponse, err := forteApi.ForteValidation(forteRequest)
	if err != nil {
		return nil, err
	}

	paymentMethodValidationResponse := getPaymentMethodValidationResponse(*forteResponse)
	return &paymentMethodValidationResponse, nil
}

func validatePaymentMethodValidationRequest(paymentMethodValidationRequest app.PaymentMethodValidationRequest) error {
	if !validations.IsAlphanumeric(paymentMethodValidationRequest.UserId) {
		return errors.New("userId should be alphanumeric")
	} else if paymentMethodValidationRequest.Amount <= 0 {
		return errors.New("amount should be greater than 0")
	} else if !validations.IsNumeric(paymentMethodValidationRequest.AccountNumber) {
		return errors.New("account number should be numeric")
	} else if !validations.IsNumeric(paymentMethodValidationRequest.RoutingNumber) {
		return errors.New("routing number should be numeric")
	} else if paymentMethodValidationRequest.AccountType == enums.NoneACHAccountType {
		return errors.New("account type should be checking or savings")
	}

	return nil
}

func getPaymentMethodValidationRequest(
	userId string,
	amount float32,
	firstName string,
	lastName string,
	accountNumber string,
	routingNumer string,
	accountType string) app.PaymentMethodValidationRequest {

	return app.PaymentMethodValidationRequest{
		UserId:        userId,
		Amount:        amount,
		FirstName:     firstName,
		LastName:      lastName,
		AccountNumber: accountNumber,
		RoutingNumber: routingNumer,
		AccountType:   enums.GetACHAccountTypeEnum(accountType),
	}
}

func getPaymentMethodValidationResponse(
	forteResponse commonAppModels.ForteResponse) app.PaymentMethodValidationResponse {

	validaiotnStatus := enums.DeclinedPaymentMethodValidationStatus

	if forteResponse.Response.Response_Code == "A01" || forteResponse.Response.Response_Code == "U81" ||
		forteResponse.Response.Response_Code == "U82" || forteResponse.Response.Response_Code == "U88" || forteResponse.Response.Response_Code == "U90" {
		validaiotnStatus = enums.ApprovedPaymentMethodValidationStatus
	}

	return app.PaymentMethodValidationResponse{
		Status:             validaiotnStatus,
		ValidationResponse: forteResponse,
	}
}

func getForteRequest(paymentMethodValidationRequest app.PaymentMethodValidationRequest) app.ForteRequest {
	return app.ForteRequest{
		Action:               external.ActionVerify,
		Customer_id:          hash(paymentMethodValidationRequest.UserId),
		Authorization_Amount: paymentMethodValidationRequest.Amount,
		Echeck: app.Echeck{
			Account_Holder: fmt.Sprintf("%s %s", paymentMethodValidationRequest.FirstName, paymentMethodValidationRequest.LastName),
			Account_Number: paymentMethodValidationRequest.AccountNumber,
			Routing_Number: paymentMethodValidationRequest.RoutingNumber,
			Account_Type:   paymentMethodValidationRequest.AccountType.String(),
		},
	}
}

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}
