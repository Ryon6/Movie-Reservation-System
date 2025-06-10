package lock

import "context"

type Lock interface {
	Key() string
	Value() string
	Release(ctx context.Context) error
}
