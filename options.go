package sawmill

import "fmt"

type SawmillOptions struct {
	// LogFile is the path to the log file. If empty, logs will be written to stdout.
	LogFile string `json:"log_file,omitempty"`
	// LogLevel is the logging level. Default is "info".
	LogLevel string `json:"log_level,omitempty"`
	// LogFormat is the format of the logs. Default is "text".
	LogFormat string `json:"log_format,omitempty"`
	// CallInfo indicates whether to include call information in the logs.
	CallInfo bool `json:"call_info,omitempty"`
	// MaxSize is the maximum size of the log file in megabytes before it gets rotated.
	MaxSize int `json:"max_size,omitempty"`
	// MaxBackups is the maximum number of old log files to keep.
	MaxBackups int `json:"max_backups,omitempty"`
	// MaxAge is the maximum number of days to keep old log files.
	MaxAge int `json:"max_age,omitempty"`
	// Compress indicates whether to compress old log files.
	Compress bool `json:"compress,omitempty"`
	// EnableDebug indicates whether to enable debug logging.
	EnableDebug bool `json:"enable_debug,omitempty"`
	// EnableInfo indicates whether to enable info logging.
	EnableInfo bool `json:"enable_info,omitempty"`
	// EnableWarn indicates whether to enable warn logging.
	EnableWarn bool `json:"enable_warn,omitempty"`
	// EnableError indicates whether to enable error logging.
	EnableError bool `json:"enable_error,omitempty"`
	// EnableFatal indicates whether to enable fatal logging.
	EnableFatal bool `json:"enable_fatal,omitempty"`
	// EnablePanic indicates whether to enable panic logging.
	EnablePanic bool `json:"enable_panic,omitempty"`
	// EnableTrace indicates whether to enable trace logging.
	EnableTrace bool `json:"enable_trace,omitempty"`
	// EnableMetrics indicates whether to enable metrics logging.
	EnableMetrics bool `json:"enable_metrics,omitempty"`
}

// SawmillOption is a function that configures SawmillOptions
type SawmillOption func(*SawmillOptions)

func NewSawmillOptions(options ...SawmillOption) *SawmillOptions {
	opts := &SawmillOptions{
		LogFile:       "",
		LogLevel:      "info",
		LogFormat:     "text",
		CallInfo:      true,
		MaxSize:       100, // Default to 100 MB
		MaxBackups:    7,   // Default to 7 backups
		MaxAge:        30,  // Default to 30 days
		Compress:      false,
		EnableDebug:   false,
		EnableInfo:    true,
		EnableWarn:    true,
		EnableError:   true,
		EnableFatal:   true,
		EnablePanic:   true,
		EnableTrace:   false,
		EnableMetrics: false,
	}
	
	for _, option := range options {
		option(opts)
	}
	
	return opts
}

// WithLogFile sets the log file path
func WithLogFile(logFile string) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.LogFile = logFile
	}
}

// WithLogLevel sets the log level
func WithLogLevel(logLevel string) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.LogLevel = logLevel
	}
}

// WithLogFormat sets the log format
func WithLogFormat(logFormat string) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.LogFormat = logFormat
	}
}

// WithCallInfo enables or disables call information
func WithCallInfo(callInfo bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.CallInfo = callInfo
	}
}

// WithMaxSize sets the maximum log file size
func WithMaxSize(maxSize int) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.MaxSize = maxSize
	}
}

// WithMaxBackups sets the maximum number of backup files
func WithMaxBackups(maxBackups int) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.MaxBackups = maxBackups
	}
}

// WithMaxAge sets the maximum age of log files
func WithMaxAge(maxAge int) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.MaxAge = maxAge
	}
}

// WithCompress enables or disables compression
func WithCompress(compress bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.Compress = compress
	}
}

// WithEnableDebug enables or disables debug logging
func WithEnableDebug(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableDebug = enable
	}
}

// WithEnableInfo enables or disables info logging
func WithEnableInfo(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableInfo = enable
	}
}

// WithEnableWarn enables or disables warn logging
func WithEnableWarn(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableWarn = enable
	}
}

// WithEnableError enables or disables error logging
func WithEnableError(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableError = enable
	}
}

// WithEnableFatal enables or disables fatal logging
func WithEnableFatal(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableFatal = enable
	}
}

// WithEnablePanic enables or disables panic logging
func WithEnablePanic(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnablePanic = enable
	}
}

// WithEnableTrace enables or disables trace logging
func WithEnableTrace(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableTrace = enable
	}
}

// WithEnableMetrics enables or disables metrics logging
func WithEnableMetrics(enable bool) SawmillOption {
	return func(opts *SawmillOptions) {
		opts.EnableMetrics = enable
	}
}

// Validate checks the SawmillOptions for any invalid configurations.
func (opts *SawmillOptions) Validate() error {
	if opts.MaxSize <= 0 {
		return fmt.Errorf("max_size must be greater than 0")
	}
	if opts.MaxBackups < 0 {
		return fmt.Errorf("max_backups cannot be negative")
	}
	if opts.MaxAge < 0 {
		return fmt.Errorf("max_age cannot be negative")
	}
	if opts.LogLevel != "debug" && opts.LogLevel != "info" && opts.LogLevel != "warn" && opts.LogLevel != "error" && opts.LogLevel != "fatal" && opts.LogLevel != "panic" && opts.LogLevel != "trace" {
		return fmt.Errorf("invalid log_level: %s", opts.LogLevel)
	}
	return nil
}
