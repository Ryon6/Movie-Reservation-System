package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initConfig() {
	viper.SetConfigName("config") // 配置文件名 (不带扩展)
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 查找配置文件的路径

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func initLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logger
}

var db *gorm.DB
var rdb *redis.Client

func initDB() {
	dsn := viper.GetString("database.dsn")
	if dsn == "" {
		log.Fatalf("Database DSN is not set in config file")
	}

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func main() {
	// 初始化配置
	initConfig()

	// 初始化日志
	logger := initLogger()
	defer logger.Sync() // 确保所有日志都已刷新到磁盘

	// 初始化数据库
	initDB()

	// 初始化 Redis
	initRedis()

	// 获取服务端口
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080" // 默认端口
	}

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
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
