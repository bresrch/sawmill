package sawmill

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTextHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewTextHandler(WithDestination(NewWriterDestination(buf)))

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected message in output: %s", output)
	}
	if !strings.Contains(output, "key: value") {
		t.Errorf("Expected attributes in output: %s", output)
	}
}

func TestJSONHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(WithDestination(NewWriterDestination(buf)))

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Errorf("Expected JSON message in output: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected JSON attributes in output: %s", output)
	}
}

func TestXMLHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewXMLHandler(WithDestination(NewWriterDestination(buf)))

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<record>") {
		t.Errorf("Expected XML record tag in output: %s", output)
	}
	if !strings.Contains(output, "<message>Test message</message>") {
		t.Errorf("Expected XML message in output: %s", output)
	}
}

func TestYAMLHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewYAMLHandler(WithDestination(NewWriterDestination(buf)))

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `message: "Test message"`) {
		t.Errorf("Expected YAML message in output: %s", output)
	}
	if !strings.Contains(output, "key: value") {
		t.Errorf("Expected YAML attributes in output: %s", output)
	}
}

func TestKeyValueHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewKeyValueHandler(WithDestination(NewWriterDestination(buf)))

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "message=Test message") {
		t.Errorf("Expected key-value message in output: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected key-value attributes in output: %s", output)
	}
}

func TestMultiHandler(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	handler1 := NewTextHandler(WithDestination(NewWriterDestination(buf1)))
	handler2 := NewJSONHandler(WithDestination(NewWriterDestination(buf2)))
	multiHandler := NewMultiHandler(handler1, handler2)

	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	err := multiHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("MultiHandler Handle failed: %v", err)
	}

	output1 := buf1.String()
	output2 := buf2.String()

	if !strings.Contains(output1, "Test message") {
		t.Errorf("Expected text output in first handler: %s", output1)
	}
	if !strings.Contains(output2, `"message":"Test message"`) {
		t.Errorf("Expected JSON output in second handler: %s", output2)
	}
}

func TestHandlerEnabled(t *testing.T) {
	tests := []struct {
		name        string
		handlerLevel Level
		logLevel     Level
		expected     bool
	}{
		{"Info handler with Info log", LevelInfo, LevelInfo, true},
		{"Info handler with Debug log", LevelInfo, LevelDebug, false},
		{"Info handler with Error log", LevelInfo, LevelError, true},
		{"Error handler with Info log", LevelError, LevelInfo, false},
		{"Debug handler with Trace log", LevelDebug, LevelTrace, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewTextHandler(WithLevel(test.handlerLevel))
			enabled := handler.Enabled(context.Background(), test.logLevel)
			if enabled != test.expected {
				t.Errorf("Expected enabled=%v for handler level %v and log level %v", 
					test.expected, test.handlerLevel, test.logLevel)
			}
		})
	}
}

func TestHandlerWithAttrs(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(WithDestination(NewWriterDestination(buf)))

	attrs := []slog.Attr{
		slog.String("service", "api"),
		slog.Int("version", 1),
	}

	newHandler := handler.WithAttrs(attrs)
	record := NewRecordFromPool(LevelInfo, "Test message")

	err := newHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle with attrs failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"service":"api"`) {
		t.Errorf("Expected service attribute in output: %s", output)
	}
	if !strings.Contains(output, `"version":1`) {
		t.Errorf("Expected version attribute in output: %s", output)
	}
}

func TestHandlerWithGroup(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewJSONHandler(WithDestination(NewWriterDestination(buf)))

	// Add attributes via handler with group - this is how groups work
	attrs := []slog.Attr{slog.String("method", "GET")}
	groupHandler := handler.WithGroup("request").WithAttrs(attrs)
	record := NewRecordFromPool(LevelInfo, "Test message")

	err := groupHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle with group failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"request.method":"GET"`) {
		t.Errorf("Expected grouped attribute in output: %s", output)
	}
}

