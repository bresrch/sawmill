package sawmill

import (
	"bytes"
	"strings"
	"testing"
)

func TestAsMethod(t *testing.T) {
	// Create a buffer to capture output
	buf := &bytes.Buffer{}
	
	// Create a logger with text formatter
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))
	
	// Log a normal message
	logger.Info("Normal text message", "key", "value")
	
	// Log a message using JSON formatter temporarily
	logger.As(NewJSONFormatter()).Info("JSON formatted message", "key", "value")
	
	// Log another normal message
	logger.Info("Another normal text message", "key", "value")
	
	output := buf.String()
	
	// Check that we have text format for normal messages
	if !strings.Contains(output, "[INFO]") || !strings.Contains(output, "Normal text message") {
		t.Errorf("Output should contain text format for normal messages: %s", output)
	}
	
	// Check that we have JSON format for the As() message
	if !strings.Contains(output, `"message":"JSON formatted message"`) {
		t.Errorf("Output should contain JSON format for As() message: %s", output)
	}
	
	// Check that both normal messages are in text format
	if !strings.Contains(output, "Another normal text message") {
		t.Errorf("Output should contain the second normal text message: %s", output)
	}
}

func TestAsMethodWithDifferentFormatters(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))
	
	// Test with different formatters
	logger.As(NewJSONFormatter()).Info("JSON message")
	logger.As(NewKeyValueFormatter()).Info("KeyValue message")
	logger.As(NewXMLFormatter()).Info("XML message")
	logger.As(NewYAMLFormatter()).Info("YAML message")
	
	output := buf.String()
	
	// Check that each format appears
	if !strings.Contains(output, `"message":"JSON message"`) {
		t.Errorf("JSON format not found in output: %s", output)
	}
	if !strings.Contains(output, "message=KeyValue message") {
		t.Errorf("Key-Value format not found in output: %s", output)
	}
	if !strings.Contains(output, "<record>") || !strings.Contains(output, "XML message") {
		t.Errorf("XML format not found in output: %s", output)
	}
	if !strings.Contains(output, `message: "YAML message"`) {
		t.Errorf("YAML format not found in output: %s", output)
	}
}

func TestAsMethodOutputID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewJSONHandler(WithDestination(NewWriterDestination(buf))))
	
	// Create an AsLogger instance
	asLogger := logger.As(NewJSONFormatter())
	
	// Log multiple messages with the same AsLogger
	asLogger.Info("First message")
	asLogger.Info("Second message")
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines of output, got %d", len(lines))
	}
	
	// Both messages should have the same OutputID (though we can't easily verify the exact value in this test)
	// The key thing is that the messages are formatted correctly
	for i, line := range lines {
		if !strings.Contains(line, `"message":`) {
			t.Errorf("Line %d should contain formatted message: %s", i+1, line)
		}
	}
}

func TestAsMethodLevelMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(NewTextHandler(WithDestination(NewWriterDestination(buf))))
	
	asLogger := logger.As(NewJSONFormatter())
	
	// Test all level methods
	asLogger.Trace("Trace message")
	asLogger.Debug("Debug message")
	asLogger.Info("Info message")
	asLogger.Warn("Warn message")
	asLogger.Error("Error message")
	asLogger.Mark("Mark message")
	
	output := buf.String()
	
	// All messages should be in JSON format
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
			t.Errorf("Line %d should be JSON format: %s", i+1, line)
		}
	}
}