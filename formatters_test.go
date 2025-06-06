package sawmill

import (
	"strings"
	"testing"
	"time"
)

func TestJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("JSONFormatter.Format failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Errorf("Expected message in JSON output: %s", output)
	}
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("Expected level in JSON output: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("Expected attributes in JSON output: %s", output)
	}
}

func TestJSONFormatterPrettyPrint(t *testing.T) {
	formatter := NewJSONFormatter()
	formatter.PrettyPrint = true
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("JSONFormatter.Format with pretty print failed: %v", err)
	}

	output := string(data)
	// Pretty printed JSON should have indentation
	if !strings.Contains(output, "  ") {
		t.Errorf("Expected indented JSON output: %s", output)
	}
}

func TestJSONFormatterWithColors(t *testing.T) {
	colorMappings := map[string]string{
		"user": "\033[32m", // green
	}
	formatter := NewJSONFormatterWithColors(colorMappings)
	formatter.ColorOutput = true
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Attributes.SetFast("user", "john")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("JSONFormatter.Format with colors failed: %v", err)
	}

	output := string(data)
	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("Expected colored output: %s", output)
	}
}

func TestJSONFormatterWithCustomAttributesKey(t *testing.T) {
	formatter := NewJSONFormatterWithKey("custom_attrs")
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("JSONFormatter.Format with custom key failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, `"custom_attrs":`) {
		t.Errorf("Expected custom attributes key in output: %s", output)
	}
}

func TestJSONFormatterContentType(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter.ContentType() != "application/json" {
		t.Errorf("Expected content type application/json, got %s", formatter.ContentType())
	}
}

func TestXMLFormatter(t *testing.T) {
	formatter := NewXMLFormatter()
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("XMLFormatter.Format failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "<record>") {
		t.Errorf("Expected XML record tag: %s", output)
	}
	if !strings.Contains(output, "<message>Test message</message>") {
		t.Errorf("Expected message in XML output: %s", output)
	}
	if !strings.Contains(output, "<level>INFO</level>") {
		t.Errorf("Expected level in XML output: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected attributes in XML output: %s", output)
	}
}

func TestXMLFormatterContentType(t *testing.T) {
	formatter := NewXMLFormatter()
	if formatter.ContentType() != "application/xml" {
		t.Errorf("Expected content type application/xml, got %s", formatter.ContentType())
	}
}

func TestYAMLFormatter(t *testing.T) {
	formatter := NewYAMLFormatter()
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("YAMLFormatter.Format failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, `message: "Test message"`) {
		t.Errorf("Expected message in YAML output: %s", output)
	}
	if !strings.Contains(output, "level: INFO") {
		t.Errorf("Expected level in YAML output: %s", output)
	}
	if !strings.Contains(output, "key: value") {
		t.Errorf("Expected attributes in YAML output: %s", output)
	}
}

func TestYAMLFormatterContentType(t *testing.T) {
	formatter := NewYAMLFormatter()
	if formatter.ContentType() != "application/x-yaml" {
		t.Errorf("Expected content type application/x-yaml, got %s", formatter.ContentType())
	}
}

func TestTextFormatter(t *testing.T) {
	formatter := NewTextFormatter()
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("TextFormatter.Format failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected message in text output: %s", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected level in text output: %s", output)
	}
	if !strings.Contains(output, "key: value") {
		t.Errorf("Expected attributes in text output: %s", output)
	}
}

func TestTextFormatterMarkLevel(t *testing.T) {
	formatter := NewTextFormatter()
	
	record := NewRecordFromPool(LevelMark, "Test mark")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("TextFormatter.Format mark failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "MARK") || !strings.Contains(output, "=") {
		t.Errorf("Expected mark formatting in output: %s", output)
	}
}

func TestTextFormatterWithColors(t *testing.T) {
	colorMappings := map[string]string{
		"user": "\033[32m", // green
	}
	formatter := NewTextFormatterWithColors(colorMappings)
	formatter.ColorOutput = true
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Attributes.SetFast("user", "john")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("TextFormatter.Format with colors failed: %v", err)
	}

	output := string(data)
	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("Expected colored output: %s", output)
	}
}

