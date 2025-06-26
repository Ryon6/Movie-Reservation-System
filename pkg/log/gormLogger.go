package log

import (
	"context"
	"errors"
	"fmt"
	"mrs/internal/infrastructure/config"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// parseLevel 将字符串日志级别转换为 GORM 日志级别
func parseLevel(level string) logger.LogLevel {
	switch strings.ToLower(level) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// GormLoggerAdapter 是一个适配器，它实现了 gorm/logger.Interface，
// 并将日志事件转发到你自定义的 Logger 接口。
type GormLoggerAdapter struct {
	AppLogger     Logger
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

// NewGormLoggerAdapter 创建一个新的适配器实例。
func NewGormLoggerAdapter(appLogger Logger, config config.LogConfig) *GormLoggerAdapter {
	return &GormLoggerAdapter{
		AppLogger:     appLogger,
		LogLevel:      parseLevel(config.Level),
		SlowThreshold: config.SlowThreshold,
	}
}

// LogMode 设置日志级别
func (l *GormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	// 返回一个新的实例，以避免并发问题
	newAdapter := *l
	newAdapter.LogLevel = level
	return &newAdapter
}

// Info 将 GORM 的 Info 日志转发到 AppLogger.Info
func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.AppLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 将 GORM 的 Warn 日志转发到 AppLogger.Warn
func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.AppLogger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 将 GORM 的 Error 日志转发到 AppLogger.Error
func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.AppLogger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 是核心方法，处理SQL日志
func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)

	switch {
	// 情况1: 发生错误 (且不是 "record not found")
	case err != nil && l.LogLevel >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.AppLogger.Error("gorm query error",
			Error(err),
			Duration("elapsed", elapsed),
			Int64("rows", rows),
			String("sql", sql),
		)
	// 情况2: 慢查询
	// 这里的 rows 是 fc() 返回的第二个参数，表示 SQL 语句影响的行数（rows affected）。
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		sql, rows := fc()
		l.AppLogger.Warn("gorm slow query",
			Duration("elapsed", elapsed),
			Int64("rows", rows),
			String("sql", sql),
		)
	// 情况3: 普通SQL日志
	case l.LogLevel >= logger.Info:
		sql, rows := fc()
		// 使用你的 With 方法来添加通用字段
		l.AppLogger.With(String("module", "gorm")).Info("query",
			Duration("elapsed", elapsed),
			Int64("rows", rows),
			String("sql", sql),
		)
	}
}
