package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// === Basic Key-Value Format ===
	
	// Simple key-value logger without colors
	kvLogger := sawmill.New(sawmill.NewKeyValueHandlerWithDefaults())
	
	kvLogger.Info("User authentication",
		"user_id", 12345,
		"username", "alice_doe",
		"login_successful", true,
		"login_time", "2025-05-31T14:30:00Z",
		"ip_address", "192.168.1.100",
	)

	kvLogger.Warn("Rate limit approaching",
		"user_id", 12345,
		"current_requests", 95,
		"limit", 100,
		"window_minutes", 15,
		"reset_time", "2025-05-31T14:45:00Z",
	)

	// === Colored Key-Value Format ===
	
	// Custom color mappings for different key patterns
	colorMappings := map[string]string{
		"user":        sawmill.ColorBrightBlue,   // All user.* keys
		"request":     sawmill.ColorBrightGreen,  // All request.* keys
		"response":    sawmill.ColorBrightCyan,   // All response.* keys
		"server":      sawmill.ColorYellow,       // All server.* keys
		"error":       sawmill.ColorBrightRed,    // All error.* keys
		"db":          sawmill.ColorMagenta,      // All db.* keys
		"metric":      sawmill.ColorBrightWhite,  // All metric.* keys
		"status_code": sawmill.ColorCyan,         // Specific key
		"duration_ms": sawmill.ColorGreen,        // Specific key
	}

	coloredKVLogger := sawmill.New(sawmill.NewKeyValueHandler(
		sawmill.NewHandlerOptions().
			WithColorsEnabled(true).
			WithColorMappings(colorMappings),
	))

	// Demonstrate colored key-value output with nested keys
	coloredKVLogger.Info("HTTP request processed",
		"request.method", "POST",
		"request.path", "/api/users",
		"request.headers.content_type", "application/json",
		"request.body_size_bytes", 256,
		"response.status_code", 201,
		"response.body_size_bytes", 512,
		"response.headers.location", "/api/users/12345",
		"user.id", 12345,
		"user.role", "admin",
		"server.hostname", "api-01",
		"server.instance_id", "i-0123456789abcdef0",
		"duration_ms", 45,
		"success", true,
	)

	// === Error Scenarios with Key-Value ===
	
	coloredKVLogger.Error("Database connection failed",
		"error.code", "DB_CONNECTION_TIMEOUT",
		"error.message", "Connection timed out after 5000ms",
		"error.stack_trace", "db.Connect() at line 42",
		"db.host", "postgres-primary.internal",
		"db.port", 5432,
		"db.database", "user_service",
		"db.connection_pool.size", 10,
		"db.connection_pool.active", 8,
		"db.connection_pool.idle", 2,
		"retry_count", 3,
		"will_retry", true,
		"next_retry_seconds", 30,
	)

	// === Performance Metrics with Key-Value ===
	
	coloredKVLogger.Info("Performance metrics collected",
		"metric.cpu.usage_percent", 68.5,
		"metric.memory.used_gb", 3.2,
		"metric.memory.total_gb", 8.0,
		"metric.memory.usage_percent", 40.0,
		"metric.disk.used_gb", 45.2,
		"metric.disk.total_gb", 100.0,
		"metric.network.bytes_in", 1024000,
		"metric.network.bytes_out", 2048000,
		"server.uptime_hours", 72,
		"server.load_average_1m", 1.25,
		"server.load_average_5m", 1.15,
		"server.load_average_15m", 1.05,
	)

	// === Batch Processing with Key-Value ===
	
	coloredKVLogger.Info("Batch processing started",
		"batch.id", "batch-2025-05-31-001",
		"batch.type", "user_data_export",
		"batch.source.database", "analytics",
		"batch.source.table", "user_events",
		"batch.filter.date_start", "2025-05-01",
		"batch.filter.date_end", "2025-05-31",
		"batch.expected_records", 50000,
		"user.processor_id", "worker-05",
		"server.queue.pending", 3,
		"server.queue.processing", 1,
	)

	// === Mark Function with Key-Value Format ===
	
	coloredKVLogger.Mark("Data Processing Phase Started",
		"phase", "data_processing",
		"workflow_id", "wf-2025-001",
		"step", 2,
		"total_steps", 5,
		"estimated_duration_minutes", 15,
	)

	coloredKVLogger.Info("Processing chunk", "chunk", 1, "size", 1000)
	coloredKVLogger.Info("Processing chunk", "chunk", 2, "size", 1000)
	coloredKVLogger.Info("Processing chunk", "chunk", 3, "size", 1000)

	coloredKVLogger.Mark("Data Processing Phase Complete",
		"phase", "data_processing",
		"workflow_id", "wf-2025-001",
		"step", 2,
		"status", "success",
		"actual_duration_minutes", 12,
		"records_processed", 3000,
	)

	// === Complex Nested Data with Key-Value ===
	
	coloredKVLogger.Info("API gateway metrics",
		"request.api.version", "v2",
		"request.api.endpoint", "/api/v2/users/search",
		"request.client.id", "client-abc-123",
		"request.client.version", "1.5.2",
		"request.auth.type", "bearer_token",
		"request.auth.user_id", 67890,
		"response.cache.hit", false,
		"response.cache.ttl_seconds", 300,
		"response.pagination.page", 1,
		"response.pagination.per_page", 25,
		"response.pagination.total", 1250,
		"response.data.count", 25,
		"server.region", "us-east-1",
		"server.az", "us-east-1a",
		"duration_ms", 125,
		"status_code", 200,
	)

	// === Different Data Types Showcase ===
	
	coloredKVLogger.Debug("Data types showcase",
		"string_value", "hello world",
		"integer_positive", 42,
		"integer_negative", -17,
		"float_positive", 3.14159,
		"float_negative", -2.71828,
		"boolean_true", true,
		"boolean_false", false,
		"null_value", nil,
		"user.preferences.theme", "dark",
		"user.preferences.language", "en-US",
		"user.settings.notifications", true,
		"user.settings.beta_features", false,
	)
}