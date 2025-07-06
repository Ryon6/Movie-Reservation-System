// TODO: 添加CinemaCache expire=0 时的检查

package main

import (
	"fmt"
	"log"
	config "mrs/internal/infrastructure/config"
	"mrs/internal/infrastructure/persistence/decorators/repository_circuitbreaker"
	"os"

	"github.com/spf13/viper"
)

func main() {
	// 确保日志目录存在 (这个逻辑可以保留)
	if err := os.MkdirAll("./var/log", 0755); err != nil {
		log.Fatalf("Failed to ensure log directory: %v", err)
	}

	// 调用 Wire 生成的 Injector 函数
	// 所有依赖注入的细节全部被隐藏
	engine, cleanup, err := InitializeServer(config.ConfigInput{
		Path: "config",
		Name: "app.dev",
		Type: "yaml",
	})
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer cleanup() // cleanup 会负责关闭数据库、Redis等连接

	// 从 Viper 中获取配置 (如果其他地方还需要的话)
	cfg := viper.Get(config.EnvConfig).(*config.Config)
	port := cfg.ServerConfig.Port

	// 配置熔断器
	repository_circuitbreaker.ConfigMovieRepositoryBreakers()
	repository_circuitbreaker.ConfigShowtimeRepositoryBreakers()

	fmt.Println("Starting server on port " + port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
