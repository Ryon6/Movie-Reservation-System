package config

import "time"

type Config struct {
	ServerPort int
	Database
	Redis
	JWT
}

type Database struct {
	ConnectionString   string
	MaxOpenConnections int
	MaxIdleConnections int
}

type Redis struct {
	Address  string // ip:port
	Password string // 默认不需要
	DB       int    // 数据库编号
}

type JWT struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}
