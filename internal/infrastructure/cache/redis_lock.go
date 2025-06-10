package cache

import (
	"context"
	"fmt"
	applog "mrs/pkg/log"

	"github.com/go-redis/redis/v8"
)

type redisLock struct {
	lockKey   string
	lockValue string
	client    *redis.Client
	cancel    context.CancelFunc
	logger    applog.Logger
}

func (l *redisLock) Key() string {
	return l.lockKey
}

func (l *redisLock) Value() string {
	return l.lockValue
}

func (l *redisLock) Release(ctx context.Context) error {
	logger := l.logger.With(applog.String("Method", "Release"), applog.String("lock_key", l.lockKey))
	// 取消续约
	l.cancel()

	// 判断是否是当前进程持有的锁
	luaScript := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`

	_, err := l.client.Eval(ctx, luaScript, []string{l.lockKey}, []string{l.lockValue}).Result()
	if err != nil {
		logger.Error("redis lock release failed", applog.Error(err))
		return fmt.Errorf("redis lock release failed: %w", err)
	}

	logger.Info("redis lock released")
	return nil
}