func TestMultiHandlerWithAttrs(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	handler1 := NewTextHandler(WithDestination(NewWriterDestination(buf1)))
	handler2 := NewJSONHandler(WithDestination(NewWriterDestination(buf2)))
	multiHandler := NewMultiHandler(handler1, handler2)

	attrs := []slog.Attr{slog.String("service", "test")}
	newMultiHandler := multiHandler.WithAttrs(attrs)

	record := NewRecordFromPool(LevelInfo, "Test message")
	err := newMultiHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("MultiHandler WithAttrs Handle failed: %v", err)
	}

	output1 := buf1.String()
	output2 := buf2.String()

	if !strings.Contains(output1, "service: test") {
		t.Errorf("Expected service attribute in text output: %s", output1)
	}
	if !strings.Contains(output2, `"service":"test"`) {
		t.Errorf("Expected service attribute in JSON output: %s", output2)
	}
}

func TestMultiHandlerWithGroup(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	handler1 := NewTextHandler(WithDestination(NewWriterDestination(buf1)))
	handler2 := NewJSONHandler(WithDestination(NewWriterDestination(buf2)))
	multiHandler := NewMultiHandler(handler1, handler2)

	// Add attributes via handler with group
	attrs := []slog.Attr{slog.String("key", "value")}
	groupHandler := multiHandler.WithGroup("test").WithAttrs(attrs)
	record := NewRecordFromPool(LevelInfo, "Test message")

	err := groupHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("MultiHandler WithGroup Handle failed: %v", err)
	}

	output1 := buf1.String()
	output2 := buf2.String()

	if !strings.Contains(output1, "test:") || !strings.Contains(output1, "key: value") {
		t.Errorf("Expected grouped attribute in text output: %s", output1)
	}
	if !strings.Contains(output2, `"test.key":"value"`) {
		t.Errorf("Expected grouped attribute in JSON output: %s", output2)
	}
}

func TestMultiHandlerEnabled(t *testing.T) {
	handler1 := NewTextHandler(WithLevel(LevelError))
	handler2 := NewJSONHandler(WithLevel(LevelInfo))
	multiHandler := NewMultiHandler(handler1, handler2)

	// Should be enabled if any handler is enabled
	if !multiHandler.Enabled(context.Background(), LevelInfo) {
		t.Error("MultiHandler should be enabled if any sub-handler is enabled")
	}

	// Should be disabled if no handlers are enabled
	if multiHandler.Enabled(context.Background(), LevelTrace) {
		t.Error("MultiHandler should be disabled if no sub-handlers are enabled")
	}
}

func TestHandlerWithDefaultOptions(t *testing.T) {
	handlers := []Handler{
		NewTextHandlerWithDefaults(),
		NewJSONHandlerWithDefaults(),
		NewXMLHandlerWithDefaults(),
		NewYAMLHandlerWithDefaults(),
		NewKeyValueHandlerWithDefaults(),
	}

	for i, handler := range handlers {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			if handler == nil {
				t.Fatal("Handler with defaults returned nil")
			}
			
			record := NewRecordFromPool(LevelInfo, "Test message")
			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("Handle with defaults failed: %v", err)
			}
		})
	}
}

func TestDeprecatedHandlerFunctions(t *testing.T) {
	dest := NewWriterDestination(os.Stdout)
	opts := &SawmillOptions{}

	// Test deprecated functions don't panic
	jsonHandler := NewJSONHandlerWithKey(dest, opts, "custom")
	if jsonHandler == nil {
		t.Error("NewJSONHandlerWithKey returned nil")
	}

	xmlHandler := NewXMLHandlerWithKey(dest, opts, "custom")
	if xmlHandler == nil {
		t.Error("NewXMLHandlerWithKey returned nil")
	}

	yamlHandler := NewYAMLHandlerWithKey(dest, opts, "custom")
	if yamlHandler == nil {
		t.Error("NewYAMLHandlerWithKey returned nil")
	}

	textHandler := NewTextHandlerWithColors(dest, opts, map[string]string{"key": "red"}, true)
	if textHandler == nil {
		t.Error("NewTextHandlerWithColors returned nil")
	}

	jsonHandlerColors := NewJSONHandlerWithColors(dest, opts, map[string]string{"key": "blue"}, true)
	if jsonHandlerColors == nil {
		t.Error("NewJSONHandlerWithColors returned nil")
	}
}

