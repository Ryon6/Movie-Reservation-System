package log

import (
	"context"
	"mrs/internal/infrastructure/config"
	"sync"
	"testing"
	"time"
)

// mockLogger 是一个用于测试的 Logger 实现
type mockLogger struct {
	mu     sync.Mutex
	logs   []mockLog
	fields []Field
}

type mockLog struct {
	level   string
	message string
	fields  []Field
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		logs: make([]mockLog, 0),
	}
}

func (m *mockLogger) Debug(msg string, fields ...Field) { m.log("debug", msg, fields) }
func (m *mockLogger) Info(msg string, fields ...Field)  { m.log("info", msg, fields) }
func (m *mockLogger) Warn(msg string, fields ...Field)  { m.log("warn", msg, fields) }
func (m *mockLogger) Error(msg string, fields ...Field) { m.log("error", msg, fields) }
func (m *mockLogger) Panic(msg string, fields ...Field) { m.log("panic", msg, fields) }
func (m *mockLogger) Fatal(msg string, fields ...Field) { m.log("fatal", msg, fields) }
func (m *mockLogger) With(fields ...Field) Logger {
	m.fields = append(m.fields, fields...)
	return m
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) log(level, msg string, fields []Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	allFields := append(m.fields, fields...)
	m.logs = append(m.logs, mockLog{
		level:   level,
		message: msg,
		fields:  allFields,
	})
}

func (m *mockLogger) getLogs() []mockLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs
}

func TestGormLoggerAdapter_SlowQuery(t *testing.T) {
	// 准备测试数据
	mockLog := newMockLogger()
	cfg := config.LogConfig{
		Level:         "info",
		SlowThreshold: 100 * time.Millisecond,
	}

	adapter := NewGormLoggerAdapter(mockLog, cfg)

	// 模拟一个慢查询
	ctx := context.Background()
	begin := time.Now().Add(-200 * time.Millisecond) // 模拟查询耗时200ms
	fc := func() (string, int64) {
		return "SELECT * FROM users", 10
	}

	// 执行Trace方法
	adapter.Trace(ctx, begin, fc, nil)

	// 验证日志输出
	logs := mockLog.getLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.level != "warn" {
		t.Errorf("Expected warn level for slow query, got %s", log.level)
	}
	if log.message != "gorm slow query" {
		t.Errorf("Expected 'gorm slow query' message, got %s", log.message)
	}

	// 验证字段
	expectedFields := map[string]bool{
		"elapsed": false,
		"rows":    false,
		"sql":     false,
	}

	for _, field := range log.fields {
		delete(expectedFields, field.Key)
	}

	for field := range expectedFields {
		t.Errorf("Missing expected field: %s", field)
	}
}

func TestGormLoggerAdapter_NormalQuery(t *testing.T) {
	// 准备测试数据
	mockLog := newMockLogger()
	cfg := config.LogConfig{
		Level:         "info",
		SlowThreshold: 100 * time.Millisecond,
	}

	adapter := NewGormLoggerAdapter(mockLog, cfg)

	// 模拟一个正常查询
	ctx := context.Background()
	begin := time.Now().Add(-50 * time.Millisecond) // 模拟查询耗时50ms
	fc := func() (string, int64) {
		return "SELECT * FROM users", 10
	}

	// 执行Trace方法
	adapter.Trace(ctx, begin, fc, nil)

	// 验证日志输出
	logs := mockLog.getLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.level != "info" {
		t.Errorf("Expected info level for normal query, got %s", log.level)
	}

	// 验证是否包含 module=gorm 字段
	hasModuleField := false
	for _, field := range log.fields {
		if field.Key == "module" && field.Value == "gorm" {
			hasModuleField = true
			break
		}
	}

	if !hasModuleField {
		t.Error("Missing module=gorm field in normal query log")
	}
}
