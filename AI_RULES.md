# Sawmill Logging Library - AI Integration Guide

Enhanced Go logging library extending `log/slog` with nested attributes, dynamic callbacks, and multi-format output.

## Code Quality Rules
- Code comments describe functionality only, not reasoning or decision-making
- Use descriptive phrases, avoid explanatory sentences
- No comments starting with "We", "This", "I", or "The AI"

## Comment Examples
- Good: // User validation endpoint
- Good: // Converts to milliseconds  
- Good: // Required format for API
- Bad: // We need this because the API requires it
- Bad: // I chose this approach for better performance
- Bad: // This was added to handle edge cases

## Core Architecture

Sawmill extends Go's `log/slog` with nested attributes, dynamic callbacks, and multi-format output. Key architectural components:

- **Logger**: Immutable logger instances with method chaining
- **Handlers**: Output destination management (console, file, multi-output)
- **Formatters**: Output format control (JSON, XML, YAML, text, key-value)
- **Callbacks**: Runtime context injection via middleware pattern
- **RecursiveMap**: Hierarchical attribute storage with infinite depth

## Implementation Decision Tree

### 1. Handler Selection
```
New application → NewTextHandlerWithDefaults() for development
Existing slog → Drop-in replacement with sawmill.New(handler)
Production → NewJSONHandler() with file output
Monitoring → NewKeyValueHandler() for machine parsing
Multiple outputs → NewMultiHandler()
```

### 2. Configuration Pattern
```go
// Always use options pattern for production
logger := sawmill.New(sawmill.NewJSONHandler(
    sawmill.NewHandlerOptions().
        WithLevel(sawmill.LevelInfo).
        WithFile("/var/log/app.jsonl", 100*1024*1024, true).
        WithColorsEnabled(false),
))
```

### 3. Attribute Organization
```
Simple key-value → logger.Info("message", "key", value)
Nested data → logger.WithDot("user.profile.email", email)
Request context → Use callbacks for cross-cutting concerns
Business logic → Direct attributes for domain-specific data
```

## Code Generation Patterns

### Basic Logging Implementation
```go
func (s *Service) ProcessUser(userID int) error {
    logger := s.logger.WithDot("user.id", userID)
    
    logger.Info("Processing started")
    
    if err := s.validateUser(userID); err != nil {
        logger.Error("Validation failed", "error", err.Error())
        return err
    }
    
    logger.Info("Processing completed")
    return nil
}
```

### HTTP Middleware Pattern
```go
func LoggingMiddleware(logger sawmill.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestLogger := logger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
                record.WithDot("request.method", r.Method)
                record.WithDot("request.path", r.URL.Path)
                record.WithDot("request.id", getRequestID(r))
                return record
            })
            
            ctx := context.WithValue(r.Context(), "logger", requestLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Error Handling Integration
```go
func (s *Service) HandleError(err error, logger sawmill.Logger) {
    switch e := err.(type) {
    case *ValidationError:
        logger.Warn("Validation error", 
            "error.type", "validation",
            "error.field", e.Field,
            "error.value", e.Value)
    case *DatabaseError:
        logger.Error("Database error",
            "error.type", "database",
            "error.query", e.Query,
            "error.duration_ms", e.Duration.Milliseconds())
    default:
        logger.Error("Unexpected error", "error", err.Error())
    }
}
```

## Environment-Specific Configurations

### Development
```go
func NewDevelopmentLogger() sawmill.Logger {
    return sawmill.New(sawmill.NewTextHandler(
        sawmill.NewHandlerOptions().
            WithLevel(sawmill.LevelDebug).
            WithColorsEnabled(true).
            WithAttributeFormat("nested"),
    ))
}
```

### Production
```go
func NewProductionLogger() sawmill.Logger {
    return sawmill.New(sawmill.NewMultiHandler(
        // Structured logs for aggregation
        sawmill.NewJSONHandler(
            sawmill.NewHandlerOptions().
                WithFile("/var/log/app.jsonl", 200*1024*1024, true).
                WithLevel(sawmill.LevelInfo),
        ),
        // Console warnings/errors
        sawmill.NewTextHandler(
            sawmill.NewHandlerOptions().
                WithLevel(sawmill.LevelWarn).
                WithStdout(),
        ),
    ))
}
```

### Monitoring Integration
```go
func NewMonitoringLogger() sawmill.Logger {
    return sawmill.New(sawmill.NewKeyValueHandler(
        sawmill.NewHandlerOptions().
            WithFile("/var/log/metrics.log", 50*1024*1024, false).
            WithLevel(sawmill.LevelInfo),
    ))
}
```

## Performance Considerations

### High-Throughput Applications
- Use `NewJSONHandler()` with file output for >1000 logs/sec
- Avoid deep nesting in hot paths (>5 levels)
- Limit callback count to <3 per logger
- Use `WithLevel()` to filter logs at handler level

### Memory Management
- Logger instances are immutable - safe for concurrent use
- `WithDot()` and `WithCallback()` create new logger instances
- Callbacks executed per log entry - keep lightweight
- RecursiveMap creates object allocation for nested attributes

### File I/O Optimization
```go
// Buffered file output with rotation
opts := sawmill.NewHandlerOptions().
    WithFile("/var/log/app.log", 100*1024*1024, true). // 100MB, compressed
    WithLevel(sawmill.LevelInfo)
