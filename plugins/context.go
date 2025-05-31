package plugins

import (
	"context"
	"runtime"
	"time"

	"github.com/bresrch/sawmill"
)

// ContextOptions configures what data to extract from context
type ContextOptions struct {
	IncludeDeadline bool
	IncludeValues   []interface{} // Specific context keys to extract
	FieldPrefix     string        // Prefix for context fields
	ValuePrefix     string        // Prefix for context values
}

// DefaultContextOptions returns sensible defaults for context extraction
func DefaultContextOptions() *ContextOptions {
	return &ContextOptions{
		IncludeDeadline: true,
		IncludeValues:   []interface{}{"request_id", "user_id", "trace_id", "span_id"},
		FieldPrefix:     "context.",
		ValuePrefix:     "context.values.",
	}
}

// WithContext extracts data from a context using default options
func WithContext(logger sawmill.Logger, ctx context.Context) sawmill.Logger {
	return WithContextOptions(logger, ctx, DefaultContextOptions())
}

// WithContextOptions extracts data from a context using custom options
func WithContextOptions(logger sawmill.Logger, ctx context.Context, opts *ContextOptions) sawmill.Logger {
	if ctx == nil || opts == nil {
		return logger
	}

	result := logger

	// Include deadline information
	if opts.IncludeDeadline {
		if deadline, ok := ctx.Deadline(); ok {
			result = result.WithDot(opts.FieldPrefix+"deadline", deadline.Unix())
			result = result.WithDot(opts.FieldPrefix+"time_left_ms", time.Until(deadline).Milliseconds())
			result = result.WithDot(opts.FieldPrefix+"has_deadline", true)
		} else {
			result = result.WithDot(opts.FieldPrefix+"has_deadline", false)
		}
	}

	// Include specific context values
	for _, key := range opts.IncludeValues {
		if value := ctx.Value(key); value != nil {
			fieldName := opts.ValuePrefix + normalizeContextKey(key)
			result = result.WithDot(fieldName, value)
		}
	}

	return result
}

// WithRuntime adds runtime information to the logger
func WithRuntime(logger sawmill.Logger) sawmill.Logger {
	return logger.
		WithDot("runtime.goroutines", runtime.NumGoroutine()).
		WithDot("runtime.timestamp", time.Now().Unix()).
		WithDot("runtime.timestamp_ns", time.Now().UnixNano())
}

// WithRuntimeMemory adds memory statistics to the logger (useful for error logs)
func WithRuntimeMemory(logger sawmill.Logger) sawmill.Logger {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return logger.
		WithDot("runtime.memory.alloc_bytes", m.Alloc).
		WithDot("runtime.memory.total_alloc_bytes", m.TotalAlloc).
		WithDot("runtime.memory.sys_bytes", m.Sys).
		WithDot("runtime.memory.gc_count", m.NumGC)
}

// normalizeContextKey converts context keys to string format
func normalizeContextKey(key interface{}) string {
	switch k := key.(type) {
	case string:
		return k
	default:
		return "unknown_key"
	}
}
