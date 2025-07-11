// **配置模块 (`internal/infrastructure/config`)**:
// *   实现 `loader.go` (使用 Viper) 从 `configs/app.yaml` 加载配置。
// *   定义 `model.go` 存放配置结构体 (服务端口, 数据库连接字符串, Redis 地址, JWT 密钥等)。

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	EnvConfigPath = "CONFIG_PATH"
	EnvConfigName = "CONFIG_NAME"
	EnvConfigType = "CONFIG_TYPE"

	EnvConfig = "CONFIG"
)

type ConfigInput struct {
	Path string
	Name string
	Type string
}

// LoadConfig 加载配置文件并支持环境变量覆盖
func LoadConfig(input ConfigInput) (*Config, error) {
	var config Config

	// 优先从环境变量获取配置路径、名称、类型
	if env := os.Getenv(EnvConfigPath); env != "" {
		input.Path = env
	}
	if env := os.Getenv(EnvConfigName); env != "" {
		input.Name = env
	}
	if env := os.Getenv(EnvConfigType); env != "" {
		input.Type = env
	}

	viper.SetConfigName(input.Name)
	viper.SetConfigType(input.Type)
	viper.AddConfigPath(input.Path)

	// 支持环境变量自动覆盖（如 DATABASE_USER 覆盖 database.user）
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// 使用 Unmarshal 自动映射配置
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// 设置默认端口
	if config.ServerConfig.Port == "" {
		config.ServerConfig.Port = "8080"
	}

	viper.Set(EnvConfig, &config) // 设置全局配置

	return &config, nil
}
