// **`cmd/server/main.go` 基础**:
// *   实现基础的 HTTP 服务器启动 (使用 Gin)。
// *   初步的依赖注入逻辑框架。

package main

import (
	"fmt"

	"log"
	"mrs/internal/api"
	"mrs/internal/api/handlers"
	"mrs/internal/api/middleware"
	"mrs/internal/app"
	"mrs/internal/infrastructure/cache"
	config "mrs/internal/infrastructure/config"
	appmysql "mrs/internal/infrastructure/persistence/mysql"
	"mrs/internal/utils"
	applog "mrs/pkg/log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

const (
	DefaultRoleName        = "USER"                            // RoleRepo
	DefaultHasherCost      = 0                                 // Hansher
	DefaultSecretKey       = "Rome will return like lightning" // JWT
	DefaultIssuer          = "Peng"                            //JWT
	DefaultExpirationHours = 1                                 //JWT
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

func initLogger(cfg *config.Config) applog.Logger {
	zapcfg := zap.NewDevelopmentConfig()
	zapcfg.OutputPaths = cfg.LogConfig.OutputPaths
	zapcfg.ErrorOutputPaths = cfg.LogConfig.ErrorOutputPaths
	zapLevel, _ := zapcore.ParseLevel(cfg.LogConfig.Level)
	zapcfg.Level = zap.NewAtomicLevelAt(zapLevel) // 设置日志级别为 Debug
	logger, err := applog.NewZapLogger(zapcfg)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger
}

var db *gorm.DB
var rdb *redis.Client

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
	// logger, err := zap.NewDevelopment()
	defer logger.Sync() // 确保所有日志都已刷新到磁盘

	logger.Debug("Logger initialized successfully")

	// 初始化数据库
	dbFactory := appmysql.NewMysqlDBFactory(logger)
	db, err = dbFactory.CreateDBConnection(cfg.DatabaseConfig)
	if err != nil {
		logger.Fatal("Failed to connect to MySQL", applog.Error(err))
	}

	// 初始化 Redis
	rdb, err = cache.NewRedisClient(cfg.RedisConfig, logger)

	// 获取服务端口
	port := cfg.ServerConfig.Port

	// utiis
	hasher := utils.NewBcryptHasher(DefaultHasherCost) // 采用默认
	jwtManager, err := utils.NewJWTManagerImpl(DefaultSecretKey, DefaultIssuer, DefaultExpirationHours)
	if err != nil {
		logger.Error("failed to create jwtManager")
	}

	// 基础设施层
	userRepo := appmysql.NewGormUserRepository(db, logger)
	roleRepo := appmysql.NewGormRoleRepository(db, logger)

	// 应用层
	userService := app.NewUserService(DefaultRoleName, userRepo, roleRepo, hasher, logger)
	authService := app.NewAuthService(userRepo, hasher, jwtManager, logger)

	// 接口层
	healthHandler := handlers.NewHealthHandler(db, rdb, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)

	r := api.SetupRouter(healthHandler,
		authHandler,
		userHandler,
		middleware.AuthMiddleware(jwtManager, logger))

	// 启动 HTTP 服务器
	logger.Info("Starting server on port " + port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", applog.Field{Key: "error", Value: err.Error()})
	}
}
