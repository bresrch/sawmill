package sawmill

import (
	"bytes"
	"strings"
	"testing"
)

// Test structs for masking functionality
type UserCredentials struct {
	Username string `sawmill:""`
	Password string `sawmill:"mask"`
	Email    string `sawmill:"mask[4]"`
	APIKey   string `sawmill:"mask[8]"`
	Token    string `sawmill:"mask[0]"`
	ID       int    `sawmill:"mask[2]"`
}

type NestedUser struct {
	Name        string
	Credentials UserCredentials
	SessionID   string `sawmill:"mask"`
}

func TestStructMasking(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	// Test basic masking
	user := UserCredentials{
		Username: "john_doe",
		Password: "secret123",
		Email:    "john@example.com",
		APIKey:   "abc123def456ghi789",
		Token:    "very_secret_token",
		ID:       12345,
	}

	logger.Info("User login", "user", user)
	output := buf.String()

	// Test that password is fully masked
	if strings.Contains(output, "secret123") {
		t.Errorf("Password should be masked, but found in output: %s", output)
	}
	if !strings.Contains(output, "\"user.password\":\"*********\"") {
		t.Errorf("Password should be masked with asterisks: %s", output)
	}

	// Test that email shows first 4 characters
	if !strings.Contains(output, "\"user.email\":\"john************\"") {
		t.Errorf("Email should show first 4 characters: %s", output)
	}

	// Test that API key shows first 8 characters
	if !strings.Contains(output, "\"user.apikey\":\"abc123de**********\"") {
		t.Errorf("API key should show first 8 characters: %s", output)
	}

	// Test that token is fully masked (mask[0])
	if strings.Contains(output, "very_secret_token") {
		t.Errorf("Token should be fully masked: %s", output)
	}
	if !strings.Contains(output, "\"user.token\":\"*****************\"") {
		t.Errorf("Token should be fully masked with asterisks: %s", output)
	}

	// Test that ID (integer) masking works
	if !strings.Contains(output, "\"user.id\":\"12***\"") {
		t.Errorf("ID should show first 2 digits: %s", output)
	}

	// Test that username is not masked (no mask tag)
	if !strings.Contains(output, "\"user.username\":\"john_doe\"") {
		t.Errorf("Username should not be masked: %s", output)
	}
}

func TestNestedStructMasking(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	nested := NestedUser{
		Name: "John Doe",
		Credentials: UserCredentials{
			Username: "admin",
			Password: "topsecret",
			Email:    "admin@company.com",
		},
		SessionID: "session123456",
	}

	logger.Info("User session", "user", nested)
	output := buf.String()

	// Test nested struct masking
	if strings.Contains(output, "topsecret") {
		t.Errorf("Nested password should be masked: %s", output)
	}
	if !strings.Contains(output, "\"user.credentials.password\":\"*********\"") {
		t.Errorf("Nested password should be masked with asterisks: %s", output)
	}

	// Test that top-level field with mask tag works
	if strings.Contains(output, "session123456") {
		t.Errorf("SessionID should be masked: %s", output)
	}
	if !strings.Contains(output, "\"user.sessionid\":\"*************\"") {
		t.Errorf("SessionID should be masked with asterisks: %s", output)
	}

	// Test that non-masked fields are preserved
	if !strings.Contains(output, "\"user.name\":\"John Doe\"") {
		t.Errorf("Name should not be masked: %s", output)
	}
	if !strings.Contains(output, "\"user.credentials.username\":\"admin\"") {
		t.Errorf("Nested username should not be masked: %s", output)
	}
}

func TestMaskingEdgeCases(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	type EdgeCase struct {
		EmptyString  string  `sawmill:"mask"`
		ShortString  string  `sawmill:"mask[5]"`
		ExactLength  string  `sawmill:"mask[3]"`
		LongerUnmask string  `sawmill:"mask[10]"`
		NilPointer   *string `sawmill:"mask"`
	}

	edge := EdgeCase{
		EmptyString:  "",
		ShortString:  "ab",
		ExactLength:  "abc",
		LongerUnmask: "short",
	}

	logger.Info("Edge cases", "edge", edge)
	output := buf.String()

	// Empty string should remain empty when masked
	if !strings.Contains(output, "\"edge.emptystring\":\"\"") {
		t.Errorf("Empty string should remain empty: %s", output)
	}

	// Short string with mask[5] should not be masked (unmask count >= length)
	if !strings.Contains(output, "\"edge.shortstring\":\"ab\"") {
		t.Errorf("Short string should not be masked when unmask count >= length: %s", output)
	}

	// Exact length should not be masked
	if !strings.Contains(output, "\"edge.exactlength\":\"abc\"") {
		t.Errorf("String with exact unmask length should not be masked: %s", output)
	}

	// Longer unmask than string length should not mask
	if !strings.Contains(output, "\"edge.longerunmask\":\"short\"") {
		t.Errorf("String with unmask count > length should not be masked: %s", output)
	}
}

func TestMaskValueFunction(t *testing.T) {
	attrs := NewFlatAttributes()

	// Test mask function directly
	tests := []struct {
		input    interface{}
		maskTag  string
		expected string
	}{
		{"password123", "mask", "***********"},
		{"email@test.com", "mask[5]", "email*********"},
		{"short", "mask[10]", "short"}, // unmask count >= length
		{12345, "mask[2]", "12***"},
		{"", "mask", ""},
		{"test", "mask[0]", "****"},
		{"invalid", "invalid_tag", "invalid"},
		{"test", "", "test"},
	}

	for i, test := range tests {
		result := attrs.maskValue(test.input, test.maskTag)
		if result != test.expected {
			t.Errorf("Test %d: maskValue(%v, %s) = %v, expected %v",
				i, test.input, test.maskTag, result, test.expected)
		}
	}
}

func TestMaskingWithDifferentFormatters(t *testing.T) {
	user := UserCredentials{
		Username: "testuser",
		Password: "secret",
		Email:    "test@example.com",
	}

	// Test with Text formatter
	buf := &bytes.Buffer{}
	textLogger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))
	textLogger.Info("Text test", "user", user)
	textOutput := buf.String()

	if strings.Contains(textOutput, "secret") {
		t.Errorf("Text formatter should mask password: %s", textOutput)
	}
	if !strings.Contains(textOutput, "password: ******") {
		t.Errorf("Text formatter should show masked password: %s", textOutput)
	}

	// Test with KeyValue formatter
	buf.Reset()
	kvLogger := New(NewKeyValueHandler(WithDestination(NewWriterDestination(buf))))
	kvLogger.Info("KeyValue test", "user", user)
	kvOutput := buf.String()

	if strings.Contains(kvOutput, "secret") {
		t.Errorf("KeyValue formatter should mask password: %s", kvOutput)
	}
}