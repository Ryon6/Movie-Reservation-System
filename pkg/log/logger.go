// **日志模块 (`pkg/log` 及应用)**:
// *   在 `pkg/log` 中封装结构化日志库Zap。
// *   在 `cmd/server/main.go` 和其他关键初始化流程中集成日志记录。

// 日志模块功能设计
// **`pkg/log/logger.go`**:
// *   提供一个日志接口 (e.g., `Logger`) 和一个或多个具体实现 (e.g., 基于 Zap 的 `ZapLogger`)。
// *   使得应用代码可以解耦具体的日志库。

package log

// Logger 定义了应用程序中所有日志记录器应遵循的接口。
type Logger interface {
	// Debug 记录调试级别的日志。
	// 通常包含详细的诊断信息，主要用于开发和故障排除。
	Debug(msg string, fields ...Field)

	// Info 记录信息级别的日志。
	// 用于记录常规操作事件，例如服务的启动、重要流程的执行等。
	Info(msg string, fields ...Field)

	// Warn 记录警告级别的日志。
	// 表示可能出现问题或非预期情况，但应用仍能继续运行。
	Warn(msg string, fields ...Field)

	// Error 记录错误级别的日志。
	// 表示发生了错误，某些功能可能无法正常工作，但应用本身不会因此终止。
	Error(msg string, fields ...Field)

	// Fatal 记录致命错误级别的日志，并随后调用 os.Exit(1)。
	// 表示发生了严重错误，应用无法继续运行。
	Fatal(msg string, fields ...Field)

	// Panic 记录紧急错误级别的日志，并随后调用 panic()。
	// 通常用于表示不可恢复的错误或编程错误。
	Panic(msg string, fields ...Field)

	// With 返回一个新的 Logger 实例，该实例会将其所有条目附加上下文字段。
	// 例如 logger.With(log.String("request_id", "123"))
	With(fields ...Field) Logger

	// Sync 确保所有缓冲的日志都被写入。
	// 这通常在应用程序关闭时调用，以确保所有日志都被正确记录。
	Sync() error
}

// Field 定义日志字段，通常是一个键值对
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions to create Fields easily
func String(key string, value string) Field   { return Field{Key: key, Value: value} }
func Int(key string, value int) Field         { return Field{Key: key, Value: value} }
func Int64(key string, value int64) Field     { return Field{Key: key, Value: value} }
func Float64(key string, value float64) Field { return Field{Key: key, Value: value} }
func Bool(key string, value bool) Field       { return Field{Key: key, Value: value} }
func Error(err error) Field                   { return Field{Key: "error", Value: err.Error()} } // Common field for errors
func Any(key string, value interface{}) Field { return Field{Key: key, Value: value} }
