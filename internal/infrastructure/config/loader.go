// **配置模块 (`internal/infrastructure/config`)**:
// *   实现 `loader.go` (使用 Viper) 从 `configs/app.yaml` 加载配置。
// *   定义 `model.go` 存放配置结构体 (服务端口, 数据库连接字符串, Redis 地址, JWT 密钥等)。

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string, configName string, configType string) (*Config, error) {
	var config Config

	// 使用调用者提供的配置参数
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		err = fmt.Errorf("error reading config file: %w", err)
		return nil, err
	}

	// 使用 Unmarshal 自动映射配置
	if err := viper.Unmarshal(&config); err != nil {
		err = fmt.Errorf("error unmarshalling config: %w", err)
		return nil, err
	}

	// 设置默认值
	if config.ServerConfig.Port == "" {
		config.ServerConfig.Port = "8080" // 默认端口
	}

	return &config, nil
}
