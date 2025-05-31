package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// Create a JSON logger for better nested attribute visualization
	logger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.NewHandlerOptions().WithPrettyPrint(true),
	))

	// Using dot notation for nested attributes
	logger.WithDot("user.profile.name", "John Doe").
		WithDot("user.profile.email", "john@example.com").
		WithDot("user.preferences.theme", "dark").
		WithDot("user.preferences.notifications.email", true).
		Info("User authenticated with nested attributes")

	// Using key paths for nested attributes
	logger.WithNested([]string{"request", "headers", "user-agent"}, "Mozilla/5.0").
		WithNested([]string{"request", "headers", "content-type"}, "application/json").
		WithNested([]string{"response", "status"}, 200).
		WithNested([]string{"response", "timing", "total_ms"}, 150).
		Info("HTTP request processed")

	// Using groups for hierarchical organization
	userLogger := logger.WithGroup("user")
	userLogger.Info("User action", "action", "login", "user_id", 123, "session_id", "abc-123")
	
	// Nested groups
	requestLogger := userLogger.WithGroup("request")
	requestLogger.Info("API call", "method", "POST", "endpoint", "/api/users", "duration_ms", 45)

	// Complex nested structure with mixed approaches
	logger.WithDot("server.region", "us-east-1").
		WithDot("server.instance.id", "i-1234567890abcdef0").
		WithDot("server.instance.type", "t3.medium").
		WithDot("metrics.cpu.usage", 42.5).
		WithDot("metrics.memory.used_gb", 2.1).
		WithDot("metrics.memory.total_gb", 4.0).
		Info("Server metrics collected")
}