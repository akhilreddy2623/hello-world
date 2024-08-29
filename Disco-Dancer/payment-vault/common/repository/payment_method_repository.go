package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-vault-common/models/db"
)

const (
	user              = "payment-vault"
	paymentmethod_Add = `INSERT INTO public.payment_method ("UserId", "CallerApp", "PaymentMethodType", "NickName", "AccountIdentifier", "RoutingNumber", "Last4AccountIdentifier",
				 									"PaymentExtendedData", "WalletStatus", "Status", "AccountValidationDate", "WalletAccess",
													"CreatedDate", "CreatedBy", "UpdatedDate", "UpdatedBy") 
						VALUES 
	  						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		 				RETURNING "PaymentMethodId"`

	paymentmethod_IsACHPaymentMethodExist  = `SELECT EXISTS(SELECT 1 FROM public.payment_method WHERE "UserId"=$1 AND "AccountIdentifier"=$2 AND "RoutingNumber"=$3)`
	paymentmethod_IsCardPaymentMethodExist = `SELECT EXISTS(SELECT 1 FROM public.payment_method WHERE "UserId"=$1 AND "AccountIdentifier"=$2)`
	unhandledExceptionOccurred             = "unhandled exception occurred"
	paymentMethodValidationFailed          = "paymentMethod validation failed"
)

type PaymentMethodRepositoryInterface interface {
	StorePaymentMethod(*db.PaymentMethod) error
}

type PaymentMethodRepository struct {
}

func (PaymentMethodRepository) StorePaymentMethod(paymentMethod *db.PaymentMethod) error {

	encryptedAccountIdentifier, errordetails := crypto.EncryptPaymentInfo(paymentMethod.AccountIdentifier)
	if errordetails != nil {
		return errordetails
	}
	paymentMethod.EncryptedAccountIdentifier = *encryptedAccountIdentifier

	// Ensure that Payment Method does not exist in database

	// TODO :  Additional validation required before inserting into database like wallet status active and frad status

	if errordetails := isPaymentMethodExists(paymentMethod); errordetails != nil {
		return errordetails
	}

	// TODO : It will be enabled later once we know about how user details can be  inserted into database
	//if errordetails := IsUserExists(paymentMethod.UserID); errordetails != nil {
	//	return errordetails
	//}

	//
	// paymentMethodValidation := db.PaymentMethodValidation{
	// 	BankAccountNumber:          paymentMethod.AccountIdentifier,
	// 	EncryptedBankAccountNumber: paymentMethod.EncryptedAccountIdentifier,
	// 	RoutingNumber:              paymentMethod.RoutingNumber,
	// 	FirstName:                  paymentMethod.PaymentExtendedData.FirstName,
	// 	LastName:                   paymentMethod.PaymentExtendedData.LastName,
	// 	AccountType:                paymentMethod.PaymentExtendedData.ACHAccountType,
	// 	UserID:                     paymentMethod.UserID,
	// 	CallerApp:                  paymentMethod.CallerApp,
	// 	Amount:                     0.01,
	// }
	// errordetails = PaymentMethodValidation(paymentMethodValidation)
	// if errordetails != nil {
	// 	return errordetails
	// }

	// Store Payment Method in DB
	if errordetails := storePaymentMethodInDB(paymentMethod); errordetails != nil {
		return errordetails
	}

	return nil // Successfully Saved

}

func isPaymentMethodExists(paymentMethod *db.PaymentMethod) error {
	var exists bool
	var query string
	var err error

	switch paymentMethod.PaymentMethodType {
	case enums.Card:
		query = paymentmethod_IsCardPaymentMethodExist
		err = database.NewDbContext().Database.QueryRow(context.Background(), query, paymentMethod.UserID, paymentMethod.EncryptedAccountIdentifier).Scan(&exists)
	case enums.ACH:
		query = paymentmethod_IsACHPaymentMethodExist
		err = database.NewDbContext().Database.QueryRow(context.Background(), query, paymentMethod.UserID, paymentMethod.EncryptedAccountIdentifier, paymentMethod.RoutingNumber).Scan(&exists)
	}

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error in checking existence of payment method")
		return errors.New(unhandledExceptionOccurred)
	}

	if exists {
		return errors.New("payment method already exists")

	}
	return nil

}

func storePaymentMethodInDB(paymentMethod *db.PaymentMethod) error {

	paymentExtendedDataJSON, err := json.Marshal(paymentMethod.PaymentExtendedData)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "Unable to marshal PaymentExtended Data and error occurred while create new payment method")
		return errors.New(unhandledExceptionOccurred)
	}
	currentTimestamp := time.Now()

	dbError := database.NewDbContext().Database.QueryRow(context.Background(), paymentmethod_Add,
		paymentMethod.UserID,
		paymentMethod.CallerApp,
		paymentMethod.PaymentMethodType,
		paymentMethod.NickName,
		paymentMethod.EncryptedAccountIdentifier,
		paymentMethod.RoutingNumber,
		paymentMethod.Last4AccountIdentifier,
		paymentExtendedDataJSON,
		true, // TODO : it will be false when Onetime Payment is made
		enums.Active.EnumIndex(),
		currentTimestamp,
		true,
		currentTimestamp,
		paymentMethod.UserID,
		currentTimestamp,
		paymentMethod.UserID,
	).Scan(&paymentMethod.PaymentMethodId)

	if dbError != nil {
		repositoryLog.Error(context.Background(), dbError, "error in storing  payment method from payment vault")
		return errors.New(unhandledExceptionOccurred)
	}
	return nil
}
