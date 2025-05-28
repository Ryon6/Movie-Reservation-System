// **配置模块 (`internal/infrastructure/config`)**:
// *   实现 `loader.go` (使用 Viper) 从 `configs/app.yaml` 加载配置。
// *   定义 `model.go` 存放配置结构体 (服务端口, 数据库连接字符串, Redis 地址, JWT 密钥等)。

package config

import (
	"log"
	"strconv"

	"github.com/spf13/viper"
)

func LoadConfig() (*Config, error) {
	var config Config
	viper.SetConfigName("app") // 配置文件名
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Load Config error: %v", err)
	}

	db := Database{
		ConnectionString:   viper.GetString("database.dsn"),
		MaxOpenConnections: viper.GetInt("database.maxOpenConnections"),
		MaxIdleConnections: viper.GetInt("database.maxIdleConnections"),
	}

	redis := Redis{
		Address:  viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}

	jwt := JWT{
		SecretKey:            viper.GetString("jwt.secretKey"),
		AccessTokenDuration:  viper.GetDuration("jwt.accessTokenDuration"),
		RefreshTokenDuration: viper.GetDuration("jwt.refreshTokenDuration"),
	}

	config.Database = db
	config.Redis = redis
	config.JWT = jwt
	if port := viper.GetString("server.port"); port != "" {
		config.ServerPort, _ = strconv.Atoi(port)
	} else {
		config.ServerPort = 8080
	}

	return &config, nil
}
