package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	sugar *zap.SugaredLogger
}

func NewZapLogger(config zap.Config, options ...zap.Option) (Logger, error) {
	baseLogger, err := config.Build(options...)
	if err != nil {
		return nil, err
	}
	return &ZapLogger{sugar: baseLogger.Sugar()}, nil
}

// NewDevelopmentZapLogger 是一个便捷函数，用于创建开发环境的 Zap Logger。
// 它提供了彩色的、人类可读的输出。
func NewDevelopmentZapLogger() (Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	// 你可以在这里进一步定制 cfg，例如修改时间格式
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder         // 彩色级别
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339) // 更标准的时间格式
	return NewZapLogger(cfg)
}

// NewProductionZapLogger 是一个便捷函数，用于创建生产环境的 Zap Logger。
// 它通常输出 JSON 格式，级别为 Info，性能更高。
func NewProductionZapLogger() (Logger, error) {
	cfg := zap.NewProductionConfig()
	// 你可以在这里进一步定制 cfg
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 时间格式
	return NewZapLogger(cfg)
}

// zapFieldsToZap converts our Field type to []zap.Field for zap.Logger
// or prepares for zap.SugaredLogger's key-value pair arguments.
func convertFields(fields ...Field) []interface{} {
	zapArgs := make([]interface{}, 0, len(fields)*2)
	for _, field := range fields {
		zapArgs = append(zapArgs, field.Key, field.Value)
	}
	return zapArgs
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.sugar.Debugw(msg, convertFields(fields...)...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.sugar.Infow(msg, convertFields(fields...)...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.sugar.Warnw(msg, convertFields(fields...)...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.sugar.Errorw(msg, convertFields(fields...)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.sugar.Fatalw(msg, convertFields(fields...)...)
	// Zap's Fatalw already calls os.Exit(1)
}

func (l *ZapLogger) Sync() error {
	// Sync 确保所有缓冲的日志都被写入
	return l.sugar.Sync()
}

func (l *ZapLogger) Panic(msg string, fields ...Field) {
	l.sugar.Panicw(msg, convertFields(fields...)...)
	// Zap's Panicw already calls panic()
}

func (l *ZapLogger) With(fields ...Field) Logger {
	// SugaredLogger.With 返回一个新的 SugaredLogger
	// 我们需要将其包装回我们的 zapLogger 类型
	return &ZapLogger{sugar: l.sugar.With(convertFields(fields...)...)}
}
