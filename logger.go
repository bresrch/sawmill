package sawmill

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// logger implements the Logger interface
type logger struct {
	handler   Handler
	attrs     *FlatAttributes
	groups    []string
	callbacks []CallbackFunc
	mu        sync.RWMutex
}

// New creates a new logger with the specified handler
func New(handler Handler) Logger {
	return &logger{
		handler:   handler,
		attrs:     NewFlatAttributes(),
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

	record := NewRecordFromPool(level, msg)
	record.Context = ctx

	// Only capture frame if the handler/formatter might need it
	if l.needsSourceCapture() {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		record.PC = pcs[0]
	}

	record.Attributes.Merge(l.attrs)
	l.processArgsOptimized(record, args...)

	l.mu.RLock()
	for _, callback := range l.callbacks {
		record = callback(record)
	}
	l.mu.RUnlock()

	err := l.handler.Handle(ctx, record)

	// Return record to pool after use
	ReturnRecordToPool(record)

	if err != nil {
		// Log error handling could be added here if needed
	}
}

// needsSourceCapture checks if source capture is needed
func (l *logger) needsSourceCapture() bool {
	// Check if handler implements SourceHandler interface
	if sh, ok := l.handler.(SourceHandler); ok {
		return sh.NeedsSource()
	}

	// Check if handler has NeedsSource method (BaseHandler)
	if bh, ok := l.handler.(*BaseHandler); ok {
		return bh.NeedsSource()
	}

	// Safe default - capture source info
	return true
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

	err := l.handler.Handle(ctx, record)

	// Return record to pool after use
	ReturnRecordToPool(record)

	if err != nil {
		// Log error handling could be added here if needed
	}
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

// processArgsOptimized is an optimized version of processArgs
func (l *logger) processArgsOptimized(record *Record, args ...interface{}) {
	if len(args) == 0 {
		return
	}

	// Fast path for no groups (most common case)
	if len(l.groups) == 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 >= len(args) {
				break
			}

			key, ok := args[i].(string)
			if !ok {
				continue
			}

			value := args[i+1]
			
			// Check if value is a struct and should be expanded
			if l.shouldExpandStruct(value) {
				record.Attributes.ExpandStruct(key, value)
			} else {
				// Use optimized SetFast directly for non-struct values
				record.Attributes.SetFast(key, value)
			}
		}
		return
	}

	// Slower path with groups - use pre-allocated paths where possible
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}

		key, ok := args[i].(string)
		if !ok {
			continue
		}

		value := args[i+1]

		// Build path with groups
		keyPath := make([]string, len(l.groups)+1)
		copy(keyPath, l.groups)
		keyPath[len(l.groups)] = key

		// Check if value is a struct and should be expanded
		if l.shouldExpandStruct(value) {
			pathStr := key
			if len(l.groups) > 0 {
				pathStr = fmt.Sprintf("%s.%s", strings.Join(l.groups, "."), key)
			}
			record.Attributes.ExpandStruct(pathStr, value)
		} else {
			record.Attributes.Set(keyPath, value)
		}
	}
}

// shouldExpandStruct determines if a value should be expanded as a struct
func (l *logger) shouldExpandStruct(value interface{}) bool {
	if value == nil {
		return false
	}

	val := reflect.ValueOf(value)
	// Handle pointers to structs
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	return val.Kind() == reflect.Struct
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

// As returns a temporary logger that uses the specified formatter for a single message
func (l *logger) As(formatter Formatter) AsLogger {
	return &asLogger{
		logger:    l,
		formatter: formatter,
		outputID:  l.generateOutputID(),
	}
}

// generateOutputID creates a unique identifier for correlating multiline outputs
func (l *logger) generateOutputID() string {
	bytes := make([]byte, 4) // 8 character hex string
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// asLogger implements the AsLogger interface
type asLogger struct {
	logger    *logger
	formatter Formatter
	outputID  string
}

// Log logs a message using the temporary formatter
func (al *asLogger) Log(ctx context.Context, level Level, msg string, args ...interface{}) {
	if !al.logger.handler.Enabled(ctx, level) {
		return
	}

	record := NewRecordFromPool(level, msg)
	record.Context = ctx
	record.OutputID = al.outputID

	// Only capture frame if the handler/formatter might need it
	if al.logger.needsSourceCapture() {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		record.PC = pcs[0]
	}

	record.Attributes.Merge(al.logger.attrs)
	al.logger.processArgsOptimized(record, args...)

	al.logger.mu.RLock()
	for _, callback := range al.logger.callbacks {
		record = callback(record)
	}
	al.logger.mu.RUnlock()

	// Create a temporary handler with our custom formatter
	tempHandler := &temporaryHandler{
		originalHandler: al.logger.handler,
		formatter:       al.formatter,
	}

	err := tempHandler.Handle(ctx, record)

	// Return record to pool after use
	ReturnRecordToPool(record)

	if err != nil {
		// Log error handling could be added here if needed
	}
}

// Trace logs a message at trace level using the temporary formatter
func (al *asLogger) Trace(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelTrace, msg, args...)
}

// Debug logs a message at debug level using the temporary formatter
func (al *asLogger) Debug(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelDebug, msg, args...)
}

// Info logs a message at info level using the temporary formatter
func (al *asLogger) Info(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelInfo, msg, args...)
}

// Warn logs a message at warn level using the temporary formatter
func (al *asLogger) Warn(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelWarn, msg, args...)
}

// Error logs a message at error level using the temporary formatter
func (al *asLogger) Error(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelError, msg, args...)
}

// Fatal logs a message at fatal level using the temporary formatter
func (al *asLogger) Fatal(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelFatal, msg, args...)
}

// Panic logs a message at panic level using the temporary formatter and panics
func (al *asLogger) Panic(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelPanic, msg, args...)
	panic(fmt.Sprintf(msg, args...))
}

// Mark logs a message at mark level using the temporary formatter
func (al *asLogger) Mark(msg string, args ...interface{}) {
	al.Log(context.Background(), LevelMark, msg, args...)
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

// HTTPErrorLog returns a *log.Logger compatible with http.Server.ErrorLog
//
// Example usage:
//   logger := sawmill.Default()
//   srv := &http.Server{
//       Addr:     ":8080",
//       Handler:  router,
//       ErrorLog: logger.HTTPErrorLog(),
//   }
func (l *logger) HTTPErrorLog() *log.Logger {
	return log.New(&httpErrorLogWriter{logger: l}, "", 0)
}

// httpErrorLogWriter adapts sawmill logger for use with standard log.Logger
type httpErrorLogWriter struct {
	logger *logger
}

// Write implements io.Writer interface for HTTP error logging
func (w *httpErrorLogWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// Remove trailing newline if present
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Error(msg)
	return len(p), nil
}
