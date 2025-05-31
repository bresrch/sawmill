package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// Custom color mappings for different key patterns
	colorMappings := map[string]string{
		"user":        sawmill.ColorBrightBlue,   // All user.* keys will be bright blue
		"request":     sawmill.ColorBrightGreen,  // All request.* keys will be bright green
		"server":      sawmill.ColorYellow,       // All server.* keys will be yellow
		"error":       sawmill.ColorBrightRed,    // All error.* keys will be bright red
		"batch_id":    sawmill.ColorMagenta,      // Specific key coloring
		"status_code": sawmill.ColorCyan,         // HTTP status codes in cyan
	}

	// Text logger with colors
	textLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().
			WithColorsEnabled(true).
			WithColorMappings(colorMappings),
	))

	// JSON logger with colors
	jsonLogger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().
			WithColorsEnabled(true).
			WithColorMappings(colorMappings).
			WithPrettyPrint(true),
	))

	// Demonstrate colored text output
	textLogger.Info("User authentication",
		"user.id", 12345,
		"user.name", "alice_smith",
		"user.active", true,
		"user.balance", 1250.75,
		"request.method", "POST",
		"request.path", "/api/auth",
		"server.hostname", "web01",
		"server.port", 8080,
		"response_time_ms", 150,
		"success", true,
	)

	textLogger.Warn("Rate limit warning",
		"user.id", 67890,
		"request.rate", 95,
		"request.limit", 100,
		"server.load", 85.2,
		"status_code", 429,
	)

	textLogger.Error("Database error",
		"error.code", "DB_CONNECTION_FAILED",
		"error.message", "Connection timeout",
		"error.retry_count", 3,
		"server.database.host", "db-primary.internal",
		"server.database.port", 5432,
	)

	// Demonstrate colored JSON output
	jsonLogger.Info("Batch processing",
		"batch_id", "batch-2025-001",
		"user.processor_id", "worker-05",
		"request.total_items", 1000,
		"request.processed", 750,
		"server.cpu_usage", 68.4,
		"server.memory_gb", 3.2,
		"status_code", 202,
	)

	// Different data types with colors
	jsonLogger.Debug("Type showcase",
		"string_value", "hello world",
		"integer_value", 42,
		"float_value", 3.14159,
		"boolean_true", true,
		"boolean_false", false,
		"null_value", nil,
		"user.metadata.tags", []string{"admin", "power-user"},
	)
}