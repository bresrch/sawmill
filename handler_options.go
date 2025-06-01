package sawmill

import (
	"io"
	"os"
	"time"
)

// HandlerOptions configures handler behavior using the options pattern
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

// NewHandlerOptions creates a new HandlerOptions with sensible defaults
func NewHandlerOptions() *HandlerOptions {
	return &HandlerOptions{
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
}

// WithLevel sets the minimum log level
func (opts *HandlerOptions) WithLevel(level Level) *HandlerOptions {
	opts.level = level
	return opts
}

// WithDestination sets the output destination
func (opts *HandlerOptions) WithDestination(dest Destination) *HandlerOptions {
	opts.destination = dest
	return opts
}

// WithSawmillOptions sets the sawmill configuration options
func (opts *HandlerOptions) WithSawmillOptions(sawmillOpts *SawmillOptions) *HandlerOptions {
	opts.sawmillOpts = sawmillOpts
	return opts
}

// WithAttributesKey sets the key name for attributes in formatted output
func (opts *HandlerOptions) WithAttributesKey(key string) *HandlerOptions {
	opts.attributesKey = key
	return opts
}

// WithColorMappings sets custom color mappings for specific keys
func (opts *HandlerOptions) WithColorMappings(mappings map[string]string) *HandlerOptions {
	opts.colorMappings = mappings
	return opts
}

// WithColorsEnabled enables or disables color output
func (opts *HandlerOptions) WithColorsEnabled(enabled bool) *HandlerOptions {
	opts.enableColors = enabled
	opts.colorOutput = enabled
	return opts
}

// WithTimeFormat sets the time format for timestamps
func (opts *HandlerOptions) WithTimeFormat(format string) *HandlerOptions {
	opts.timeFormat = format
	return opts
}

// WithPrettyPrint enables or disables pretty printing for JSON
func (opts *HandlerOptions) WithPrettyPrint(enabled bool) *HandlerOptions {
	opts.prettyPrint = enabled
	return opts
}

// WithSourceInfo enables or disables source location information
func (opts *HandlerOptions) WithSourceInfo(enabled bool) *HandlerOptions {
	opts.includeSource = enabled
	return opts
}

// WithLevelInfo enables or disables log level in output
func (opts *HandlerOptions) WithLevelInfo(enabled bool) *HandlerOptions {
	opts.includeLevel = enabled
	return opts
}

// WithAttributeFormat sets the attribute format ("flat" or "nested")
func (opts *HandlerOptions) WithAttributeFormat(format string) *HandlerOptions {
	opts.attrFormat = format
	return opts
}

// WithWriter is a convenience method to set a writer destination
func (opts *HandlerOptions) WithWriter(writer io.Writer) *HandlerOptions {
	opts.destination = NewWriterDestination(writer)
	return opts
}

// WithFile is a convenience method to set a file destination
func (opts *HandlerOptions) WithFile(path string, maxSize int64, compress bool) *HandlerOptions {
	opts.destination = NewFileDestination(path, maxSize, 0, compress)
	return opts
}

// WithStdout is a convenience method to set stdout as destination
func (opts *HandlerOptions) WithStdout() *HandlerOptions {
	opts.destination = NewWriterDestination(os.Stdout)
	return opts
}

// WithStderr is a convenience method to set stderr as destination
func (opts *HandlerOptions) WithStderr() *HandlerOptions {
	opts.destination = NewWriterDestination(os.Stderr)
	return opts
}
