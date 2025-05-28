package config

import (
	"time"
)

type Config struct {
	ServerPort int
	Database
	Redis
	JWT
	Log
}

type Log struct {
	Level            string   `mapstructure:"level"`
	OutputPaths      []string `mapstructure:"outputPaths"`
	ErrorOutputPaths []string `mapstructure:"errorOutputPaths"`
	Encoding         string   `mapstructure:"encoding"`
}

type Database struct {
	ConnectionString   string `mapstructure:"dsn"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
}

type Redis struct {
	Address  string `mapstructure:"addr" yaml:"addr"`         // ip:port
	Password string `mapstructure:"password" yaml:"password"` // 默认不需要
	DB       int    `mapstructure:"db" yaml:"db"`             // 数据库编号
}

type JWT struct {
	SecretKey            string        `mapstructure:"secretKey" yaml:"secretKey"`
	AccessTokenDuration  time.Duration `mapstructure:"accessTokenDuration" yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `mapstructure:"refreshTokenDuration" yaml:"refreshTokenDuration"`
}
