package sawmill

import (
	"context"
	"io"
	"log/slog"
	"time"
)

// Level represents logging levels
type Level int

const (
	LevelTrace Level = iota - 8
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
	LevelMark
)

// Record represents a single log entry with flat attributes
type Record struct {
	Time       time.Time
	Level      Level
	Message    string
	Attributes *FlatAttributes
	Context    context.Context
	PC         uintptr
	OutputID   string // Unique identifier for correlating multiline outputs
}

// NewRecord creates a new log record
func NewRecord(level Level, msg string) *Record {
	return &Record{
		Time:       time.Now(),
		Level:      level,
		Message:    msg,
		Attributes: NewFlatAttributes(),
		Context:    context.Background(),
	}
}

// With adds nested attributes to the record
func (r *Record) With(keyPath []string, value interface{}) *Record {
	r.Attributes.Set(keyPath, value)
	return r
}

// WithDot adds attributes using dot notation
func (r *Record) WithDot(dotPath string, value interface{}) *Record {
	r.Attributes.SetByDotNotation(dotPath, value)
	return r
}

// CallbackFunc represents a dynamic callback for runtime log modification
type CallbackFunc func(record *Record) *Record

// Formatter defines the interface for log formatting
type Formatter interface {
	Format(record *Record) ([]byte, error)
	ContentType() string
}

// Buffer defines the interface for output buffering
type Buffer interface {
	io.Writer
	Flush() error
	Close() error
	Size() int64
	Reset()
}

// Handler defines the interface for log handling
type Handler interface {
	Handle(ctx context.Context, record *Record) error
	WithAttrs(attrs []slog.Attr) Handler
	WithGroup(name string) Handler
	Enabled(ctx context.Context, level Level) bool
}

// SourceHandler extends Handler to indicate if source info is needed
type SourceHandler interface {
	Handler
	NeedsSource() bool
}

// Logger represents the main logging interface
type Logger interface {
	Log(ctx context.Context, level Level, msg string, args ...interface{})
	LogRecord(ctx context.Context, record *Record)

	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Panic(msg string, args ...interface{})
	Mark(msg string, args ...interface{})

	WithNested(keyPath []string, value interface{}) Logger
	WithDot(dotPath string, value interface{}) Logger
	WithGroup(name string) Logger
	WithCallback(fn CallbackFunc) Logger
	SetHandler(handler Handler)
	Handler() Handler
	As(formatter Formatter) AsLogger
}

// AsLogger provides temporary format switching for single messages
type AsLogger interface {
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Panic(msg string, args ...interface{})
	Mark(msg string, args ...interface{})
	Log(ctx context.Context, level Level, msg string, args ...interface{})
}

// Destination represents various output targets
type Destination interface {
	Write(data []byte) (int, error)
	Close() error
}

// FileDestination represents file output configuration
type FileDestination struct {
	Path     string
	MaxSize  int64
	MaxAge   time.Duration
	Compress bool
}

// WriterDestination wraps an io.Writer for output
type WriterDestination struct {
	Writer io.Writer
}

// NetworkDestination represents network output configuration
type NetworkDestination struct {
	Protocol string
	Address  string
}
