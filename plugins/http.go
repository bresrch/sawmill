package plugins

import (
	"net/http"

	"github.com/bresrch/sawmill"
)

// HTTPRequestOptions configures what data to extract from HTTP requests
type HTTPRequestOptions struct {
	IncludeMethod      bool
	IncludePath        bool
	IncludeQuery       bool
	IncludeHeaders     []string // Specific headers to include, empty means all
	IncludeRemoteAddr  bool
	IncludeUserAgent   bool
	IncludeHost        bool
	IncludeScheme      bool
	IncludeContentInfo bool
	HeaderPrefix       string // Prefix for header fields, default "http.request.headers."
	FieldPrefix        string // Prefix for all fields, default "http.request."
}

// HTTPResponseOptions configures what data to extract from HTTP responses
type HTTPResponseOptions struct {
	IncludeStatus  bool
	IncludeHeaders []string // Specific headers to include
	IncludeSize    bool     // Response size if available
	HeaderPrefix   string   // Prefix for header fields
	FieldPrefix    string   // Prefix for all fields
}

// DefaultHTTPRequestOptions returns sensible defaults for HTTP request extraction
func DefaultHTTPRequestOptions() *HTTPRequestOptions {
	return &HTTPRequestOptions{
		IncludeMethod:      true,
		IncludePath:        true,
		IncludeQuery:       true,
		IncludeHeaders:     []string{"X-Request-ID", "Authorization", "Content-Type"},
		IncludeRemoteAddr:  true,
		IncludeUserAgent:   true, // This provides http.request.user_agent
		IncludeHost:        true,
		IncludeScheme:      true,
		IncludeContentInfo: true,
		HeaderPrefix:       "http.request.headers.",
		FieldPrefix:        "http.request.",
	}
}

// DefaultHTTPResponseOptions returns sensible defaults for HTTP response extraction
func DefaultHTTPResponseOptions() *HTTPResponseOptions {
	return &HTTPResponseOptions{
		IncludeStatus:  true,
		IncludeHeaders: []string{"Content-Type", "Content-Length", "X-Response-ID"},
		IncludeSize:    true,
		HeaderPrefix:   "http.response.headers.",
		FieldPrefix:    "http.response.",
	}
}

// WithHTTPRequest extracts data from an HTTP request using default options
func WithHTTPRequest(logger sawmill.Logger, req *http.Request) sawmill.Logger {
	return WithHTTPRequestOptions(logger, req, DefaultHTTPRequestOptions())
}

// WithHTTPRequestOptions extracts data from an HTTP request using custom options
func WithHTTPRequestOptions(logger sawmill.Logger, req *http.Request, opts *HTTPRequestOptions) sawmill.Logger {
	if req == nil || opts == nil {
		return logger
	}

	result := logger

	if opts.IncludeMethod {
		result = result.WithDot(opts.FieldPrefix+"method", req.Method)
	}

	if opts.IncludePath {
		result = result.WithDot(opts.FieldPrefix+"path", req.URL.Path)
	}

	if opts.IncludeQuery && req.URL.RawQuery != "" {
		result = result.WithDot(opts.FieldPrefix+"query", req.URL.RawQuery)
	}

	if opts.IncludeRemoteAddr {
		result = result.WithDot(opts.FieldPrefix+"remote_addr", req.RemoteAddr)
	}

	if opts.IncludeUserAgent {
		if ua := req.UserAgent(); ua != "" {
			result = result.WithDot(opts.FieldPrefix+"user_agent", ua)
		}
	}

	if opts.IncludeHost {
		result = result.WithDot(opts.FieldPrefix+"host", req.Host)
	}

	if opts.IncludeScheme {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		if forwarded := req.Header.Get("X-Forwarded-Proto"); forwarded != "" {
			scheme = forwarded
		}
		result = result.WithDot(opts.FieldPrefix+"scheme", scheme)
	}

	if opts.IncludeContentInfo {
		if req.ContentLength > 0 {
			result = result.WithDot(opts.FieldPrefix+"content_length", req.ContentLength)
		}
	}

	// Include headers
	if len(opts.IncludeHeaders) > 0 {
		for _, headerName := range opts.IncludeHeaders {
			if value := req.Header.Get(headerName); value != "" {
				fieldName := opts.HeaderPrefix + normalizeHeaderName(headerName)
				result = result.WithDot(fieldName, value)
			}
		}
	}

	return result
}

// WithHTTPResponse extracts data from an HTTP response using default options
func WithHTTPResponse(logger sawmill.Logger, resp *http.Response) sawmill.Logger {
	return WithHTTPResponseOptions(logger, resp, DefaultHTTPResponseOptions())
}

// WithHTTPResponseOptions extracts data from an HTTP response using custom options
func WithHTTPResponseOptions(logger sawmill.Logger, resp *http.Response, opts *HTTPResponseOptions) sawmill.Logger {
	if resp == nil || opts == nil {
		return logger
	}

	result := logger

	if opts.IncludeStatus {
		result = result.WithDot(opts.FieldPrefix+"status_code", resp.StatusCode)
		result = result.WithDot(opts.FieldPrefix+"status", resp.Status)
	}

	// Include headers
	if len(opts.IncludeHeaders) > 0 {
		for _, headerName := range opts.IncludeHeaders {
			if value := resp.Header.Get(headerName); value != "" {
				fieldName := opts.HeaderPrefix + normalizeHeaderName(headerName)
				result = result.WithDot(fieldName, value)
			}
		}
	}

	if opts.IncludeSize && resp.ContentLength > 0 {
		result = result.WithDot(opts.FieldPrefix+"content_length", resp.ContentLength)
	}

	return result
}

// WithHTTPResponseWriter extracts data from an HTTP ResponseWriter (for middleware use)
func WithHTTPResponseWriter(logger sawmill.Logger, w http.ResponseWriter) sawmill.Logger {
	return WithHTTPResponseWriterOptions(logger, w, DefaultHTTPResponseOptions())
}

// WithHTTPResponseWriterOptions extracts data from an HTTP ResponseWriter using custom options
func WithHTTPResponseWriterOptions(logger sawmill.Logger, w http.ResponseWriter, opts *HTTPResponseOptions) sawmill.Logger {
	if w == nil || opts == nil {
		return logger
	}

	result := logger

	// Try to extract status code if the ResponseWriter supports it
	if statusWriter, ok := w.(interface{ Status() int }); ok {
		if opts.IncludeStatus {
			status := statusWriter.Status()
			result = result.WithDot(opts.FieldPrefix+"status_code", status)
		}
	}

	// Extract headers from the ResponseWriter
	if len(opts.IncludeHeaders) > 0 {
		headers := w.Header()
		for _, headerName := range opts.IncludeHeaders {
			if value := headers.Get(headerName); value != "" {
				fieldName := opts.HeaderPrefix + normalizeHeaderName(headerName)
				result = result.WithDot(fieldName, value)
			}
		}
	}

	// Try to extract response size if available
	if opts.IncludeSize {
		if sizeWriter, ok := w.(interface{ Size() int64 }); ok {
			result = result.WithDot(opts.FieldPrefix+"size", sizeWriter.Size())
		}
	}

	return result
}

// normalizeHeaderName converts header names to lowercase with underscores
func normalizeHeaderName(name string) string {
	name = strings.ToLower(name) // Convert to lowercase using standard library
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if r == '-' {
			result = append(result, '_')
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
