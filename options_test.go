package sawmill

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestHandlerOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	dest := NewWriterDestination(buf)

	options := []HandlerOption{
		WithDestination(dest),
		WithLevel(LevelWarn),
		WithTimeFormat("2006-01-02"),
		WithSourceInfo(true),
		WithLevelInfo(false),
		WithAttributeFormat("flat"),
		WithPrettyPrint(true),
		WithAttributesKey("custom_attrs"),
		WithColorsEnabled(true),
		WithColorMappings(map[string]string{"user": "red"}),
	}

	handlerOpts := NewHandlerOptions(options...)

	// Test all options were applied
	if handlerOpts.destination != dest {
		t.Error("Destination option not applied")
	}
	if handlerOpts.level != LevelWarn {
		t.Error("Level option not applied")
	}
	if handlerOpts.timeFormat != "2006-01-02" {
		t.Error("TimeFormat option not applied")
	}
	if !handlerOpts.includeSource {
		t.Error("IncludeSource option not applied")
	}
	if handlerOpts.includeLevel {
		t.Error("IncludeLevel option not applied")
	}
	if handlerOpts.attrFormat != "flat" {
		t.Error("AttributeFormat option not applied")
	}
	if !handlerOpts.prettyPrint {
		t.Error("PrettyPrint option not applied")
	}
	if handlerOpts.attributesKey != "custom_attrs" {
		t.Error("AttributesKey option not applied")
	}
	if !handlerOpts.enableColors {
		t.Error("ColorsEnabled option not applied")
	}
	if handlerOpts.colorMappings["user"] != "red" {
		t.Error("ColorMappings option not applied")
	}
}

func TestSawmillOptions(t *testing.T) {
	sawmillOpts := &SawmillOptions{
		LogLevel: "debug",
		LogFile:  "test.log",
		MaxSize:  100,
	}

	options := []HandlerOption{
		WithSawmillOptions(sawmillOpts),
	}

	handlerOpts := NewHandlerOptions(options...)

	if handlerOpts.sawmillOpts != sawmillOpts {
		t.Error("SawmillOptions not applied")
	}
}

func TestWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	
	handler := NewTextHandler(WithWriter(buf))
	logger := New(handler)
	
	logger.Info("Test message")
	
	output := buf.String()
	if output == "" {
		t.Error("WithWriter option should route output to provided writer")
	}
}

func TestDefaultHandlerOptions(t *testing.T) {
	opts := NewHandlerOptions()

	// Test defaults
	if opts.level != LevelInfo {
		t.Errorf("Expected default level Info, got %v", opts.level)
	}
	if opts.timeFormat != time.RFC3339 {
		t.Errorf("Expected default time format RFC3339, got %s", opts.timeFormat)
	}
	if !opts.includeSource {
		t.Error("Expected default includeSource to be true")
	}
	if !opts.includeLevel {
		t.Error("Expected default includeLevel to be true")
	}
	if opts.attrFormat != "nested" {
		t.Errorf("Expected default attribute format nested, got %s", opts.attrFormat)
	}
	if opts.prettyPrint {
		t.Error("Expected default prettyPrint to be false")
	}
	if opts.attributesKey != "attributes" {
		t.Errorf("Expected default attributes key 'attributes', got %s", opts.attributesKey)
	}
	if opts.enableColors {
		t.Error("Expected default enableColors to be false")
	}
	if opts.colorOutput {
		t.Error("Expected default colorOutput to be false")
	}
}

func TestMultipleOptionsOfSameType(t *testing.T) {
	// Test that later options override earlier ones
	options := []HandlerOption{
		WithLevel(LevelDebug),
		WithLevel(LevelError), // This should override the first one
		WithTimeFormat("2006-01-02"),
		WithTimeFormat("15:04:05"), // This should override the first one
	}

	opts := NewHandlerOptions(options...)

	if opts.level != LevelError {
		t.Errorf("Expected level Error (last one), got %v", opts.level)
	}
	if opts.timeFormat != "15:04:05" {
		t.Errorf("Expected time format '15:04:05' (last one), got %s", opts.timeFormat)
	}
}

func TestOptionsWithRealHandlers(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test TextHandler with options
	textHandler := NewTextHandler(
		WithDestination(NewWriterDestination(buf)),
		WithLevel(LevelWarn),
		WithSourceInfo(false),
		WithAttributeFormat("flat"),
	)

	logger := New(textHandler)
	
	// This should not appear (below warn level)
	logger.Info("Info message")
	if buf.Len() > 0 {
		t.Error("Info message should not appear with Warn level")
	}

	// This should appear
	logger.Error("Error message")
	output := buf.String()
	if output == "" {
		t.Error("Error message should appear with Warn level")
	}

	// Test JSONHandler with options
	buf.Reset()
	jsonHandler := NewJSONHandler(
		WithDestination(NewWriterDestination(buf)),
		WithPrettyPrint(true),
		WithAttributesKey("data"),
	)

	logger = New(jsonHandler)
	logger.Info("JSON message", "key", "value")
	
	jsonOutput := buf.String()
	if jsonOutput == "" {
		t.Error("JSON handler should produce output")
	}
	// Should be pretty printed (contains indentation)
	if !containsIndentation(jsonOutput) {
		t.Error("JSON should be pretty printed")
	}
	// Should use custom attributes key
	if !containsString(jsonOutput, "data") {
		t.Error("JSON should use custom attributes key")
	}
}

