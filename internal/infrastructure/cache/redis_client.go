package cache

import (
	"context"
	"fmt"
	"mrs/internal/infrastructure/config"
	applog "mrs/pkg/log"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient defines the interface for Redis operations used by the cache.
// This allows for easier mocking and testing if needed.
// It mirrors a subset of methods from *redis.Client from github.com/redis/go-redis/v8.
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Pipeline() redis.Pipeliner
	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	// Key operations
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	// Hash operations
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	// Ping(ctx context.Context) *redis.StatusCmd // If needed for health checks by cache layer
}

var _ RedisClient = (*redis.Client)(nil)

func NewRedisClient(cfg config.RedisConfig, logger applog.Logger) (RedisClient, error) {
	logger.Info("Initializing Redis client", applog.String("address", cfg.Address), applog.Int("db", cfg.DB))
	opts := &redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,     // no password set
		DB:           cfg.DB,           // use default DB
		PoolSize:     cfg.PoolSize,     // 连接池大小
		MinIdleConns: cfg.MinIdleConns, // 最小空闲连接数
		PoolTimeout:  cfg.PoolTimeout,  // 获取连接的超时时间
		ReadTimeout:  cfg.ReadTimeout,  // 读取超时
		WriteTimeout: cfg.WriteTimeout, // 写入超时
		IdleTimeout:  cfg.IdleTimeout,  // 空闲连接超时，-1 表示不关闭空闲连接
	}
	rdb := redis.NewClient(opts)

	// 使用 Ping 命令测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logger.Error("failed to connect to Redis", applog.Error(err))
		return nil, fmt.Errorf("NewRedisClient: %w", err)
	}
	logger.Info("redis client conneted successfully and ping result is ok",
		applog.String("address", cfg.Address))

	return rdb, nil
}
