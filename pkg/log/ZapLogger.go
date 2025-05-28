package log

import (
	"fmt"
	"mrs/internal/infrastructure/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(cfg config.Log) (*ZapLogger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	LoggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorOutputPaths,
		Encoding:         cfg.Encoding,
	}
	logger, err := LoggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return &ZapLogger{
		logger: logger,
	}, nil
}

// ReplaceLogger 替换底层日志记录器（可选）
func (z *ZapLogger) ReplaceLogger(newLogger *zap.Logger) {
	z.logger = newLogger
}

func (z *ZapLogger) convertToZapFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Debug(msg, zapFields...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Info(msg, zapFields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Warn(msg, zapFields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Error(msg, zapFields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	zapFields := l.convertToZapFields(fields...)
	l.logger.Fatal(msg, zapFields...)
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
