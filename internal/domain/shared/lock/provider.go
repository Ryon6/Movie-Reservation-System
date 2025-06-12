package lock

import (
	"context"
	"time"
)

const (
	DefaultLockTTL    = 10 * time.Second       // 默认锁过期时间
	DefaultBackoff    = 100 * time.Millisecond // 默认退避时间
	DefaultMaxRetries = 3                      // 默认最大重试次数
)

// LockProvider 定义了锁的获取和释放接口
type LockProvider interface {
	// 返回一个代表本次加锁的唯一值的字符串（用于安全释放）
	Acquire(ctx context.Context, lockKey string, ttl time.Duration) (Lock, error)
}
