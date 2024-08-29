package pglock

import (
	"context"
	"errors"
)

var ErrorLockAcquireFailed = errors.New("failed to acquire lock on resource")
var ErrorLockNotFound = errors.New("lock not found")

type ResourceLockId int64
type LockFunc func() error

type Lock interface {
	RunWithLock(ctx context.Context, fn LockFunc, lockKey ResourceLockId) error
}
