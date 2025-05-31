package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// Text logger with colors for better mark visualization
	textLogger := sawmill.New(sawmill.NewTextHandler(
		sawmill.NewHandlerOptions().WithColorsEnabled(true),
	))
	
	// JSON logger
	jsonLogger := sawmill.New(sawmill.NewJSONHandlerWithDefaults())

	// === Text Format Marks ===
	
	// Simple mark without attributes
	textLogger.Mark("Application Startup")
	textLogger.Info("Loading configuration")
	textLogger.Info("Connecting to database")
	textLogger.Info("Starting HTTP server")
	
	// Mark with attributes
	textLogger.Mark("Authentication Phase", 
		"phase", "auth", 
		"step", 1, 
		"expected_users", 50,
	)
	textLogger.Info("Validating user credentials", "user_id", 123)
	textLogger.Info("Checking permissions", "role", "admin")
	textLogger.Info("Authentication successful", "session_id", "sess-abc-123")

	// Mark for error scenarios
	textLogger.Mark("Error Recovery Phase", 
		"phase", "recovery", 
		"error_count", 3,
		"retry_attempt", 2,
	)
	textLogger.Warn("Database connection lost")
	textLogger.Info("Attempting reconnection")
	textLogger.Info("Connection restored")

	// === JSON Format Marks ===
	
	jsonLogger.Mark("Data Processing Phase")
	jsonLogger.Info("Processing batch", "batch_id", "batch-001", "size", 1000)
	jsonLogger.Info("Validation complete", "valid_records", 950, "invalid_records", 50)
	
	jsonLogger.Mark("Batch Complete", 
		"batch_id", "batch-001",
		"total_processed", 1000,
		"success_rate", 95.0,
		"duration_seconds", 45,
	)

	// === Complex Workflow with Multiple Marks ===
	
	textLogger.Mark("Data Import Workflow", "workflow_id", "import-2025-001")
	
	textLogger.Mark("Stage 1: File Validation", "stage", 1)
	textLogger.Info("Checking file format")
	textLogger.Info("Validating schema")
	textLogger.Info("File validation passed")
	
	textLogger.Mark("Stage 2: Data Transformation", "stage", 2)
	textLogger.Info("Parsing CSV data")
	textLogger.Info("Applying business rules")
	textLogger.Info("Data transformation complete")
	
	textLogger.Mark("Stage 3: Database Import", "stage", 3)
	textLogger.Info("Starting transaction")
	textLogger.Info("Inserting records", "count", 5000)
	textLogger.Info("Committing transaction")
	
	textLogger.Mark("Workflow Complete", 
		"workflow_id", "import-2025-001",
		"total_records", 5000,
		"status", "success",
	)

	// === Nested Attributes in Marks ===
	
	textLogger.Mark("Performance Test Results",
		"test.name", "load_test_v1",
		"test.duration.seconds", 300,
		"metrics.requests.total", 150000,
		"metrics.requests.per_second", 500,
		"metrics.latency.p50_ms", 25,
		"metrics.latency.p95_ms", 85,
		"metrics.latency.p99_ms", 150,
		"metrics.errors.count", 12,
		"metrics.errors.rate", 0.008,
	)
}