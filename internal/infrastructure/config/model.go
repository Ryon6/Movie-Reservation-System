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
	AuthConfig     `mapstructure:"auth"`
	AdminConfig    `mapstructure:"admin"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"` // 服务端口
	Host string `mapstructure:"host"` // 服务主机名或IP地址
}

type LogConfig struct {
	Level             string   `mapstructure:"level"`             // 日志级别，例如 "debug", "info", "warn", "error", "fatal"
	OutputPaths       []string `mapstructure:"outputPaths"`       // 日志编码，例如 "json", "console"
	ErrorOutputPaths  []string `mapstructure:"errorOutputPaths"`  // 日志输出路径，例如 ["stdout", "/var/log/app.log"]
	Encoding          string   `mapstructure:"encoding"`          // 错误日志输出路径，例如 ["stderr"]
	DevelopmentMode   bool     `mapstructure:"development_mode"`  // 是否使用开发模式日志配置（会覆盖上面的一些精细配置）
	DisableCaller     bool     `mapstructure:"disableCaller"`     // 是否禁用调用者信息
	DisableStacktrace bool     `mapstructure:"disableStacktrace"` // 是否禁用堆栈跟踪
}

type DatabaseConfig struct {
	// ConnectionString   string `mapstructure:"dsn"`
	User                   string `mapstructure:"user"`
	Password               string `mapstructure:"password"`
	Host                   string `mapstructure:"host"`
	Port                   string `mapstructure:"port"`
	Name                   string `mapstructure:"name"`
	Charset                string `mapstructure:"charset"`
	LogLevel               string `mapstructure:"logLevel"`      // 日志级别，例如 "debug", "info", "warn", "error", "fatal"
	SlowThreshold          int    `mapstructure:"slowThreshold"` // 慢查询阈值，单位毫秒
	MaxOpenConnections     int    `mapstructure:"maxOpenConnections"`
	MaxIdleConnections     int    `mapstructure:"maxIdleConnections"`
	ConnMaxLifetimeMinutes int    `mapstructure:"connMaxLifetimeMinutes"` // 连接最大生命周期，单位分钟
}

type RedisConfig struct {
	Address      string        `mapstructure:"address"`      // Redis 服务器地址，例如 "localhost:6379"
	Password     string        `mapstructure:"password"`     // Redis 密码，如果没有则为空字符串
	DB           int           `mapstructure:"db"`           // Redis 数据库编号，默认为 0
	PoolSize     int           `mapstructure:"poolSize"`     // 连接池大小
	MinIdleConns int           `mapstructure:"minIdleConns"` // 最小空闲连接数
	PoolTimeout  time.Duration `mapstructure:"poolTimeout"`  // 从连接池获取连接的超时时间
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`  // 读取超时
	WriteTimeout time.Duration `mapstructure:"writeTimeout"` // 写入超时
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`  // 空闲连接超时时间 5m
}

type AuthConfig struct {
	DefaultRoleName string `mapstructure:"defaultRoleName"`
	HasherCost      int    `mapstructure:"hasherCost"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	SecretKey            string        `mapstructure:"secretKey" yaml:"secretKey"`
	AccessTokenDuration  time.Duration `mapstructure:"accessTokenDuration" yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `mapstructure:"refreshTokenDuration" yaml:"refreshTokenDuration"`
	Issuer               string        `mapstructure:"issuer" yaml:"issuer"`
}
