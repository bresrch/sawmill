package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// === Basic Options Pattern Usage ===
	
	// Simple logger with default options
	defaultLogger := sawmill.New(sawmill.NewTextHandlerWithDefaults())
	defaultLogger.Info("Default configuration")

	// Logger with custom time format
	customTimeLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().WithTimeFormat("15:04:05"),
	))
	customTimeLogger.Info("Custom time format")

	// === Comprehensive Options Configuration ===
	
	// Text logger with all options configured
	advancedTextLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().
			WithLevel(sawmill.LevelDebug).
			WithTimeFormat("2006-01-02 15:04:05.000").
			WithAttributeFormat("flat").
			WithColorsEnabled(true).
			WithColorMappings(map[string]string{
				"user": sawmill.ColorBrightBlue,
				"api":  sawmill.ColorBrightGreen,
			}).
			WithSourceInfo(true).
			WithLevelInfo(true).
			WithStdout(),
	))

	advancedTextLogger.Debug("Advanced text logger", 
		"user.id", 123, 
		"api.endpoint", "/users",
		"api.method", "GET",
	)

	// === JSON Logger Options ===
	
	// JSON logger with pretty printing and custom attributes key
	jsonLogger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().
			WithPrettyPrint(true).
			WithAttributesKey("data").
			WithColorsEnabled(true).
			WithColorMappings(map[string]string{
				"request": sawmill.ColorCyan,
				"error":   sawmill.ColorBrightRed,
			}),
	))

	jsonLogger.Info("JSON with custom options",
		"request.id", "req-456",
		"request.path", "/api/users",
		"response.status", 200,
	)

	// === File Output Configuration ===
	
	// Logger writing to file with rotation
	fileLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().
			WithFile("/tmp/sawmill-example.log", 10*1024*1024, true). // 10MB, compressed
			WithTimeFormat("2006-01-02T15:04:05.000Z07:00").
			WithAttributeFormat("nested"),
	))

	fileLogger.Info("Writing to file", 
		"file.path", "/tmp/sawmill-example.log",
		"file.max_size_mb", 10,
		"file.compressed", true,
	)

	// === Multiple Format Comparison ===
	
	baseOptions := sawmill.NewHandlerOptions().
		WithColorsEnabled(true).
		WithColorMappings(map[string]string{
			"service": sawmill.ColorMagenta,
			"db":      sawmill.ColorYellow,
		})

	// Same options, different formatters
	textCompareLogger := sawmill.New(sawmill.NewTextHandler(baseOptions))
	jsonCompareLogger := sawmill.New(sawmill.NewJSONHandler(
		baseOptions.WithPrettyPrint(true),
	))
	xmlCompareLogger := sawmill.New(sawmill.NewXMLHandlerWithDefaults())
	yamlCompareLogger := sawmill.New(sawmill.NewYAMLHandlerWithDefaults())

	// Log the same event with different formatters
	logData := []interface{}{
		"service.name", "user-api",
		"service.version", "v1.2.3",
		"db.connection.pool_size", 10,
		"db.connection.active", 7,
		"request.duration_ms", 45,
	}

	textCompareLogger.Info("Service metrics (TEXT)", logData...)
	jsonCompareLogger.Info("Service metrics (JSON)", logData...)
	xmlCompareLogger.Info("Service metrics (XML)", logData...)
	yamlCompareLogger.Info("Service metrics (YAML)", logData...)

	// === Custom Attributes Key Demonstration ===
	
	// Different attribute keys for organization
	userDataLogger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().
			WithAttributesKey("user_data").
			WithPrettyPrint(true),
	))

	requestDataLogger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().
			WithAttributesKey("request_metadata").
			WithPrettyPrint(true),
	))

	userDataLogger.Info("User event", "profile.name", "Alice", "profile.role", "admin")
	requestDataLogger.Info("API request", "headers.content_type", "application/json", "method", "POST")

	// === Level Configuration ===
	
	// Different loggers with different minimum levels
	debugLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().WithLevel(sawmill.LevelDebug),
	))
	
	warnLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().WithLevel(sawmill.LevelWarn),
	))

	// These will be filtered based on level configuration
	debugLogger.Debug("Debug message (will show)")
	debugLogger.Info("Info message (will show)")
	debugLogger.Warn("Warn message (will show)")

	warnLogger.Debug("Debug message (filtered out)")
	warnLogger.Info("Info message (filtered out)")
	warnLogger.Warn("Warn message (will show)")
}