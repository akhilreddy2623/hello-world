package settlement

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
)

type DataReader struct {
}

func (DataReader) ReadData(date time.Time, executionParameters db.ExecutionParameters) ([]db.SettlementPayment, error) {

	var settlementPayments []db.SettlementPayment
	var err error
	if executionParameters.PaymentMethodType == enums.Card {
		err = fmt.Errorf("paymentmethodtype %s is not supported for settlement", executionParameters.PaymentMethodType.String())
		return settlementPayments, err
	}
	if executionParameters.PaymentFrequency == enums.Recurrring {
		err = fmt.Errorf("paymentfrequency %s is not supported for settlement", executionParameters.PaymentFrequency.String())
		return settlementPayments, err
	}
	if executionParameters.PaymentRequestType == enums.AllPaymentRequestType || executionParameters.PaymentRequestType == enums.InsuranceAutoAuctions {
		// Only fetch the New Status payments for file creation
		settlementPayments, err = getConsolidatedPayments(date, enums.New.EnumIndex())
	} else {
		err = fmt.Errorf("paymentrequesttype %s is not supported for settlement", executionParameters.PaymentRequestType.String())
	}

	return settlementPayments, err
}

func getConsolidatedPayments(date time.Time, paymentStatus int) ([]db.SettlementPayment, error) {

	var consolidatedPayments []db.SettlementPayment
	getConsolidatedPaymentsQuery :=
		`SELECT 
			"Amount"::numeric::float,
			"AccountIdentifier",
			"RoutingNumber",
			"PaymentExtendedData", 
			"PaymentRequestType", 
			"ConsolidatedId", 
			"PaymentDate"
		FROM public.consolidated_request
		WHERE 
		"PaymentDate" <= $1
		AND "Status" = $2`

	rows, err := database.NewDbContext().Database.Query(
		context.Background(),
		getConsolidatedPaymentsQuery,
		date,
		paymentStatus)

	if err != nil {
		log.Error(context.Background(), err, "error executing getConsolidatedPayments query")
		return nil, err
	}

	for rows.Next() {
		consolidatedPayment := db.SettlementPayment{}
		var paymentExtendedData string

		err := rows.Scan(
			&consolidatedPayment.Amount,
			&consolidatedPayment.AccountIdentifier,
			&consolidatedPayment.RoutingNumber,
			&paymentExtendedData,
			&consolidatedPayment.PaymentRequestType,
			&consolidatedPayment.ConsolidatedId,
			&consolidatedPayment.PaymentDate,
		)
		if err != nil {
			log.Error(context.Background(), err, "error executing getConsolidatedPayments query")
			return nil, err
		}

		err = json.Unmarshal([]byte(paymentExtendedData), &consolidatedPayment.PaymentExtendedData)
		if err != nil {
			log.Error(context.Background(), err, "Error in unmarshalling paymentextendeddata")
			return nil, err
		}

		consolidatedPayments = append(consolidatedPayments, consolidatedPayment)
	}

	return consolidatedPayments, nil
}
