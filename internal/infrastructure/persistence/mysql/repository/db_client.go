package repository

import (
	"fmt"
	"mrs/internal/infrastructure/config"
	applog "mrs/pkg/log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func CreateDBConnection(dbConfig config.DatabaseConfig,
	logConfig config.LogConfig,
	logger applog.Logger) (*gorm.DB, func(), error) {
	logger.Info("Initializing MySQL database connection",
		applog.String("host", dbConfig.Host),
		applog.String("port", dbConfig.Port),
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: applog.NewGormLoggerAdapter(logger, logConfig),
	})

	if err != nil {
		return nil, nil, fmt.Errorf(("failed to connect to MySQL: %w"), err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if dbConfig.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConnections)
	}
	if dbConfig.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConnections)
	}
	if dbConfig.ConnMaxLifetimeMinutes > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetimeMinutes) * time.Minute)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close() // 确保在失败时关闭连接
		return nil, nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	cleanup := func() {
		sqlDB.Close()
	}

	logger.Info("MySQL database connection established successfully", applog.String("dsn", dsn))
	return db, cleanup, nil
}
