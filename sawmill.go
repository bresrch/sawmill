// Package sawmill provides an enhanced logging library that improves upon Go's log/slog
// with nested key-value support, dynamic callbacks, flexible output formatting, and color syntax highlighting.
package sawmill

// DefaultLogger is the global default logger instance
var DefaultLogger Logger

func init() {
	DefaultLogger = New(NewTextHandlerWithDefaults())
}

// Trace logs a message at trace level
func Trace(msg string, args ...interface{}) {
	DefaultLogger.Trace(msg, args...)
}

// Debug logs a message at debug level
func Debug(msg string, args ...interface{}) {
	DefaultLogger.Debug(msg, args...)
}

// Info logs a message at info level
func Info(msg string, args ...interface{}) {
	DefaultLogger.Info(msg, args...)
}

// Warn logs a message at warn level
func Warn(msg string, args ...interface{}) {
	DefaultLogger.Warn(msg, args...)
}

// Error logs a message at error level
func Error(msg string, args ...interface{}) {
	DefaultLogger.Error(msg, args...)
}

// Fatal logs a message at fatal level
func Fatal(msg string, args ...interface{}) {
	DefaultLogger.Fatal(msg, args...)
}

// Panic logs a message at panic level and panics
func Panic(msg string, args ...interface{}) {
	DefaultLogger.Panic(msg, args...)
}

// Mark logs a message at mark level for logical separation
func Mark(msg string, args ...interface{}) {
	DefaultLogger.Mark(msg, args...)
}

// WithNested returns a logger with nested attributes
func WithNested(keyPath []string, value interface{}) Logger {
	return DefaultLogger.WithNested(keyPath, value)
}

// WithDot returns a logger with dot notation attributes
func WithDot(dotPath string, value interface{}) Logger {
	return DefaultLogger.WithDot(dotPath, value)
}

// WithGroup returns a logger with a group
func WithGroup(name string) Logger {
	return DefaultLogger.WithGroup(name)
}

// WithCallback returns a logger with a callback
func WithCallback(fn CallbackFunc) Logger {
	return DefaultLogger.WithCallback(fn)
}

// SetDefaultHandler sets the handler for the default logger
func SetDefaultHandler(handler Handler) {
	DefaultLogger.SetHandler(handler)
}