```

## Common Anti-Patterns to Avoid

### ❌ String Interpolation
```go
logger.Info(fmt.Sprintf("User %d logged in", userID)) // Don't do this
```

### ✅ Structured Attributes
```go
logger.Info("User logged in", "user.id", userID) // Do this
```

### ❌ Global Logger Mutation
```go
var globalLogger sawmill.Logger
globalLogger = globalLogger.WithDot("service", "auth") // Don't do this
```

### ✅ Context-Specific Loggers
```go
func (s *AuthService) Login(userID int) {
    logger := s.logger.WithDot("user.id", userID) // Do this
    logger.Info("Login attempt")
}
```

## Integration Testing Patterns

### Test Logger Setup
```go
func TestWithSawmill(t *testing.T) {
    var logOutput bytes.Buffer
    testLogger := sawmill.New(sawmill.NewJSONHandler(
        sawmill.NewHandlerOptions().
            WithWriter(&logOutput).
            WithLevel(sawmill.LevelDebug),
    ))
    
    // Use testLogger in tests
    service := NewService(testLogger)
    service.ProcessUser(123)
    
    // Verify log output
    assert.Contains(t, logOutput.String(), `"user.id":123`)
}
```

### Mock Logger Pattern
```go
type MockLogger struct {
    Entries []LogEntry
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
    m.Entries = append(m.Entries, LogEntry{Level: "INFO", Message: msg, Args: args})
}
```

## Dependency Injection

### Constructor Pattern
```go
type Service struct {
    logger sawmill.Logger
    db     *sql.DB
}

func NewService(logger sawmill.Logger, db *sql.DB) *Service {
    return &Service{
        logger: logger.WithDot("service", "user"),
        db:     db,
    }
}
```

### Interface-Based Injection
```go
type Logger interface {
    Info(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    WithDot(key string, value interface{}) Logger
}

// Sawmill Logger implements this interface naturally
```

## Architecture Patterns

### Environment-Based Configuration
```go
func NewLogger(env string) sawmill.Logger {
    switch env {
    case "development":
        return sawmill.New(sawmill.NewTextHandler(
            sawmill.NewHandlerOptions().
                WithLevel(sawmill.LevelDebug).
                WithColorsEnabled(true)))
    case "production":
        return sawmill.New(sawmill.NewJSONHandler(
            sawmill.NewHandlerOptions().
                WithFile("/var/log/app.jsonl", 100*1024*1024, true).
                WithLevel(sawmill.LevelInfo)))
    default:
        return sawmill.DefaultLogger
    }
}
```

### Request Context Injection
```go
func WithRequestContext(logger sawmill.Logger, r *http.Request) sawmill.Logger {
    return logger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
        record.WithDot("request.id", middleware.GetRequestID(r))
        record.WithDot("request.method", r.Method)
        record.WithDot("request.path", r.URL.Path)
        return record
    })
}
```

### Error Enrichment
```go
func WithErrorContext(logger sawmill.Logger) sawmill.Logger {
    return logger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
        if record.Level >= sawmill.LevelError {
            record.WithDot("runtime.goroutines", runtime.NumGoroutine())
            record.WithDot("runtime.stack_trace", debug.Stack())
        }
        return record
    })
}
```

## Production Integration

### Monitoring Systems
```go
// Prometheus metrics logging
metricsLogger := sawmill.New(sawmill.NewKeyValueHandler(
    sawmill.NewHandlerOptions().WithFile("/var/log/metrics.log")))

// Application event logging  
auditLogger := sawmill.New(sawmill.NewJSONHandler(
    sawmill.NewHandlerOptions().WithFile("/var/log/audit.jsonl")))
```

### Service Mesh Integration
```go
func WithTracing(logger sawmill.Logger, span trace.Span) sawmill.Logger {
    return logger.WithCallback(func(record *sawmill.Record) *sawmill.Record {
        record.WithDot("trace.span_id", span.SpanContext().SpanID())
        record.WithDot("trace.trace_id", span.SpanContext().TraceID())
        return record
    })
}
```

### Database Connection Pools
```go
func LogDBMetrics(logger sawmill.Logger, db *sql.DB) {
    stats := db.Stats()
    logger.Info("Database pool status",
        "open_connections", stats.OpenConnections,
        "in_use", stats.InUse,
        "idle", stats.Idle)
}
```

## Implementation Guidelines

1. **Structured over interpolated**: `logger.Info("user login", "user_id", 123)` not `logger.Info("user 123 login")`
2. **Consistent key naming**: Use dot notation for hierarchical data (`user.profile.email`)
3. **Level discipline**: Debug for development, Info for business events, Error for actionable failures
4. **Context isolation**: Use callbacks for request/session context, direct attributes for business data
5. **Format specialization**: JSON for aggregation, key-value for metrics, text for development

## Deployment Considerations

- Use JSON format for log aggregation systems (ELK, Splunk)
- Use key-value format for metrics collection (Prometheus)
- Configure file rotation based on disk space constraints
- Set appropriate log levels per environment
- Use callbacks for distributed tracing integration
- Consider multi-handler setup for different log consumers
