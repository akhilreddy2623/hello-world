package repository

import (
	"context"
	"time"

	"geico.visualstudio.com/Billing/plutus/database"
)

type BaseRepositoryInterface interface {
	PingDatabase() error
}

type Baserepository struct {
}

func (Baserepository) PingDatabase() error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return database.NewDbContext().Database.Ping(ctx)
}
