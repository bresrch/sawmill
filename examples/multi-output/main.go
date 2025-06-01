package main

import (
	"os"

	"github.com/bresrch/sawmill"
)

func main() {
	// === Multi-Handler Setup ===

	// Create multiple handlers with different configurations
	multiHandler := sawmill.NewMultiHandler(
		// Console output with colors (JSON)
		sawmill.NewJSONHandler(
			sawmill.WithStdout(),
			sawmill.WithColorsEnabled(true),
			sawmill.WithPrettyPrint(true),
			sawmill.WithColorMappings(map[string]string{
				"user":   sawmill.ColorBrightBlue,
				"system": sawmill.ColorYellow,
				"error":  sawmill.ColorBrightRed,
			}),
		),

		// File output (Text format for readability)
		sawmill.NewTextHandler(
			sawmill.WithFile("/tmp/sawmill-multi.log", 50*1024*1024, false), // 50MB, no compression
			sawmill.WithTimeFormat("2006-01-02 15:04:05.000"),
			sawmill.WithAttributeFormat("flat"),
		),

		// Key-value format to stdout (for monitoring tools)
		sawmill.NewKeyValueHandler(
			sawmill.WithStdout(),
			sawmill.WithColorsEnabled(false), // No colors for machine parsing
		),
	)

	logger := sawmill.New(multiHandler)

	// === Log Events to All Handlers ===

	logger.Info("Application startup",
		"system.version", "v1.2.3",
		"system.environment", "production",
		"system.pid", os.Getpid(),
		"system.startup_time_ms", 1250,
	)

	logger.Info("User session started",
		"user.id", 12345,
		"user.username", "alice_smith",
		"user.role", "admin",
		"user.session.id", "sess-abc-123-def-456",
		"user.session.timeout_minutes", 60,
		"system.load_balancer", "lb-01",
	)

	logger.Warn("High memory usage detected",
		"system.memory.used_gb", 7.2,
		"system.memory.total_gb", 8.0,
		"system.memory.usage_percent", 90.0,
		"system.cpu.usage_percent", 75.5,
		"system.instance.id", "i-0123456789abcdef0",
	)

	logger.Error("Database query timeout",
		"error.type", "QueryTimeoutException",
		"error.message", "Query execution exceeded 30 seconds",
		"error.query.table", "user_events",
		"error.query.duration_ms", 30000,
		"error.query.timeout_ms", 30000,
		"user.id", 67890,
		"system.database.host", "postgres-cluster.internal",
		"system.database.replica", "read-02",
	)

	// === Format-Specific Multi-Output ===

	// Create different handlers for different purposes
	debugHandler := sawmill.NewMultiHandler(
		// Detailed console output for development
		sawmill.NewTextHandler(
			sawmill.WithStdout(),
			sawmill.WithLevel(sawmill.LevelDebug),
			sawmill.WithColorsEnabled(true),
			sawmill.WithAttributeFormat("nested"),
			sawmill.WithSourceInfo(true),
		),

		// Machine-readable format for log aggregation
		sawmill.NewJSONHandler(
			sawmill.WithFile("/tmp/debug-structured.jsonl", 100*1024*1024, true),
			sawmill.WithLevel(sawmill.LevelDebug),
			sawmill.WithPrettyPrint(false), // Compact for storage efficiency
		),
	)

	debugLogger := sawmill.New(debugHandler)

	debugLogger.Debug("Detailed debug information",
		"function", "processUserRequest",
		"user.request.headers.authorization", "Bearer ***",
		"user.request.body.fields", []string{"name", "email", "preferences"},
		"system.cache.hit_rate", 0.85,
		"system.cache.keys_count", 15420,
		"system.performance.gc_pause_ms", 2.5,
	)

	// === Conditional Multi-Output ===

	// Different handlers based on log level
	productionHandler := sawmill.NewMultiHandler(
		// Console for immediate feedback (only warnings and errors)
		sawmill.NewTextHandler(
			sawmill.WithStdout(),
			sawmill.WithLevel(sawmill.LevelWarn),
			sawmill.WithColorsEnabled(true),
		),

		// File for all logs (info and above)
		sawmill.NewJSONHandler(
			sawmill.WithFile("/tmp/production.log", 200*1024*1024, true),
			sawmill.WithLevel(sawmill.LevelInfo),
			sawmill.WithTimeFormat("2006-01-02T15:04:05.000Z07:00"),
		),

		// Error-only file for critical issues
		sawmill.NewTextHandler(
			sawmill.WithFile("/tmp/errors.log", 50*1024*1024, false),
			sawmill.WithLevel(sawmill.LevelError),
			sawmill.WithTimeFormat("2006-01-02 15:04:05"),
		),
	)

	prodLogger := sawmill.New(productionHandler)

	// These will go to different outputs based on level
	prodLogger.Info("User registration completed") // File only
	prodLogger.Warn("API rate limit exceeded")     // Console + File
	prodLogger.Error("Payment processing failed")  // Console + File + Error file

	// === Multi-Format Comparison ===

	// Same data, multiple formats for comparison
	comparisonHandler := sawmill.NewMultiHandler(
		sawmill.NewTextHandlerWithDefaults(),
		sawmill.NewJSONHandlerWithDefaults(),
		sawmill.NewKeyValueHandlerWithDefaults(),
		sawmill.NewXMLHandlerWithDefaults(),
		sawmill.NewYAMLHandlerWithDefaults(),
	)

	comparisonLogger := sawmill.New(comparisonHandler)

	comparisonLogger.Info("Multi-format output test",
		"transaction.id", "txn-789-xyz",
		"transaction.amount", 99.99,
		"transaction.currency", "USD",
		"user.account.number", "acc-123456",
		"system.processor.id", "stripe",
		"system.region", "us-west-2",
	)

	// === Mark with Multi-Output ===

	logger.Mark("Processing Phase Complete",
		"phase", "data_validation",
		"records.total", 10000,
		"records.valid", 9850,
		"records.invalid", 150,
		"duration_seconds", 125,
		"throughput_records_per_second", 80,
	)
}
