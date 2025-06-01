package sawmill

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
)

// logger implements the Logger interface
type logger struct {
	handler   Handler
	attrs     *RecursiveMap
	groups    []string
	callbacks []CallbackFunc
	mu        sync.RWMutex
}

// New creates a new logger with the specified handler
func New(handler Handler) Logger {
	return &logger{
		handler:   handler,
		attrs:     NewRecursiveMap(),
		groups:    make([]string, 0),
		callbacks: make([]CallbackFunc, 0),
	}
}

// Default creates a logger with a default text handler to stdout
func Default() Logger {
	return New(NewTextHandlerWithDefaults())
}

// Log logs a message at the specified level with optional arguments
func (l *logger) Log(ctx context.Context, level Level, msg string, args ...interface{}) {
	if !l.handler.Enabled(ctx, level) {
		return
	}

	record := NewRecord(level, msg)
	record.Context = ctx

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record.PC = pcs[0]

	record.Attributes.Merge(l.attrs)
	l.processArgs(record, args...)

	l.mu.RLock()
	for _, callback := range l.callbacks {
		record = callback(record)
	}
	l.mu.RUnlock()

	l.handler.Handle(ctx, record)
}

// LogRecord logs a pre-constructed record
func (l *logger) LogRecord(ctx context.Context, record *Record) {
	if !l.handler.Enabled(ctx, record.Level) {
		return
	}

	record.Attributes.Merge(l.attrs)

	l.mu.RLock()
	for _, callback := range l.callbacks {
		record = callback(record)
	}
	l.mu.RUnlock()

	l.handler.Handle(ctx, record)
}

func (l *logger) processArgs(record *Record, args ...interface{}) {
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}

		key, ok := args[i].(string)
		if !ok {
			continue
		}

		value := args[i+1]

		keyPath := make([]string, len(l.groups))
		copy(keyPath, l.groups)
		keyPath = append(keyPath, key)

		record.Attributes.Set(keyPath, value)
	}
}

// Trace logs a message at trace level
func (l *logger) Trace(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelTrace, msg, args...)
}

// Debug logs a message at debug level
func (l *logger) Debug(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelDebug, msg, args...)
}

// Info logs a message at info level
func (l *logger) Info(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelInfo, msg, args...)
}

// Warn logs a message at warn level
func (l *logger) Warn(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelWarn, msg, args...)
}

// Error logs a message at error level
func (l *logger) Error(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelError, msg, args...)
}

// Fatal logs a message at fatal level
func (l *logger) Fatal(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelFatal, msg, args...)
}

// Panic logs a message at panic level and panics
func (l *logger) Panic(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelPanic, msg, args...)
	panic(fmt.Sprintf(msg, args...))
}

// Mark logs a message at mark level for logical separation
func (l *logger) Mark(msg string, args ...interface{}) {
	l.Log(context.Background(), LevelMark, msg, args...)
}

// WithNested returns a logger with nested attributes
func (l *logger) WithNested(keyPath []string, value interface{}) Logger {
	newLogger := l.clone()
	newLogger.attrs.Set(keyPath, value)
	return newLogger
}

// WithDot returns a logger with dot notation attributes
func (l *logger) WithDot(dotPath string, value interface{}) Logger {
	newLogger := l.clone()
	newLogger.attrs.SetByDotNotation(dotPath, value)
	return newLogger
}

// WithGroup returns a logger with a group
func (l *logger) WithGroup(name string) Logger {
	newLogger := l.clone()
	newLogger.groups = append(newLogger.groups, name)
	return newLogger
}

// WithCallback returns a logger with a callback
func (l *logger) WithCallback(fn CallbackFunc) Logger {
	newLogger := l.clone()
	l.mu.Lock()
	newLogger.callbacks = append(newLogger.callbacks, fn)
	l.mu.Unlock()
	return newLogger
}

// SetHandler sets the handler for the logger
func (l *logger) SetHandler(handler Handler) {
	l.mu.Lock()
	l.handler = handler
	l.mu.Unlock()
}

// Handler returns the current handler
func (l *logger) Handler() Handler {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.handler
}

func (l *logger) clone() *logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newGroups := make([]string, len(l.groups))
	copy(newGroups, l.groups)

	newCallbacks := make([]CallbackFunc, len(l.callbacks))
	copy(newCallbacks, l.callbacks)

	return &logger{
		handler:   l.handler,
		attrs:     l.attrs.Clone(),
		groups:    newGroups,
		callbacks: newCallbacks,
	}
}

// slogCompatibility provides slog compatibility
func (l *logger) WithAttrs(attrs []slog.Attr) Logger {
	newLogger := l.clone()
	for _, attr := range attrs {
		keyPath := make([]string, len(l.groups))
		copy(keyPath, l.groups)
		keyPath = append(keyPath, attr.Key)
		newLogger.attrs.Set(keyPath, attr.Value.Any())
	}
	return newLogger
}
