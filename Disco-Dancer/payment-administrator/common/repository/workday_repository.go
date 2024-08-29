package repository

import (
	"context"
	"fmt"
	"strconv"

	commonFunctions "geico.visualstudio.com/Billing/plutus/common-functions"
	commonMessagingModels "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	"geico.visualstudio.com/Billing/plutus/database"
	"geico.visualstudio.com/Billing/plutus/enums"
	dbmodels "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"
	"github.com/jackc/pgx/v5"
)

//go:generate mockery --name WorkdayRepositoryInterface
type WorkdayRepositoryInterface interface {
	UpdateIsSentToWorkdayFalse(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) error
	GetWorkdayFeedRows(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) ([]*dbmodels.WorkdayFeed, error)
	UpdatePaymentWorkdayFeedStatus(workdayFeedRows []*dbmodels.WorkdayFeed)
}

type WorkdayRepository struct {
}

func (WorkdayRepository) GetWorkdayFeedRows(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) ([]*dbmodels.WorkdayFeed, error) {
	var rows pgx.Rows
	var err error
	configHandler := commonFunctions.GetConfigHandler()
	batchCount := configHandler.GetInt("PaymentPlatform.Workday.BatchCount", 0)
	getWorkdayFeed := `SELECT p."PaymentId",
						p."Amount"::numeric::float,
						p."PaymentDate",
						i."Metadata"
					FROM public.payment AS p 
					INNER JOIN public.incoming_payment_request AS i
					ON p."RequestId" = i."RequestId"
					WHERE p."PaymentDate" <= $1 AND p."Status" = $2 AND p."IsSentToWorkday"= 'FALSE'`

	if executeTaskRequest.ExecutionParametersDb.PaymentRequestType != enums.AllPaymentRequestType {
		getWorkdayFeed = getWorkdayFeed + " AND i.\"PaymentRequestType\" = $3"
		getWorkdayFeed = getWorkdayFeed + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getWorkdayFeed,
			executeTaskRequest.TaskDate,
			enums.InProgress.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentRequestType.EnumIndex())
	} else {
		getWorkdayFeed = getWorkdayFeed + fmt.Sprintf(" FETCH FIRST %d rows Only", batchCount)
		rows, err = database.NewDbContext().Database.Query(
			context.Background(),
			getWorkdayFeed,
			executeTaskRequest.TaskDate,
			enums.InProgress.EnumIndex())
	}

	if err != nil {
		repositoryLog.Error(context.Background(), err, "error executing getWorkdayFeeds query")
		return nil, err
	}

	var workdayFeed []*dbmodels.WorkdayFeed
	for rows.Next() {
		workdatFeedItem := dbmodels.WorkdayFeed{}
		err := rows.Scan(
			&workdatFeedItem.PaymentId,
			&workdatFeedItem.Amount,
			&workdatFeedItem.PaymentDate,
			&workdatFeedItem.Metadata)
		if err != nil {
			repositoryLog.Error(context.Background(), err, "error fetching results from getWorkdayFeed query")
			return nil, err
		}

		workdayFeed = append(workdayFeed, &workdatFeedItem)
	}
	return workdayFeed, err
}

func (WorkdayRepository) UpdateIsSentToWorkdayFalse(executeTaskRequest commonMessagingModels.ExecuteTaskRequestDb) error {

	updatePayments := `UPDATE public.payment AS p
	                   SET "IsSentToWorkday" = 'FALSE' 
					   FROM public.incoming_payment_request AS i
					   WHERE p."RequestId" = i."RequestId" AND p."PaymentDate" <= $1 AND p."Status" = $2`
	var err error
	if executeTaskRequest.ExecutionParametersDb.PaymentRequestType != enums.AllPaymentRequestType {
		updatePayments = updatePayments + " AND i.\"PaymentRequestType\" = $3"
		_, err = database.NewDbContext().Database.Exec(
			context.Background(),
			updatePayments,
			executeTaskRequest.TaskDate,
			enums.InProgress.EnumIndex(),
			executeTaskRequest.ExecutionParametersDb.PaymentRequestType.EnumIndex())

	} else {
		_, err = database.NewDbContext().Database.Exec(
			context.Background(),
			updatePayments,
			executeTaskRequest.TaskDate,
			enums.InProgress.EnumIndex())
	}

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Unable to update payment status")
		return err
	}
	return nil
}

func (WorkdayRepository) UpdatePaymentWorkdayFeedStatus(workdayFeedRows []*dbmodels.WorkdayFeed) {
	paymentIds := getpaymentIdsForWordayFeed(workdayFeedRows)
	updatePaymentWorkdayFeedStatusToTrue(paymentIds)
}

func updatePaymentWorkdayFeedStatusToTrue(ids []int64) {

	var paramrefs string

	for i := range ids {
		paramrefs += `$` + strconv.Itoa(i+1) + `,`
	}
	paramrefs = paramrefs[:len(paramrefs)-1]

	query := `UPDATE public.payment SET "IsSentToWorkday" = 'TRUE' WHERE "PaymentId" IN (` + paramrefs + `)`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := database.NewDbContext().Database.Exec(context.Background(), query, args...)

	if err != nil {
		repositoryLog.Error(context.Background(), err, "Unable to update payment status for workdayfeed")

	}
}

func getpaymentIdsForWordayFeed(workdayFeedRows []*dbmodels.WorkdayFeed) []int64 {
	var paymentIds []int64
	for _, item := range workdayFeedRows {
		paymentIds = append(paymentIds, item.PaymentId)
	}
	return paymentIds
}
