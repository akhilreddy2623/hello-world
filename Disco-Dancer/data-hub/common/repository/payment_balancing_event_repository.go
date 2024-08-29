package repository

import (
	"context"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/logging"
)

const (
	insertPaymentEventQuery = `INSERT INTO public."payment_events"("PaymentId", "EventDateTime", "EventType")
								VALUES ($1, $2, $3);`

	upsertPaymentQuery = `INSERT INTO public."payments"("PaymentId", "LatestEvent", "PaymentDate", "Amount", "SettlementAmount", "PaymentRequestType", "PaymentMethodType", "LatestEventDateTime")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT ("PaymentId")
		DO UPDATE SET "LatestEvent" = $9, "LatestEventDateTime" = $10, "SettlementAmount" = $11;`
)

var repositoryLog = logging.GetLogger("data-hub-repository")

type PaymentEventRepository struct {
}

func (repo PaymentEventRepository) AddPaymentEvent(paymentEvent *commonMessagingModels.PaymentEvent) error {
	shouldInsert := shouldInsertPaymentEvent(paymentEvent.PaymentId, paymentEvent.EventDateTime)

	if shouldInsert {
		err := upsertPayment(paymentEvent)
		if err != nil {
			return err
		}
	} else {
		repositoryLog.Info(context.Background(), "Stored LatestEvent is greater than or equal to the passed latestEventId %d. Skipping insert for Payment id %d", paymentEvent.EventType.EnumIndex(), paymentEvent.PaymentId)
	}
	err := insertPaymentEvent(paymentEvent)
	return err
}

func upsertPayment(paymentEvent *commonMessagingModels.PaymentEvent) error {

	negativeSettlementAmount := 0.00 - paymentEvent.SettlementAmount // Stored as negative value

	_, dbError := database.NewDbContext().Database.Exec(context.Background(), upsertPaymentQuery,
		paymentEvent.PaymentId,
		paymentEvent.EventType.EnumIndex(),
		paymentEvent.PaymentDate,
		paymentEvent.Amount,
		negativeSettlementAmount,
		paymentEvent.PaymentRequestType.EnumIndex(),
		paymentEvent.PaymentMethodType.EnumIndex(),
		paymentEvent.EventDateTime,
		paymentEvent.EventType.EnumIndex(),
		paymentEvent.EventDateTime,
		negativeSettlementAmount,
	)

	if dbError != nil {
		repositoryLog.Error(context.Background(), dbError, "error in upsert of payment into data-hub")
		return dbError
	}

	return nil
}

func insertPaymentEvent(paymentEvent *commonMessagingModels.PaymentEvent) error {

	_, dbError := database.NewDbContext().Database.Exec(context.Background(), insertPaymentEventQuery,
		paymentEvent.PaymentId,
		paymentEvent.EventDateTime,
		paymentEvent.EventType.EnumIndex(),
	)

	if dbError != nil {
		repositoryLog.Error(context.Background(), dbError, "error in storing payment event into data-hub")
		return dbError
	}

	return nil
}

func shouldInsertPaymentEvent(paymentId int64, eventDateTime time.Time) bool {
	// Set True, if there is issue in fetching the latest event id, still it allows to insert the event
	notExists := true
	getLatestPaymentEvent := `SELECT NOT EXISTS
                             (SELECT 1 
							  FROM public.payments 
							  WHERE "PaymentId" = $1 AND "LatestEventDateTime" > $2  Order by "LatestEventDateTime" desc)`
	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getLatestPaymentEvent,
		paymentId,
		eventDateTime)
	err := row.Scan(&notExists)
	if err != nil {
		repositoryLog.Error(context.Background(), err, "error fetching payment table")
	}

	return notExists
}
