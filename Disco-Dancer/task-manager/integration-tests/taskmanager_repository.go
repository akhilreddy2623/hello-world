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
	defer tx.Rollback(context.Background())

	for _, v := range params {
		_, err = tx.Exec(context.Background(), v)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
	return nil
}
