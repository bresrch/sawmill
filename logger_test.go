package sawmill

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewTextHandler(WithDestination(NewWriterDestination(buf)))
	logger := New(handler)

	if logger == nil {
		t.Fatal("New() returned nil logger")
	}

	logger.Info("Test message")
	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected output to contain test message: %s", output)
	}
}

func TestDefaultLogger(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Fatal("Default() returned nil logger")
	}
}

func TestLoggerLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf)), WithLevel(LevelTrace)))

	tests := []struct {
		level    string
		logFunc  func(string, ...interface{})
		expected string
	}{
		{"TRACE", logger.Trace, "TRACE"},
		{"DEBUG", logger.Debug, "DEBUG"},
		{"INFO", logger.Info, "INFO"},
		{"WARN", logger.Warn, "WARN"},
		{"ERROR", logger.Error, "ERROR"},
		{"FATAL", logger.Fatal, "FATAL"},
		{"MARK", logger.Mark, "MARK"},
	}

	for _, test := range tests {
		buf.Reset()
		test.logFunc("Test "+test.level+" message")
		output := buf.String()
		
		if !strings.Contains(output, test.expected) {
			t.Errorf("Expected %s level output to contain '%s': %s", test.level, test.expected, output)
		}
		// MARK level has special formatting, so only check message for other levels
		if test.level != "MARK" && !strings.Contains(output, "Test "+test.level+" message") {
			t.Errorf("Expected output to contain message: %s", output)
		}
		// For MARK level, just check that it contains MARKED (special format)
		if test.level == "MARK" && !strings.Contains(output, "MARKED") {
			t.Errorf("Expected MARK output to contain MARKED: %s", output)
		}
	}
}

func TestLoggerPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Panic to panic")
		}
	}()

	logger.Panic("Test panic message")
}

func TestLoggerWithNested(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	nestedLogger := logger.WithNested([]string{"user", "profile"}, "john")
	nestedLogger.Info("User info")

	output := buf.String()
	if !strings.Contains(output, "\"user.profile\":\"john\"") {
		t.Errorf("Expected nested attribute in output: %s", output)
	}
}

func TestLoggerWithDot(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	dotLogger := logger.WithDot("user.email", "john@example.com")
	dotLogger.Info("User info")

	output := buf.String()
	if !strings.Contains(output, "\"user.email\":\"john@example.com\"") {
		t.Errorf("Expected dot notation attribute in output: %s", output)
	}
}

func TestLoggerWithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	groupLogger := logger.WithGroup("request")
	groupLogger.Info("Request processed", "method", "GET", "path", "/api/users")

	output := buf.String()
	if !strings.Contains(output, "\"request.method\":\"GET\"") || !strings.Contains(output, "\"request.path\":\"/api/users\"") {
		t.Errorf("Expected grouped attributes in output: %s", output)
	}
}

func TestLoggerWithCallback(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))

	callbackLogger := logger.WithCallback(func(record *Record) *Record {
		record.Message = "[MODIFIED] " + record.Message
		return record
	})

	callbackLogger.Info("Original message")
	output := buf.String()

	if !strings.Contains(output, "[MODIFIED] Original message") {
		t.Errorf("Expected callback to modify message: %s", output)
	}
}

func TestLoggerSetHandler(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	handler1 := NewTextHandler(WithDestination(NewWriterDestination(buf1)))
	handler2 := NewJSONHandler(WithDestination(NewWriterDestination(buf2)))

	logger := New(handler1)
	logger.Info("Text message")

	if buf1.Len() == 0 {
		t.Error("Expected text handler to write output")
	}

	logger.SetHandler(handler2)
	logger.Info("JSON message")

	if buf2.Len() == 0 {
		t.Error("Expected JSON handler to write output after SetHandler")
	}
}

func TestLoggerHandler(t *testing.T) {
	handler := NewTextHandler()
	logger := New(handler)

	if logger.Handler() != handler {
		t.Error("Handler() should return the same handler that was set")
	}
}

func TestLoggerWithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(WithDestination(NewWriterDestination(buf)))

	attrs := []slog.Attr{
		slog.String("service", "api"),
		slog.Int("port", 8080),
	}

	attrHandler := handler.WithAttrs(attrs)
	logger := New(attrHandler)
	logger.Info("Service started")

	output := buf.String()
	if !strings.Contains(output, "\"service\":\"api\"") || !strings.Contains(output, "\"port\":8080") {
		t.Errorf("Expected slog attributes in output: %s", output)
	}
}

func TestLoggerLogRecord(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))

	record := NewRecordFromPool(LevelInfo, "Test record message")
	record.Attributes.SetFast("key", "value")

	logger.LogRecord(context.Background(), record)
	
	output := buf.String()
	if !strings.Contains(output, "Test record message") {
		t.Errorf("Expected record message in output: %s", output)
	}
	if !strings.Contains(output, "key: value") {
		t.Errorf("Expected record attributes in output: %s", output)
	}
}

func TestLoggerStructExpansion(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: 123, Name: "John Doe"}
	logger.Info("User created", "user", user)

	output := buf.String()
	if !strings.Contains(output, "\"user.id\":123") || !strings.Contains(output, "\"user.name\":\"John Doe\"") {
		t.Errorf("Expected struct expansion in output: %s", output)
	}
}

func TestLoggerHTTPErrorLog(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))

	httpLogger := logger.HTTPErrorLog()
	if httpLogger == nil {
		t.Fatal("HTTPErrorLog() returned nil")
	}

	httpLogger.Println("HTTP error message")
	output := buf.String()

	if !strings.Contains(output, "HTTP error message") {
		t.Errorf("Expected HTTP error message in output: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected ERROR level in output: %s", output)
	}
}

func TestNeedsSourceCapture(t *testing.T) {
	tests := []struct {
		name     string
		handler  Handler
		expected bool
	}{
		{
			name:     "TextHandler with source enabled",
			handler:  NewTextHandler(WithSourceInfo(true)),
			expected: true,
		},
		{
			name:     "TextHandler with source disabled", 
			handler:  NewTextHandler(WithSourceInfo(false)),
			expected: false,
		},
		{
			name:     "JSONHandler with source enabled",
			handler:  NewJSONHandler(WithSourceInfo(true)),
			expected: true,
		},
		{
			name:     "JSONHandler with source disabled",
			handler:  NewJSONHandler(WithSourceInfo(false)),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := New(test.handler).(*logger)
			if got := logger.needsSourceCapture(); got != test.expected {
				t.Errorf("needsSourceCapture() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Test that global functions work without panicking
	Trace("Test trace")
	Debug("Test debug")  
	Info("Test info")
	Warn("Test warn")
	Error("Test error")
	Mark("Test mark")

	// Test global With functions
	WithNested([]string{"test"}, "value").Info("Nested test")
	WithDot("test.dot", "value").Info("Dot test")
	WithGroup("testgroup").Info("Group test")
	WithCallback(func(r *Record) *Record { return r }).Info("Callback test")
}

func TestSetDefaultHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewTextHandler(WithDestination(NewWriterDestination(buf)))
	
	SetDefaultHandler(handler)
	Info("Test message with new handler")
	
	output := buf.String()
	if !strings.Contains(output, "Test message with new handler") {
		t.Errorf("Expected message in output after SetDefaultHandler: %s", output)
	}
}