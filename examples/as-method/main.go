package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// Create a logger with default text formatting
	logger := sawmill.Default()

	// Regular text formatted message
	logger.Info("This is a regular text message", "key", "value", "number", 42)

	// Temporarily switch to JSON format for this message only
	logger.As(sawmill.NewJSONFormatter()).Info("This message will be in JSON format", "user", "john", "action", "login")

	// Back to regular text format
	logger.Info("This is back to text format", "status", "success")

	// Temporarily switch to key-value format
	logger.As(sawmill.NewKeyValueFormatter()).Info("This is key-value format", "service", "api", "response_time", 150)

	// Temporarily switch to XML format
	logger.As(sawmill.NewXMLFormatter()).Info("This is XML format", "category", "system")

	// Use the same AsLogger instance for multiple messages with the same output ID
	asLogger := logger.As(sawmill.NewJSONFormatter())
	asLogger.Info("First JSON message in sequence", "sequence", 1)
	asLogger.Info("Second JSON message in sequence", "sequence", 2)

	// Final message back to regular format
	logger.Info("Final message in regular format")
}
