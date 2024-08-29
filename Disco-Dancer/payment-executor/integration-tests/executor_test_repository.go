package integrationtest

import (
	"context"

	"geico.visualstudio.com/Billing/plutus/database"
	"github.com/jackc/pgx/v5"
)

func SeedDataToPostgresTables(params ...string) error {

	tx, err := database.NewDbContext().Database.BeginTx(context.Background(), pgx.TxOptions{})

	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(context.Background())

	for _, v := range params {
		_, err = tx.Exec(context.Background(), v)
		if err != nil {
			return err
		}
	}
	// Commit the transaction.
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func GetSettlementIdentifier(consolidatedId int64) (string, error) {
	var SettlementIdentifier string

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		get_settlement_identifier,
		consolidatedId)

	err := row.Scan(&SettlementIdentifier)

	if err != nil {
		return "", err
	}
	return SettlementIdentifier, nil
}

func GetExecutionRequestStatus(paymentId int64) (int32, error) {
	var Status int32

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getPaymentStatus,
		paymentId)

	err := row.Scan(&Status)

	if err != nil {
		return 0, err
	}
	return Status, nil
}

func GetConsolidatedStatus(consolidatedId int64) (int32, error) {
	var Status int32

	row := database.NewDbContext().Database.QueryRow(
		context.Background(),
		getConsolidatedStatus,
		consolidatedId)

	err := row.Scan(&Status)

	if err != nil {
		return 0, err
	}
	return Status, nil
}
