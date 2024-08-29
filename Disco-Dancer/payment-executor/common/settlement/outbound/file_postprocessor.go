package settlement

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	"geico.visualstudio.com/Billing/plutus/messaging"
	"geico.visualstudio.com/Billing/plutus/payment-executor-common/models/db"
	"github.com/lib/pq"
)

var DataHubPaymentEventsTopic string

type FilePostProcessor struct {
}

func (FilePostProcessor) FilePostProcessor(settlementData []db.SettlementPayment) error {

	var consolidatedIds = getConsolidatedIdsToMarkAsInProgress(settlementData)

	publishSentToBankPaymentEventMessage(context.Background(), consolidatedIds)

	status := enums.InProgress.EnumIndex()
	err := updateConsolidatedRequestStatus(consolidatedIds, status)

	return err
}

func createPaymentEventJson(paymentEvent commonMessagingModels.PaymentEvent) (*string, error) {
	paymentEvent.Version = 1
	paymentEvent.EventType = enums.SentToBank
	paymentEvent.EventDateTime = time.Now()
	paymentEvent.SettlementAmount = paymentEvent.Amount

	paymentEventJson, err := json.Marshal(paymentEvent)
	if err != nil {
		log.Error(context.Background(), err, "unable to marshal processPaymentRequest")
		return nil, err
	}

	paymentEventJsonStr := string(paymentEventJson)
	return &paymentEventJsonStr, nil
}

func publishSentToBankPaymentEventMessage(ctx context.Context, consolidatedIds []int64) {
    // Max batch size to fetch from db
	batchCount := 5000
    // Offset to fetch from db, it is like paging
	offset := 0
	// Chunk size to process the consolidated ids, to avoid large query
	chunkSize := 100
    // Loop through the consolidated ids and get the payment details
	// Dont send all the consolidated ids in one go, send in chunks
	// Max batch size to fetch from db is 1000

    for {
        processedRows := false
        for i := 0; i < len(consolidatedIds); i += chunkSize {
            end := i + chunkSize
            if end > len(consolidatedIds) {
                end = len(consolidatedIds)
            }

            rowsProcessed := getPaymentDetailsByConsolidatedId(ctx, consolidatedIds[i:end], batchCount, offset, enums.Consolidated.EnumIndex())
            if rowsProcessed {
                processedRows = true
            }
        }

        if !processedRows {
            break
        }
        offset += batchCount
    }
}


func getPaymentDetailsByConsolidatedId(ctx context.Context, consolidataIds []int64, batchCount int, offset int, status int) (bool) {
	getPaymentRecordsQuery := `
        SELECT 
            "PaymentId",
            "Amount"::numeric::float,
            "PaymentDate",
            "PaymentRequestType",
			"PaymentMethodType"
        FROM public.execution_request
        WHERE 
            "ConsolidatedId" = ANY($1)
            AND "Status" = $2
        ORDER BY 1 ASC
        LIMIT $3 OFFSET $4`

	rows, err := database.NewDbContext().Database.Query(
		context.Background(),
		getPaymentRecordsQuery,
		pq.Array(consolidataIds),
		status,
		batchCount,
		offset)

	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in fetching payments to update status to Complete from execution_request")
		return false
	}

	defer rows.Close()

	processedRows := false
    	 
	for rows.Next() {
		processedRows = true
		paymentEvents := commonMessagingModels.PaymentEvent{}

		err := rows.Scan(&paymentEvents.PaymentId, &paymentEvents.Amount, &paymentEvents.PaymentDate, &paymentEvents.PaymentRequestType, &paymentEvents.PaymentMethodType)
		if err != nil {
			log.Error(ctx, err, "Payment-Executor - error in fetching payments events and unable to post to datahub for paymentId: %d", paymentEvents.PaymentId)
			continue
		}
		processPaymentEvent, err := createPaymentEventJson(paymentEvents)
		if err != nil {
			log.Error(ctx, err, "error before publishing the payment event message to datahub, payment id '%d'", paymentEvents.PaymentId)
			continue
		}

		err = messaging.KafkaPublish(DataHubPaymentEventsTopic, *processPaymentEvent)
		if err != nil {
			log.Error(ctx, err, "error while publishing the payment event message to datahub, payment id '%d'", paymentEvents.PaymentId)
			continue
		}
	}

	return  processedRows
}

func updateConsolidatedRequestStatus(consolidatedIds []int64, status int) error {

	var updateStatusQuery = `UPDATE public.consolidated_request SET "Status" = $%d WHERE "ConsolidatedId" IN (%s)`

	placeholders := make([]string, len(consolidatedIds))
	for i := range consolidatedIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	updateStatusQuery = fmt.Sprintf(updateStatusQuery, len(consolidatedIds)+1, strings.Join(placeholders, ","))
	args := make([]interface{}, len(consolidatedIds)+1)

	for i, id := range consolidatedIds {
		args[i] = id
	}
	args[len(consolidatedIds)] = status

	rows, err := database.NewDbContext().Database.Query(context.Background(), updateStatusQuery, args...)

	if err != nil {
		log.Error(context.Background(), err, "Payment-Executor - error in updating the status of consolidated_request to InProgress after file creation")
		return err
	}

	defer rows.Close()

	return nil
}

func getConsolidatedIdsToMarkAsInProgress(settlementData []db.SettlementPayment) []int64 {
	var consolidatedIds []int64
	for _, settlement := range settlementData {
		consolidatedIds = append(consolidatedIds, settlement.ConsolidatedId)

	}
	return consolidatedIds
}
