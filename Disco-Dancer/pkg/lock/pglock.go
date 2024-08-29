package pglock

import (
	"context"

	"geico.visualstudio.com/Billing/plutus/database"
)

const pg_lock_query = "select pg_try_advisory_lock($1)"
const pg_unlock_query = "select pg_advisory_unlock($1)"

type PostgresMutexLock struct {
}

type PSqlLockResource struct {
	LockKey ResourceLockId
}

func NewPSqlLockService() (Lock, error) {
	return &PostgresMutexLock{}, nil
}

func (pl *PostgresMutexLock) RunWithLock(ctx context.Context, fn LockFunc, lockKey ResourceLockId) error {

	conn := database.NewDbContext().Database
	//defer conn.Close()

	row := conn.QueryRow(ctx, pg_lock_query, lockKey)

	var success bool

	err := row.Scan(&success)

	if err != nil {
		return err
	}

	if !success {
		return ErrorLockAcquireFailed
	}

	defer func() {
		conn.Exec(ctx, pg_unlock_query, lockKey)
	}()

	err = fn()

	return err
}