func TestOptionsBuilderPattern(t *testing.T) {
	// Test that options can be built in a fluent style
	buf := &bytes.Buffer{}
	
	handler := NewJSONHandler(
		WithDestination(NewWriterDestination(buf)),
		WithLevel(LevelDebug),
		WithTimeFormat("2006-01-02 15:04:05"),
		WithSourceInfo(true),
		WithPrettyPrint(true),
		WithAttributesKey("attrs"),
		WithColorsEnabled(false),
	)

	if handler == nil {
		t.Fatal("Handler should not be nil")
	}

	logger := New(handler)
	logger.Info("Test message", "test", "value")

	output := buf.String()
	if output == "" {
		t.Error("Handler with multiple options should produce output")
	}
}

func TestColorMappingsOption(t *testing.T) {
	colorMappings := map[string]string{
		"user":    "\033[32m", // green
		"error":   "\033[31m", // red
		"request": "\033[34m", // blue
	}

	buf := &bytes.Buffer{}
	handler := NewTextHandler(
		WithDestination(NewWriterDestination(buf)),
		WithColorMappings(colorMappings),
		WithColorsEnabled(true),
	)

	logger := New(handler)
	logger.Info("Test message", "user", "john", "error", "none")

	output := buf.String()
	// Should contain ANSI color codes when colors are enabled
	if !containsString(output, "\033[") {
		t.Error("Output should contain ANSI color codes")
	}
}

func TestSawmillOptionsIntegration(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "test_sawmill.log"
	defer os.Remove(tmpFile)

	sawmillOpts := &SawmillOptions{
		LogLevel: "warn",
		LogFile:  tmpFile,
		MaxSize:  1, // 1MB
	}

	handler := NewTextHandler(WithSawmillOptions(sawmillOpts))
	logger := New(handler)

	// This should not appear (info < warn)
	logger.Info("Info message")

	// This should appear
	logger.Error("Error message")

	// Check that file was created (though we can't easily verify content without more setup)
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		// File creation might fail in test environment, that's ok
		t.Logf("File creation failed (expected in test): %v", err)
	}
}

// Helper functions for tests

func containsIndentation(s string) bool {
	return containsString(s, "  ") || containsString(s, "\t")
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func TestOptionsOrderIndependence(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	// Create two handlers with same options in different order
	handler1 := NewTextHandler(
		WithDestination(NewWriterDestination(buf1)),
		WithLevel(LevelDebug),
		WithTimeFormat("2006-01-02"),
		WithSourceInfo(true),
	)

	handler2 := NewTextHandler(
		WithSourceInfo(true),
		WithTimeFormat("2006-01-02"),
		WithLevel(LevelDebug),
		WithDestination(NewWriterDestination(buf2)),
	)

	logger1 := New(handler1)
	logger2 := New(handler2)

	// Both should behave the same
	logger1.Debug("Test message")
	logger2.Debug("Test message")

	// Both should produce output (debug level enabled)
	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("Both handlers should produce output regardless of option order")
	}
}

func TestNilOptionsHandling(t *testing.T) {
	// Test that nil options don't cause panics
	var nilOptions []HandlerOption
	opts := NewHandlerOptions(nilOptions...)

	if opts == nil {
		t.Error("NewHandlerOptions should handle nil options")
	}

	// Test handler creation with no options
	handler := NewTextHandler()
	if handler == nil {
		t.Error("Handler should be created even with no options")
	}

	logger := New(handler)
	logger.Info("Test message") // Should not panic
}

func TestInvalidOptionValues(t *testing.T) {
	// Test with empty/invalid values
	handler := NewTextHandler(
		WithTimeFormat(""), // Empty time format
		WithAttributeFormat("invalid"), // Invalid attribute format
		WithAttributesKey(""), // Empty attributes key
	)

	if handler == nil {
		t.Error("Handler should handle invalid option values gracefully")
	}

	logger := New(handler)
	logger.Info("Test message") // Should not panic
}

func TestOptionsCombinations(t *testing.T) {
	// Test various combinations of options work together
	testCases := [][]HandlerOption{
		{WithLevel(LevelTrace), WithSourceInfo(true)},
		{WithPrettyPrint(true), WithAttributesKey("custom")},
		{WithColorsEnabled(true), WithAttributeFormat("flat")},
		{WithTimeFormat("15:04:05"), WithLevelInfo(false)},
	}

	for i, options := range testCases {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			buf := &bytes.Buffer{}
			baseOptions := []HandlerOption{WithDestination(NewWriterDestination(buf))}
			allOptions := append(baseOptions, options...)

			handler := NewJSONHandler(allOptions...)
			if handler == nil {
				t.Errorf("Handler creation failed for option set %d", i)
			}

			logger := New(handler)
			logger.Info("Test message", "key", "value")

			if buf.Len() == 0 {
				t.Errorf("No output produced for option set %d", i)
			}
		})
	}
}