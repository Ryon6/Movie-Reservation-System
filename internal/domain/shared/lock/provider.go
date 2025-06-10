package lock

import (
	"context"
	"time"
)

// LockProvider 定义了锁的获取和释放接口
type LockProvider interface {
	// 返回一个代表本次加锁的唯一值的字符串（用于安全释放）
	Acquire(ctx context.Context, lockKey string, ttl time.Duration) (Lock, error)
}
