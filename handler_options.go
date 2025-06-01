package sawmill

import (
	"io"
	"os"
	"time"
)

// HandlerOptions configures handler behavior using the functional options pattern
type HandlerOptions struct {
	level         Level
	destination   Destination
	sawmillOpts   *SawmillOptions
	attributesKey string
	colorMappings map[string]string
	enableColors  bool
	timeFormat    string
	prettyPrint   bool
	includeSource bool
	includeLevel  bool
	colorOutput   bool
	attrFormat    string
}

// HandlerOption is a function that configures HandlerOptions
type HandlerOption func(*HandlerOptions)

// NewHandlerOptions creates a new HandlerOptions with the given options
func NewHandlerOptions(options ...HandlerOption) *HandlerOptions {
	opts := &HandlerOptions{
		level:         LevelInfo,
		destination:   NewWriterDestination(os.Stdout),
		sawmillOpts:   nil,
		attributesKey: "attributes",
		colorMappings: make(map[string]string),
		enableColors:  false,
		timeFormat:    time.RFC3339,
		prettyPrint:   false,
		includeSource: true,
		includeLevel:  true,
		colorOutput:   false,
		attrFormat:    "nested",
	}

	for _, option := range options {
		option(opts)
	}

	return opts
}

// WithLevel sets the minimum log level
func WithLevel(level Level) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.level = level
	}
}

// WithDestination sets the output destination
func WithDestination(dest Destination) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.destination = dest
	}
}

// WithSawmillOptions sets the sawmill configuration options
func WithSawmillOptions(sawmillOpts *SawmillOptions) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.sawmillOpts = sawmillOpts
	}
}

// WithAttributesKey sets the key name for attributes in formatted output
func WithAttributesKey(key string) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.attributesKey = key
	}
}

// WithColorMappings sets custom color mappings for specific keys
func WithColorMappings(mappings map[string]string) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.colorMappings = mappings
	}
}

// WithColorsEnabled enables or disables color output
func WithColorsEnabled(enabled bool) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.enableColors = enabled
		opts.colorOutput = enabled
	}
}

// WithTimeFormat sets the time format for timestamps
func WithTimeFormat(format string) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.timeFormat = format
	}
}

// WithPrettyPrint enables or disables pretty printing for JSON
func WithPrettyPrint(enabled bool) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.prettyPrint = enabled
	}
}

// WithSourceInfo enables or disables source location information
func WithSourceInfo(enabled bool) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.includeSource = enabled
	}
}

// WithLevelInfo enables or disables log level in output
func WithLevelInfo(enabled bool) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.includeLevel = enabled
	}
}

// WithAttributeFormat sets the attribute format ("flat" or "nested")
func WithAttributeFormat(format string) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.attrFormat = format
	}
}

// WithWriter is a convenience method to set a writer destination
func WithWriter(writer io.Writer) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.destination = NewWriterDestination(writer)
	}
}

// WithFile is a convenience method to set a file destination
func WithFile(path string, maxSize int64, compress bool) HandlerOption {
	return func(opts *HandlerOptions) {
		opts.destination = NewFileDestination(path, maxSize, 0, compress)
	}
}

// WithStdout is a convenience method to set stdout as destination
func WithStdout() HandlerOption {
	return func(opts *HandlerOptions) {
		opts.destination = NewWriterDestination(os.Stdout)
	}
}

// WithStderr is a convenience method to set stderr as destination
func WithStderr() HandlerOption {
	return func(opts *HandlerOptions) {
		opts.destination = NewWriterDestination(os.Stderr)
	}
}
