package cache

import (
	"context"
	"fmt"
	"mrs/internal/infrastructure/config"
	applog "mrs/pkg/log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClientWrapper struct {
	Client *redis.Client
	logger applog.Logger
}

func NewRedisClient(cfg config.RedisConfig, logger applog.Logger) (*redis.Client, error) {
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
