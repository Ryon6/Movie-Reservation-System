package lock

import "errors"

var (
	ErrLockAlreadyAcquired = errors.New("lock already acquired")
	ErrRetryLockFailed     = errors.New("retry to acquire lock failed")
)