func TestBaseHandlerNeedsSource(t *testing.T) {
	tests := []struct {
		name     string
		handler  Handler
		expected bool
	}{
		{
			name:     "Text handler with source",
			handler:  NewTextHandler(WithSourceInfo(true)),
			expected: true,
		},
		{
			name:     "Text handler without source",
			handler:  NewTextHandler(WithSourceInfo(false)),
			expected: false,
		},
		{
			name:     "JSON handler with source",
			handler:  NewJSONHandler(WithSourceInfo(true)),
			expected: true,
		},
		{
			name:     "JSON handler without source",
			handler:  NewJSONHandler(WithSourceInfo(false)),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if baseHandler, ok := test.handler.(*TextHandler); ok {
				if got := baseHandler.BaseHandler.NeedsSource(); got != test.expected {
					t.Errorf("NeedsSource() = %v, want %v", got, test.expected)
				}
			} else if baseHandler, ok := test.handler.(*JSONHandler); ok {
				if got := baseHandler.BaseHandler.NeedsSource(); got != test.expected {
					t.Errorf("NeedsSource() = %v, want %v", got, test.expected)
				}
			}
		})
	}
}

func TestGetDestinationBuffer(t *testing.T) {
	// Test with nil destination
	buffer := getDestinationBuffer(nil)
	if buffer == nil {
		t.Error("getDestinationBuffer with nil should return default buffer")
	}

	// Test with writer destination
	buf := &bytes.Buffer{}
	writerDest := NewWriterDestination(buf)
	buffer = getDestinationBuffer(writerDest)
	if buffer == nil {
		t.Error("getDestinationBuffer with WriterDestination should return buffer")
	}

	// Test with file destination
	fileDest := NewFileDestination("test.log", 1024, 86400, false)
	buffer = getDestinationBuffer(fileDest)
	if buffer == nil {
		t.Error("getDestinationBuffer with FileDestination should return buffer")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"trace", LevelTrace},
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"panic", LevelPanic},
		{"mark", LevelMark},
		{"invalid", LevelInfo}, // default
		{"", LevelInfo},        // default
	}

	for _, test := range tests {
		result := parseLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLevel(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestDestinationMethods(t *testing.T) {
	// Test WriterDestination
	buf := &bytes.Buffer{}
	writerDest := NewWriterDestination(buf)
	
	n, err := writerDest.Write([]byte("test"))
	if err != nil {
		t.Errorf("WriterDestination.Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("WriterDestination.Write returned %d, want 4", n)
	}
	
	err = writerDest.Close()
	if err != nil {
		t.Errorf("WriterDestination.Close failed: %v", err)
	}

	// Test FileDestination methods
	fileDest := NewFileDestination("test.log", 1024, 86400, false)
	
	_, err = fileDest.Write([]byte("test"))
	if err == nil {
		t.Error("FileDestination.Write should return error - not implemented")
	}
	
	err = fileDest.Close()
	if err != nil {
		t.Errorf("FileDestination.Close failed: %v", err)
	}

	// Test NetworkDestination methods
	networkDest := &NetworkDestination{}
	
	_, err = networkDest.Write([]byte("test"))
	if err == nil {
		t.Error("NetworkDestination.Write should return error - not implemented")
	}
	
	err = networkDest.Close()
	if err != nil {
		t.Errorf("NetworkDestination.Close failed: %v", err)
	}
}

func TestTemporaryHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	originalHandler := NewTextHandler(WithDestination(NewWriterDestination(buf)))
	formatter := NewJSONFormatter()
	
	tempHandler := &temporaryHandler{
		originalHandler: originalHandler,
		formatter:       formatter,
	}

	record := NewRecordFromPool(LevelInfo, "Test message")
	err := tempHandler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("temporaryHandler.Handle failed: %v", err)
	}

	output := buf.String()
	// Should be JSON formatted even though original handler is text
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Errorf("Expected JSON output from temporary handler: %s", output)
	}

	// Test WithAttrs
	attrs := []slog.Attr{slog.String("key", "value")}
	newTempHandler := tempHandler.WithAttrs(attrs)
	if newTempHandler == nil {
		t.Error("temporaryHandler.WithAttrs returned nil")
	}

	// Test WithGroup
	groupHandler := tempHandler.WithGroup("test")
	if groupHandler == nil {
		t.Error("temporaryHandler.WithGroup returned nil")
	}

	// Test Enabled
	if !tempHandler.Enabled(context.Background(), LevelInfo) {
		t.Error("temporaryHandler should be enabled for info level")
	}
}