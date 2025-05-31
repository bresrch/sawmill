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

func NewSawmillOptions() *SawmillOptions {
	return &SawmillOptions{
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
