package settlement

import (
	"context"
	"fmt"

	"geico.visualstudio.com/Billing/plutus/crypto"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/app"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
)

type DataMapper struct {
}

func (DataMapper) MapData(settlementPayments []db.SettlementPayment) ([]app.SettlementPayment, error) {

	var mappedPayments []app.SettlementPayment

	for _, settlementPayment := range settlementPayments {
		mappedPayment := app.SettlementPayment{
			AccountName:          getMappedAccountName(settlementPayment.PaymentExtendedData.FirstName, settlementPayment.PaymentExtendedData.LastName),
			Amount:               settlementPayment.Amount,
			AccountIdentifier:    getMappedAccountIdentifier(settlementPayment.AccountIdentifier),
			RoutingNumber:        settlementPayment.RoutingNumber,
			PaymentRequestType:   settlementPayment.PaymentRequestType,
			ConsolidatedId:       settlementPayment.ConsolidatedId,
			PaymentDate:          settlementPayment.PaymentDate,
			SettlementIdentifier: settlementPayment.SettlementIdentifier,
		}
		mappedPayments = append(mappedPayments, mappedPayment)
	}
	return mappedPayments, nil
}

func getMappedAccountIdentifier(accountIdentifier string) string {
	decryptedAccountIdentifier, err := crypto.DecryptPaymentInfo(accountIdentifier)
	if err != nil {
		log.Error(context.Background(), err, "error in descrypting account identifier")
	}
	return *decryptedAccountIdentifier
}

func getMappedAccountName(firstName string, lastName string) string {
	var accountName string
	if firstName == "" {
		accountName = lastName
	} else if lastName == "" {
		accountName = firstName
	} else {
		accountName = fmt.Sprintf("%s %s", firstName, lastName)
	}
	return accountName
}
