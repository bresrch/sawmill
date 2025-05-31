package main

import (
	"github.com/bresrch/sawmill"
)

func main() {
	// Basic logging with default text handler
	sawmill.Info("Application started", "version", "1.0.0", "env", "production")
	
	// Create a JSON logger
	jsonLogger := sawmill.New(sawmill.NewJSONHandlerWithDefaults())
	jsonLogger.Info("User authenticated", "user_id", 123, "ip", "192.168.1.1")
	
	// Different log levels
	logger := sawmill.New(sawmill.NewTextHandlerWithDefaults())
	logger.Trace("Detailed debug information")
	logger.Debug("Debug information")
	logger.Info("Informational message")
	logger.Warn("Warning message")
	logger.Error("Error occurred")
}