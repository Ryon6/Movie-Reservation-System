// **日志模块 (`pkg/log` 及应用)**:
// *   在 `pkg/log` 中封装结构化日志库Zap。
// *   在 `cmd/server/main.go` 和其他关键初始化流程中集成日志记录。

// 日志模块功能设计
// **`pkg/log/logger.go`**:
// *   提供一个日志接口 (e.g., `Logger`) 和一个或多个具体实现 (e.g., 基于 Zap 的 `ZapLogger`)。
// *   使得应用代码可以解耦具体的日志库。

package log

// Logger 定义通用的日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Sync() error
}

// Field 定义日志字段，通常是一个键值对
type Field struct {
	Key   string
	Value interface{}
}
