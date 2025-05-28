// **`cmd/server/main.go` 基础**:
// *   实现基础的 HTTP 服务器启动 (使用 Gin)。
// *   初步的依赖注入逻辑框架。

package main

import (
	"context"
	"fmt"

	"log"
	config "mrs/internal/infrastructure/config"
	mrslog "mrs/pkg/log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 使用已实现的 LoadConfig 函数加载配置
func initConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func ensureLogDirectory() error {
	if err := os.MkdirAll("./var/log", 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	return nil
}

func initLogger(cfg *config.Config) mrslog.Logger {
	logger, err := mrslog.NewZapLogger(cfg.Log)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	return logger
}

var db *gorm.DB
var rdb *redis.Client

func initDB(cfg *config.Config) {
	dsn := cfg.Database.ConnectionString
	if dsn == "" {
		panic("Database DSN is not set in config")
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
}

func initRedis(cfg *config.Config) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
}

func main() {
	// 初始化配置
	cfg, err := initConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	viper.Set("config", cfg)

	// 确保日志目录存在
	if err := ensureLogDirectory(); err != nil {
		if err := ensureLogDirectory(); err != nil {
			log.Fatalf("Failed to ensure log directory: %v", err)
		}
	}

	// 初始化日志
	logger := initLogger(cfg)
	defer logger.Sync() // 确保所有日志都已刷新到磁盘

	logger.Info("Logger initialized successfully")

	// 初始化数据库
	initDB(cfg)

	// 初始化 Redis
	initRedis(cfg)

	// 获取服务端口
	port := fmt.Sprintf("%d", cfg.ServerPort)

	// 创建 Gin 引擎
	r := gin.Default()

	// 健康检查 API
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 启动 HTTP 服务器
	logger.Info("Starting server on port " + port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", mrslog.Field{Key: "error", Value: err.Error()})
	}
}
