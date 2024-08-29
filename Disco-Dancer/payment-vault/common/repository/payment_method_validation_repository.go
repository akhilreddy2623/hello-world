package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	commonAppModels "geico.visualstudio.com/Billing/plutus/common-models/app"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/clients"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
)

func PaymentMethodValidation(paymentMethodValidation db.PaymentMethodValidation) error {

	paymentMethodValidationStatus := getPaymentMethodValidationStatus(paymentMethodValidation.EncryptedBankAccountNumber, paymentMethodValidation.RoutingNumber)
	err := paymentMethodValidationCheck(paymentMethodValidationStatus)
	if err != nil && strings.Contains(err.Error(), "validation failed") {

		paymentMethodValidationStatus = performRealTimePaymentMethodValidationAndStoreResultsInDb(paymentMethodValidation)
		return paymentMethodValidationCheck(paymentMethodValidationStatus)
	}
	return err

}

func getPaymentMethodValidationStatus(bankAccountNumber string, routingNumber string) enums.PaymentMethodValidationStatus {

	var paymentMethodValidationStatus int
	getPaymentMethodValidationStatusRequest := `SELECT "ValidationStatus" 
	                                            FROM public.ach_validation_history 
	                      						WHERE "BankAccountNumber" = $1 AND "RoutingNumber"= $2
												  AND (("ValidationStatus" = 2 AND "CreatedDate" > current_date - interval '90' day)   
												  OR ("ValidationStatus" = 1 AND "CreatedDate" > current_date - interval '5' year) )
											    ORDER BY "CreatedDate" DESC
											    FETCH FIRST 1 row only`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentMethodValidationStatusRequest,
		bankAccountNumber,
		routingNumber)
	err := row.Scan(&paymentMethodValidationStatus)
	if err != nil {
		repositoryLog.Info(context.Background(), "error fetching from ach_validation_history table, there may be no entry in the table with given condition")
	}
	return enums.GetPaymentMethodValidationStatusEnumByIndex(paymentMethodValidationStatus)
}

func paymentMethodValidationCheck(paymentMethodValidationStatus enums.PaymentMethodValidationStatus) error {

	if paymentMethodValidationStatus == enums.ApprovedPaymentMethodValidationStatus {
		repositoryLog.Info(context.Background(), "payment method validation approved")
		return nil
	} else if paymentMethodValidationStatus == enums.DeclinedPaymentMethodValidationStatus {
		repositoryLog.Info(context.Background(), "payment method validation declined")
		return errors.New("cannot add payment method as payment method validation was declined")
	}

	repositoryLog.Info(context.Background(), paymentMethodValidationFailed)
	return errors.New(paymentMethodValidationFailed)
}

func performRealTimePaymentMethodValidationAndStoreResultsInDb(paymentMethodValidation db.PaymentMethodValidation) enums.PaymentMethodValidationStatus {
	forteValidationResponse := commonAppModels.ForteResponse{}

	paymentMethodValidationResponse, err := clients.ValidatePaymentMethod(
		paymentMethodValidation.UserID,
		paymentMethodValidation.Amount,
		paymentMethodValidation.FirstName,
		paymentMethodValidation.LastName,
		paymentMethodValidation.BankAccountNumber,
		paymentMethodValidation.RoutingNumber,
		paymentMethodValidation.AccountType.String())

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error calling payment method validation")
		return enums.NonePaymentMethodValidationStatus
	}
	err = json.Unmarshal([]byte(paymentMethodValidationResponse.ValidationResponse), &forteValidationResponse)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error unmarshalling payment method validation response")
		return enums.NonePaymentMethodValidationStatus
	}

	paymentMethodValidationStatus := enums.GetPaymentMethodValidationStatusEnum(paymentMethodValidationResponse.Status)
	rawResponse, err := json.Marshal(forteValidationResponse)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in marshalling payment validation request")
		return enums.NonePaymentMethodValidationStatus
	}
	currentTimestamp := time.Now()

	addPaymentMethodValidation := `INSERT INTO public.ach_validation_history
									("UserId",
									"BankAccountNumber",
									"RoutingNumber", 
									"Amount", 
									"ResponseType", 
									"ResponseCode",
									"ValidationStatus", 
									"RawResponse",
									"ProductIdentifier",
									"CallerApp",
									"CreatedDate",
									"CreatedBy",
									"UpdatedDate",
									"UpdatedBy"
									)
								   VALUES(
									$1, $2, $3,	$4, $5, $6, $7,	$8, $9, $10, $11, $12, $13, $14
									)`

	_, err = database.NewDbContext().Database.Exec(
		context.Background(),
		addPaymentMethodValidation,
		paymentMethodValidation.UserID,
		paymentMethodValidation.EncryptedBankAccountNumber,
		paymentMethodValidation.RoutingNumber,
		paymentMethodValidation.Amount,
		forteValidationResponse.Response.Response_Type,
		forteValidationResponse.Response.Response_Code,
		paymentMethodValidationStatus.EnumIndex(),
		rawResponse,
		nil,
		paymentMethodValidation.CallerApp.EnumIndex(),
		currentTimestamp,
		user,
		currentTimestamp,
		user)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "unable to insert into ach_validation_history table")
	}

	return paymentMethodValidationStatus
}
