package config

import (
	"time"
)

type Config struct {
	ServerConfig   `mapstructure:"server"`
	DatabaseConfig `mapstructure:"database"`
	RedisConfig    `mapstructure:"redis"`
	JWTConfig      `mapstructure:"jwt"`
	LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"` // 服务端口
	Host string `mapstructure:"host"` // 服务主机名或IP地址
}

type LogConfig struct {
	Level            string   `mapstructure:"level"`            // 日志级别，例如 "debug", "info", "warn", "error", "fatal"
	OutputPaths      []string `mapstructure:"outputPaths"`      // 日志编码，例如 "json", "console"
	ErrorOutputPaths []string `mapstructure:"errorOutputPaths"` // 日志输出路径，例如 ["stdout", "/var/log/app.log"]
	Encoding         string   `mapstructure:"encoding"`         // 错误日志输出路径，例如 ["stderr"]
	DevelopmentMode  bool     `mapstructure:"development_mode"` // 是否使用开发模式日志配置（会覆盖上面的一些精细配置）
}

type DatabaseConfig struct {
	// ConnectionString   string `mapstructure:"dsn"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	Host               string `mapstructure:"host"`
	Port               string `mapstructure:"port"`
	Name               string `mapstructure:"name"`
	Charset            string `mapstructure:"charset"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
}

type RedisConfig struct {
	Address  string `mapstructure:"addr" yaml:"addr"`         // ip:port
	Password string `mapstructure:"password" yaml:"password"` // 默认不需要
	DB       int    `mapstructure:"db" yaml:"db"`             // 数据库编号
}

type JWTConfig struct {
	SecretKey            string        `mapstructure:"secretKey" yaml:"secretKey"`
	AccessTokenDuration  time.Duration `mapstructure:"accessTokenDuration" yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `mapstructure:"refreshTokenDuration" yaml:"refreshTokenDuration"`
}