func TestTextFormatterFlatAttributes(t *testing.T) {
	formatter := NewTextFormatter()
	formatter.AttributeFormat = "flat"
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Attributes.SetFast("key1", "value1")
	record.Attributes.SetFast("key2", "value2")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("TextFormatter.Format with flat attributes failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "key1=value1") || !strings.Contains(output, "key2=value2") {
		t.Errorf("Expected flat attributes in output: %s", output)
	}
}

func TestTextFormatterContentType(t *testing.T) {
	formatter := NewTextFormatter()
	if formatter.ContentType() != "text/plain" {
		t.Errorf("Expected content type text/plain, got %s", formatter.ContentType())
	}
}

func TestKeyValueFormatter(t *testing.T) {
	formatter := NewKeyValueFormatter()
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record.Attributes.SetFast("key", "value")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("KeyValueFormatter.Format failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "message=Test message") {
		t.Errorf("Expected message in key-value output: %s", output)
	}
	if !strings.Contains(output, "level=INFO") {
		t.Errorf("Expected level in key-value output: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected attributes in key-value output: %s", output)
	}
}

func TestKeyValueFormatterMarkLevel(t *testing.T) {
	formatter := NewKeyValueFormatter()
	
	record := NewRecordFromPool(LevelMark, "Test mark")
	record.Time = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("KeyValueFormatter.Format mark failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "MARK") || !strings.Contains(output, "=") {
		t.Errorf("Expected mark formatting in output: %s", output)
	}
}

func TestKeyValueFormatterWithColors(t *testing.T) {
	colorMappings := map[string]string{
		"user": "\033[32m", // green
	}
	formatter := NewKeyValueFormatterWithColors(colorMappings)
	formatter.ColorOutput = true
	
	record := NewRecordFromPool(LevelInfo, "Test message")
	record.Attributes.SetFast("user", "john")

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("KeyValueFormatter.Format with colors failed: %v", err)
	}

	output := string(data)
	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Errorf("Expected colored output: %s", output)
	}
}

func TestKeyValueFormatterStructExpansion(t *testing.T) {
	formatter := NewKeyValueFormatter()
	
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	record := NewRecordFromPool(LevelInfo, "User info")
	user := User{ID: 123, Name: "John"}
	record.Attributes.ExpandStruct("user", user)

	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("KeyValueFormatter.Format with struct failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "user.id=123") || !strings.Contains(output, "user.name=John") {
		t.Errorf("Expected expanded struct attributes: %s", output)
	}
}

func TestKeyValueFormatterContentType(t *testing.T) {
	formatter := NewKeyValueFormatter()
	if formatter.ContentType() != "text/plain" {
		t.Errorf("Expected content type text/plain, got %s", formatter.ContentType())
	}
}

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelTrace, "TRACE"},
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{LevelPanic, "PANIC"},
		{LevelMark, "MARK"},
		{Level(999), "UNKNOWN"},
	}

	for _, test := range tests {
		result := levelToString(test.level)
		if result != test.expected {
			t.Errorf("levelToString(%v) = %s, want %s", test.level, result, test.expected)
		}
	}
}

func TestJSONFormatterOptimizedFunctions(t *testing.T) {
	formatter := NewJSONFormatter()
	
	// Test writeJSONEscapedString
	record := NewRecordFromPool(LevelInfo, "Test with \"quotes\" and \n newlines")
	data, err := formatter.Format(record)
	if err != nil {
		t.Fatalf("Format with special characters failed: %v", err)
	}
	
	output := string(data)
	if !strings.Contains(output, `\"quotes\"`) {
		t.Errorf("Expected escaped quotes in output: %s", output)
	}
	if !strings.Contains(output, `\n`) {
		t.Errorf("Expected escaped newline in output: %s", output)
	}
}

