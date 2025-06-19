package cache

import (
	"context"
	"fmt"
	"mrs/internal/domain/shared/lock"
	applog "mrs/pkg/log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// redisLockProvider 实现了 lock.LockProvider 接口
type redisLockProvider struct {
	client *redis.Client
	logger applog.Logger
}

// NewRedisLockProvider 创建一个 Redis 锁提供者
func NewRedisLockProvider(redisClient *redis.Client, logger applog.Logger) lock.LockProvider {
	return &redisLockProvider{client: redisClient, logger: logger.With(applog.String("Component", "redisLockProvider"))}
}

// Acquire 获取一个锁
func (p *redisLockProvider) Acquire(ctx context.Context, lockKey string, ttl time.Duration) (lock.Lock, error) {
	logger := p.logger.With(applog.String("Method", "Acquire"), applog.String("lock_key", lockKey))
	lockValue := uuid.New().String()

	// 原子操作，如果锁不存在，则设置锁并返回成功
	success, err := p.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		logger.Error("redis setnx failed", applog.Error(err))
		return nil, fmt.Errorf("redis setnx failed: %w", err)
	}

	// 如果锁存在，则返回失败
	if !success {
		logger.Error("lock provider error", applog.String("lock_key", lockKey))
		return nil, fmt.Errorf("lock provider error: %w", lock.ErrLockAlreadyAcquired)
	}

	renewCtx, cancel := context.WithCancel(ctx)
	// 自动续约
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				luaScript := `
				if redis.call("get", KEYS[1]) == ARGV[1] then
					return redis.call("expire", KEYS[1], ARGV[2])
				else
					return 0
				end
				`
				result, err := p.client.Eval(renewCtx, luaScript, []string{lockKey}, lockValue, ttl).Result()
				if err != nil || result.(int64) == 0 {
					logger.Error("failed to auto renew lock", applog.Error(err))
					return
				}
				logger.Info("lock auto renewed")
			case <-renewCtx.Done():
				logger.Info("lock auto renew stopped")
				return
			}
		}
	}()

	logger.Info("lock acquired", applog.String("lock_key", lockKey))
	return &redisLock{lockKey: lockKey, lockValue: lockValue, client: p.client, cancel: cancel, logger: logger}, nil
}
