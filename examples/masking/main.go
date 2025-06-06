package main

import (
	"github.com/bresrch/sawmill"
)

// Example structs demonstrating masking functionality
type User struct {
	Name      string `sawmill:""`           // No masking
	Email     string `sawmill:"mask[3]"`    // Show first 3 characters
	Password  string `sawmill:"mask"`       // Fully masked
	APIKey    string `sawmill:"mask[8]"`    // Show first 8 characters
	Token     string `sawmill:"mask[0]"`    // Fully masked (equivalent to "mask")
	ID        int    `sawmill:"mask[2]"`    // Show first 2 digits
}

type Session struct {
	UserID    int    `sawmill:"mask[1]"`
	SessionID string `sawmill:"mask"`
	User      User   // Nested struct with its own masking rules
}

func main() {
	// Create a logger with JSON output for clear visibility
	logger := sawmill.New(sawmill.NewJSONHandler(
		sawmill.WithPrettyPrint(true),
	))

	// Example user data
	user := User{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "super_secret_password",
		APIKey:   "sk_live_abc123def456ghi789",
		Token:    "very_secret_token_xyz",
		ID:       12345,
	}

	// Log user information - sensitive fields will be masked
	logger.Info("User logged in", "user", user)

	// Example with nested structs
	session := Session{
		UserID:    67890,
		SessionID: "session_abc123xyz789",
		User:      user,
	}

	logger.Info("Session created", "session", session)

	// Example with text formatter
	textLogger := sawmill.New(sawmill.NewTextHandler())
	textLogger.Info("User details (text format)", "user", user)

	// Example with key-value formatter
	kvLogger := sawmill.New(sawmill.NewKeyValueHandler())
	kvLogger.Info("User details (key-value format)", "user", user)
}