func TestFormatterWithCustomKeys(t *testing.T) {
	// Test JSON formatter with custom key
	jsonFormatter := NewJSONFormatterWithKey("data")
	record := NewRecordFromPool(LevelInfo, "Test")
	record.Attributes.SetFast("key", "value")
	
	data, err := jsonFormatter.Format(record)
	if err != nil {
		t.Fatalf("JSON format with custom key failed: %v", err)
	}
	if !strings.Contains(string(data), `"data":`) {
		t.Errorf("Expected custom attributes key in JSON output: %s", string(data))
	}

	// Test XML formatter with custom key
	xmlFormatter := NewXMLFormatterWithKey("data")
	data, err = xmlFormatter.Format(record)
	if err != nil {
		t.Fatalf("XML format with custom key failed: %v", err)
	}
	
	// Test YAML formatter with custom key
	yamlFormatter := NewYAMLFormatterWithKey("data")
	data, err = yamlFormatter.Format(record)
	if err != nil {
		t.Fatalf("YAML format with custom key failed: %v", err)
	}
	if !strings.Contains(string(data), "data:") {
		t.Errorf("Expected custom attributes key in YAML output: %s", string(data))
	}
}

func TestFormatterIncludeSource(t *testing.T) {
	tests := []struct {
		name      string
		formatter Formatter
	}{
		{"JSON", NewJSONFormatter()},
		{"XML", NewXMLFormatter()},
		{"YAML", NewYAMLFormatter()},
		{"Text", NewTextFormatter()},
		{"KeyValue", NewKeyValueFormatter()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set include source to true
			switch f := test.formatter.(type) {
			case *JSONFormatter:
				f.IncludeSource = true
			case *XMLFormatter:
				f.IncludeSource = true
			case *YAMLFormatter:
				f.IncludeSource = true
			case *TextFormatter:
				f.IncludeSource = true
			case *KeyValueFormatter:
				f.IncludeSource = true
			}

			record := NewRecordFromPool(LevelInfo, "Test message")
			record.PC = 1 // Set a non-zero PC to trigger source capture
			
			data, err := test.formatter.Format(record)
			if err != nil {
				t.Fatalf("%s formatter with source failed: %v", test.name, err)
			}

			// Just verify it doesn't crash - source info may or may not be included
			// depending on whether we can resolve the PC
			if len(data) == 0 {
				t.Errorf("%s formatter produced empty output", test.name)
			}
		})
	}
}

func TestFormatterIncludeLevel(t *testing.T) {
	tests := []struct {
		name      string
		formatter Formatter
	}{
		{"JSON", NewJSONFormatter()},
		{"XML", NewXMLFormatter()},
		{"YAML", NewYAMLFormatter()},
		{"Text", NewTextFormatter()},
		{"KeyValue", NewKeyValueFormatter()},
	}

	for _, test := range tests {
		t.Run(test.name+" with level", func(t *testing.T) {
			// Ensure include level is true (default for most)
			switch f := test.formatter.(type) {
			case *JSONFormatter:
				f.IncludeLevel = true
			case *XMLFormatter:
				f.IncludeLevel = true
			case *YAMLFormatter:
				f.IncludeLevel = true
			case *TextFormatter:
				f.IncludeLevel = true
			case *KeyValueFormatter:
				f.IncludeLevel = true
			}

			record := NewRecordFromPool(LevelError, "Test error")
			data, err := test.formatter.Format(record)
			if err != nil {
				t.Fatalf("%s formatter with level failed: %v", test.name, err)
			}

			output := string(data)
			if !strings.Contains(output, "ERROR") {
				t.Errorf("%s formatter should include ERROR level: %s", test.name, output)
			}
		})

		t.Run(test.name+" without level", func(t *testing.T) {
			// Set include level to false
			switch f := test.formatter.(type) {
			case *JSONFormatter:
				f.IncludeLevel = false
			case *XMLFormatter:
				f.IncludeLevel = false
			case *YAMLFormatter:
				f.IncludeLevel = false
			case *TextFormatter:
				f.IncludeLevel = false
			case *KeyValueFormatter:
				f.IncludeLevel = false
			}

			record := NewRecordFromPool(LevelError, "Test error")
			data, err := test.formatter.Format(record)
			if err != nil {
				t.Fatalf("%s formatter without level failed: %v", test.name, err)
			}

			// Output should still be valid
			if len(data) == 0 {
				t.Errorf("%s formatter produced empty output", test.name)
			}
		})
	}
}