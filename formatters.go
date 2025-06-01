package sawmill

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// JSONFormatter implements Formatter for JSON output
type JSONFormatter struct {
	TimeFormat    string       // Time format for timestamps
	PrettyPrint   bool         // Whether to pretty-print JSON
	IncludeSource bool         // Whether to include source location
	IncludeLevel  bool         // Whether to include log level
	AttributesKey string       // Key name for attributes in JSON
	ColorOutput   bool         // Whether to apply color highlighting
	ColorScheme   *ColorScheme // Color scheme for syntax highlighting
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		TimeFormat:    time.RFC3339,
		PrettyPrint:   false,
		IncludeSource: false, // Default to false for better performance - can be enabled if needed
		IncludeLevel:  true,
		AttributesKey: "attributes",
		ColorOutput:   false,
		ColorScheme:   DefaultColorScheme(),
	}
}

// NewJSONFormatterWithColors creates a JSON formatter with custom color mappings
func NewJSONFormatterWithColors(colorMappings map[string]string) *JSONFormatter {
	formatter := NewJSONFormatter()
	formatter.ColorScheme = NewColorScheme(colorMappings)
	return formatter
}

func (f *JSONFormatter) Format(record *Record) ([]byte, error) {
	buf := GetBuffer()
	defer ReturnBuffer(buf)

	buf.WriteByte('{')
	first := true

	// Write timestamp
	buf.WriteString(`"timestamp":"`)
	buf.WriteString(record.Time.Format(f.TimeFormat))
	buf.WriteByte('"')
	first = false

	// Write message - use custom JSON escaping to avoid reflection
	if !first {
		buf.WriteByte(',')
	}
	buf.WriteString(`"message":"`)
	f.writeJSONEscapedString(buf, record.Message)
	buf.WriteByte('"')

	// Write level
	if f.IncludeLevel {
		buf.WriteByte(',')
		buf.WriteString(`"level":"`)
		buf.WriteString(f.levelString(record.Level))
		buf.WriteByte('"')
	}

	// Write source
	if f.IncludeSource && record.PC != 0 {
		if frame, ok := f.getFrame(record.PC); ok {
			buf.WriteByte(',')
			buf.WriteString(`"source":{"function":"`)
			buf.WriteString(frame.Function)
			buf.WriteString(`","file":"`)
			buf.WriteString(frame.File)
			buf.WriteString(`","line":`)
			f.writeJSONInt(buf, frame.Line)
			buf.WriteByte('}')
		}
	}

	// Write attributes using optimized MarshalJSON
	if !record.Attributes.IsEmpty() {
		attributesKey := f.AttributesKey
		if attributesKey == "" {
			attributesKey = "attributes"
		}
		
		buf.WriteByte(',')
		buf.WriteByte('"')
		buf.WriteString(attributesKey)
		buf.WriteString(`":`)
		
		attrBytes, err := record.Attributes.MarshalJSON()
		if err != nil {
			return nil, err
		}
		buf.Write(attrBytes)
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')

	// Handle pretty printing if needed
	var result []byte
	if f.PrettyPrint {
		// For pretty printing, fall back to standard JSON marshaling
		output := make(map[string]interface{})
		output["timestamp"] = record.Time.Format(f.TimeFormat)
		output["message"] = record.Message
		if f.IncludeLevel {
			output["level"] = f.levelString(record.Level)
		}
		if f.IncludeSource && record.PC != 0 {
			if frame, ok := f.getFrame(record.PC); ok {
				output["source"] = map[string]interface{}{
					"function": frame.Function,
					"file":     frame.File,
					"line":     frame.Line,
				}
			}
		}
		if !record.Attributes.IsEmpty() {
			attributesKey := f.AttributesKey
			if attributesKey == "" {
				attributesKey = "attributes"
			}
			output[attributesKey] = record.Attributes.ToNestedMap()
		}
		
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return nil, err
		}
		result = append(data, '\n')
	} else {
		// Copy optimized buffer contents
		result = make([]byte, buf.Len())
		copy(result, buf.Bytes())
	}

	if f.ColorOutput && f.ColorScheme != nil {
		f.ColorScheme.Enabled = true
		coloredJSON := f.ColorScheme.colorizeJSON(string(result))
		return []byte(coloredJSON), nil
	}

	return result, nil
}

func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

// XMLFormatter implements Formatter for XML output
type XMLFormatter struct {
	TimeFormat    string
	IncludeSource bool
	IncludeLevel  bool
	AttributesKey string
}

// XMLRecord represents the XML structure for log records
type XMLRecord struct {
	XMLName   xml.Name               `xml:"record"`
	Timestamp string                 `xml:"timestamp"`
	Level     string                 `xml:"level,omitempty"`
	Message   string                 `xml:"message"`
	Source    *XMLSource             `xml:"source,omitempty"`
	Data      map[string]interface{} `xml:",any"`
}

type XMLSource struct {
	Function string `xml:"function"`
	File     string `xml:"file"`
	Line     int    `xml:"line"`
}

// NewXMLFormatter creates a new XML formatter
func NewXMLFormatter() *XMLFormatter {
	return &XMLFormatter{
		TimeFormat:    time.RFC3339,
		IncludeSource: true,
		IncludeLevel:  true,
		AttributesKey: "attributes",
	}
}

func (f *XMLFormatter) Format(record *Record) ([]byte, error) {
	xmlRecord := XMLRecord{
		Timestamp: record.Time.Format(f.TimeFormat),
		Message:   record.Message,
		Data:      make(map[string]interface{}),
	}

	if f.IncludeLevel {
		xmlRecord.Level = f.levelString(record.Level)
	}

	if f.IncludeSource && record.PC != 0 {
		if frame, ok := f.getFrame(record.PC); ok {
			xmlRecord.Source = &XMLSource{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			}
		}
	}

	if !record.Attributes.IsEmpty() {
		attributesKey := f.AttributesKey
		if attributesKey == "" {
			attributesKey = "attributes"
		}
		xmlRecord.Data[attributesKey] = record.Attributes.ToNestedMap()
	}

	data, err := xml.MarshalIndent(xmlRecord, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add newline to separate XML records
	data = append(data, '\n')
	return data, nil
}

func (f *XMLFormatter) ContentType() string {
	return "application/xml"
}

// YAMLFormatter implements Formatter for YAML output
type YAMLFormatter struct {
	TimeFormat    string
	IncludeSource bool
	IncludeLevel  bool
	AttributesKey string
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{
		TimeFormat:    time.RFC3339,
		IncludeSource: true,
		IncludeLevel:  true,
		AttributesKey: "attributes",
	}
}

func (f *YAMLFormatter) Format(record *Record) ([]byte, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("timestamp: %s\n", record.Time.Format(f.TimeFormat)))

	if f.IncludeLevel {
		output.WriteString(fmt.Sprintf("level: %s\n", f.levelString(record.Level)))
	}

	output.WriteString(fmt.Sprintf("message: %q\n", record.Message))

	if f.IncludeSource && record.PC != 0 {
		if frame, ok := f.getFrame(record.PC); ok {
			output.WriteString("source:\n")
			output.WriteString(fmt.Sprintf("  function: %s\n", frame.Function))
			output.WriteString(fmt.Sprintf("  file: %s\n", frame.File))
			output.WriteString(fmt.Sprintf("  line: %d\n", frame.Line))
		}
	}

	if !record.Attributes.IsEmpty() {
		attributesKey := f.AttributesKey
		if attributesKey == "" {
			attributesKey = "attributes"
		}
		output.WriteString(fmt.Sprintf("%s:\n", attributesKey))
		f.writeYAMLAttributes(&output, record.Attributes, 1)
	}

	return []byte(output.String()), nil
}

func (f *YAMLFormatter) writeYAMLAttributes(output *strings.Builder, attrs *FlatAttributes, indent int) {
	indentStr := strings.Repeat("  ", indent)
	
	// For flat attributes, we can either output them flat or convert to nested
	// Let's use the flat approach for simplicity and performance
	attrs.Walk(func(path []string, value interface{}) {
		key := strings.Join(path, ".")
		output.WriteString(fmt.Sprintf("%s%s: %v\n", indentStr, key, value))
	})
}

func (f *YAMLFormatter) ContentType() string {
	return "application/x-yaml"
}

// TextFormatter implements Formatter for human-readable text output
type TextFormatter struct {
	TimeFormat      string
	IncludeSource   bool
	IncludeLevel    bool
	AttributeFormat string // "flat" or "nested"
	ColorOutput     bool
	AttributesKey   string       // Key name for attributes (unused in text format)
	ColorScheme     *ColorScheme // Color scheme for syntax highlighting
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		TimeFormat:      "2006-01-02 15:04:05",
		IncludeSource:   false, // Default to false for better performance
		IncludeLevel:    true,
		AttributeFormat: "nested",
		ColorOutput:     false,
		AttributesKey:   "attributes",
		ColorScheme:     DefaultColorScheme(),
	}
}

// NewTextFormatterWithColors creates a text formatter with custom color mappings
func NewTextFormatterWithColors(colorMappings map[string]string) *TextFormatter {
	formatter := NewTextFormatter()
	formatter.ColorScheme = NewColorScheme(colorMappings)
	return formatter
}

func (f *TextFormatter) Format(record *Record) ([]byte, error) {
	var output strings.Builder

	if record.Level == LevelMark {
		return f.formatMark(record)
	}

	output.WriteString(record.Time.Format(f.TimeFormat))

	if f.IncludeLevel {
		level := f.levelString(record.Level)
		if f.ColorOutput {
			level = f.colorizeLevel(level, record.Level)
		}
		output.WriteString(fmt.Sprintf(" [%s]", level))
	}

	if f.IncludeSource && record.PC != 0 {
		if frame, ok := f.getFrame(record.PC); ok {
			output.WriteString(fmt.Sprintf(" %s:%d", frame.File, frame.Line))
		}
	}

	output.WriteString(fmt.Sprintf(" %s", record.Message))
	if !record.Attributes.IsEmpty() {
		if f.ColorOutput && f.ColorScheme != nil {
			f.ColorScheme.Enabled = true
			coloredAttrs := f.ColorScheme.ColorizeAttributes(record.Attributes, f.AttributeFormat)
			output.WriteString(coloredAttrs)
		} else {
			if f.AttributeFormat == "flat" {
				f.writeTextAttributesFlat(&output, record.Attributes)
			} else {
				f.writeTextAttributesNested(&output, record.Attributes, 0)
			}
		}
	}

	output.WriteString("\n")
	return []byte(output.String()), nil
}

func (f *TextFormatter) formatMark(record *Record) ([]byte, error) {
	var output strings.Builder

	separator := strings.Repeat("=", 80)

	if f.ColorOutput {
		output.WriteString(fmt.Sprintf("\033[44m%s\033[0m\n", separator))
		output.WriteString(fmt.Sprintf("\033[1;44m MARK: %s \033[0m\n", record.Message))

		// Apply color to timestamp label and value
		if f.ColorScheme != nil {
			f.ColorScheme.Enabled = true
			timestampLabel := f.ColorScheme.colorizeKey("timestamp")
			timestampValue := f.ColorScheme.colorizeValue(record.Time.Format(f.TimeFormat))
			output.WriteString(fmt.Sprintf("%s: %s", timestampLabel, timestampValue))
		} else {
			output.WriteString(fmt.Sprintf("timestamp: %s", record.Time.Format(f.TimeFormat)))
		}
	} else {
		output.WriteString(fmt.Sprintf("%s\n", separator))
		output.WriteString(fmt.Sprintf(" MARKED @ %s ", record.Time.Format(f.TimeFormat)))
	}

	if !record.Attributes.IsEmpty() {
		if f.ColorOutput && f.ColorScheme != nil {
			f.ColorScheme.Enabled = true
			coloredAttrs := f.ColorScheme.ColorizeAttributes(record.Attributes, f.AttributeFormat)
			output.WriteString(coloredAttrs)
		} else {
			if f.AttributeFormat == "flat" {
				f.writeTextAttributesFlat(&output, record.Attributes)
			} else {
				f.writeTextAttributesNested(&output, record.Attributes, 0)
			}
		}
	}

	output.WriteString("\n")

	if f.ColorOutput {
		output.WriteString(fmt.Sprintf("\033[44m%s\033[0m\n", separator))
	} else {
		output.WriteString(fmt.Sprintf("%s\n", separator))
	}

	return []byte(output.String()), nil
}

func (f *TextFormatter) writeTextAttributesFlat(output *strings.Builder, attrs *FlatAttributes) {
	attrs.Walk(func(path []string, value interface{}) {
		key := strings.Join(path, ".")
		output.WriteString(fmt.Sprintf(" %s=%v", key, value))
	})
}

func (f *TextFormatter) writeTextAttributesNested(output *strings.Builder, attrs *FlatAttributes, indent int) {
	// For FlatAttributes, we can convert to nested structure and format
	nested := attrs.ToNestedMap()
	f.writeNestedMap(output, nested, indent)
}

func (f *TextFormatter) writeNestedMap(output *strings.Builder, data map[string]interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)
	
	for key, value := range data {
		output.WriteString(fmt.Sprintf("\n%s%s:", indentStr, key))
		if nestedMap, ok := value.(map[string]interface{}); ok {
			f.writeNestedMap(output, nestedMap, indent+1)
		} else {
			output.WriteString(fmt.Sprintf(" %v", value))
		}
	}
}

func (f *TextFormatter) colorizeLevel(level string, l Level) string {
	switch l {
	case LevelTrace:
		return fmt.Sprintf("\033[37m%s\033[0m", level) // White
	case LevelDebug:
		return fmt.Sprintf("\033[36m%s\033[0m", level) // Cyan
	case LevelInfo:
		return fmt.Sprintf("\033[32m%s\033[0m", level) // Green
	case LevelWarn:
		return fmt.Sprintf("\033[33m%s\033[0m", level) // Yellow
	case LevelError:
		return fmt.Sprintf("\033[31m%s\033[0m", level) // Red
	case LevelFatal:
		return fmt.Sprintf("\033[35m%s\033[0m", level) // Magenta
	case LevelPanic:
		return fmt.Sprintf("\033[41m%s\033[0m", level) // Red background
	case LevelMark:
		return fmt.Sprintf("\033[1;44m%s\033[0m", level) // Bold white on blue background
	default:
		return level
	}
}

func (f *TextFormatter) ContentType() string {
	return "text/plain"
}

// Common helper methods
func (f *JSONFormatter) levelString(level Level) string {
	return levelToString(level)
}

func (f *XMLFormatter) levelString(level Level) string {
	return levelToString(level)
}

func (f *YAMLFormatter) levelString(level Level) string {
	return levelToString(level)
}

func (f *TextFormatter) levelString(level Level) string {
	return levelToString(level)
}

func levelToString(level Level) string {
	switch level {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	case LevelPanic:
		return "PANIC"
	case LevelMark:
		return "MARK"
	default:
		return "UNKNOWN"
	}
}

func (f *JSONFormatter) getFrame(pc uintptr) (runtime.Frame, bool) {
	return getFrame(pc)
}

func (f *XMLFormatter) getFrame(pc uintptr) (runtime.Frame, bool) {
	return getFrame(pc)
}

func (f *YAMLFormatter) getFrame(pc uintptr) (runtime.Frame, bool) {
	return getFrame(pc)
}

func (f *TextFormatter) getFrame(pc uintptr) (runtime.Frame, bool) {
	return getFrame(pc)
}

func getFrame(pc uintptr) (runtime.Frame, bool) {
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, ok := frames.Next()
	return frame, ok
}

// NewJSONFormatterWithKey creates a JSON formatter with a custom attributes key
func NewJSONFormatterWithKey(attributesKey string) *JSONFormatter {
	formatter := NewJSONFormatter()
	if attributesKey != "" {
		formatter.AttributesKey = attributesKey
	}
	return formatter
}

// NewXMLFormatterWithKey creates an XML formatter with a custom attributes key
func NewXMLFormatterWithKey(attributesKey string) *XMLFormatter {
	formatter := NewXMLFormatter()
	if attributesKey != "" {
		formatter.AttributesKey = attributesKey
	}
	return formatter
}

// NewYAMLFormatterWithKey creates a YAML formatter with a custom attributes key
func NewYAMLFormatterWithKey(attributesKey string) *YAMLFormatter {
	formatter := NewYAMLFormatter()
	if attributesKey != "" {
		formatter.AttributesKey = attributesKey
	}
	return formatter
}

// KeyValueFormatter implements Formatter for key=value output
type KeyValueFormatter struct {
	TimeFormat    string
	IncludeSource bool
	IncludeLevel  bool
	ColorOutput   bool
	ColorScheme   *ColorScheme
}

// NewKeyValueFormatter creates a new key-value formatter
func NewKeyValueFormatter() *KeyValueFormatter {
	return &KeyValueFormatter{
		TimeFormat:    "2006-01-02 15:04:05",
		IncludeSource: true,
		IncludeLevel:  true,
		ColorOutput:   false,
		ColorScheme:   DefaultColorScheme(),
	}
}

// NewKeyValueFormatterWithColors creates a key-value formatter with custom color mappings
func NewKeyValueFormatterWithColors(colorMappings map[string]string) *KeyValueFormatter {
	formatter := NewKeyValueFormatter()
	formatter.ColorScheme = NewColorScheme(colorMappings)
	return formatter
}

func (f *KeyValueFormatter) Format(record *Record) ([]byte, error) {
	var output strings.Builder

	if record.Level == LevelMark {
		return f.formatMark(record)
	}

	// Start with timestamp
	if f.ColorOutput && f.ColorScheme != nil {
		f.ColorScheme.Enabled = true
		output.WriteString(f.formatKeyValue("timestamp", record.Time.Format(f.TimeFormat), false))
	} else {
		output.WriteString(fmt.Sprintf("timestamp=%s", record.Time.Format(f.TimeFormat)))
	}

	// Add level
	if f.IncludeLevel {
		level := f.levelString(record.Level)
		if f.ColorOutput && f.ColorScheme != nil {
			output.WriteString(" ")
			output.WriteString(f.formatKeyValue("level", level, false))
		} else {
			output.WriteString(fmt.Sprintf(" level=%s", level))
		}
	}

	// Add source
	if f.IncludeSource && record.PC != 0 {
		if frame, ok := f.getFrame(record.PC); ok {
			sourceValue := fmt.Sprintf("%s:%d", frame.File, frame.Line)
			if f.ColorOutput && f.ColorScheme != nil {
				output.WriteString(" ")
				output.WriteString(f.formatKeyValue("source", sourceValue, false))
			} else {
				output.WriteString(fmt.Sprintf(" source=%s", sourceValue))
			}
		}
	}

	// Add message
	if f.ColorOutput && f.ColorScheme != nil {
		output.WriteString(" ")
		output.WriteString(f.formatKeyValue("message", record.Message, false))
	} else {
		output.WriteString(fmt.Sprintf(" message=%s", record.Message))
	}

	// Add attributes in flat key=value format
	if !record.Attributes.IsEmpty() {
		f.writeKeyValueAttributes(&output, record.Attributes)
	}

	output.WriteString("\n")
	return []byte(output.String()), nil
}

func (f *KeyValueFormatter) formatMark(record *Record) ([]byte, error) {
	var output strings.Builder

	separator := strings.Repeat("=", 80)

	if f.ColorOutput {
		output.WriteString(fmt.Sprintf("\033[44m%s\033[0m\n", separator))
		output.WriteString(fmt.Sprintf("\033[1;44m MARK: %s \033[0m\n", record.Message))

		// Apply color to timestamp label and value
		if f.ColorScheme != nil {
			f.ColorScheme.Enabled = true
			output.WriteString(f.formatKeyValue("timestamp", record.Time.Format(f.TimeFormat), true))
		} else {
			output.WriteString(fmt.Sprintf("timestamp=%s", record.Time.Format(f.TimeFormat)))
		}
	} else {
		output.WriteString(fmt.Sprintf("%s\n", separator))
		output.WriteString(fmt.Sprintf(" MARKED @ %s ", record.Time.Format(f.TimeFormat)))
	}

	if !record.Attributes.IsEmpty() {
		if f.ColorOutput && f.ColorScheme != nil {
			f.ColorScheme.Enabled = true
			f.writeKeyValueAttributes(&output, record.Attributes)
		} else {
			f.writeKeyValueAttributes(&output, record.Attributes)
		}
	}

	output.WriteString("\n")

	if f.ColorOutput {
		output.WriteString(fmt.Sprintf("\033[44m%s\033[0m\n", separator))
	} else {
		output.WriteString(fmt.Sprintf("%s\n", separator))
	}

	return []byte(output.String()), nil
}

func (f *KeyValueFormatter) writeKeyValueAttributes(output *strings.Builder, attrs *FlatAttributes) {
	attrs.Walk(func(path []string, value interface{}) {
		f.writeExpandedValue(output, path, value)
	})
}

func (f *KeyValueFormatter) writeExpandedValue(output *strings.Builder, path []string, value interface{}) {
	// Use reflection to check if this is a struct and expand it
	if f.shouldExpandStruct(value) {
		// For FlatAttributes, we can use the built-in struct expansion
		prefix := strings.Join(path, ".")
		attrs := NewFlatAttributes()
		attrs.ExpandStruct(prefix, value)
		
		// Walk the expanded attributes
		attrs.Walk(func(expandedPath []string, expandedValue interface{}) {
			key := strings.Join(expandedPath, ".")
			if f.ColorOutput && f.ColorScheme != nil {
				output.WriteString(" ")
				output.WriteString(f.formatKeyValue(key, fmt.Sprintf("%+v", expandedValue), false))
			} else {
				output.WriteString(fmt.Sprintf(" %s=%+v", key, expandedValue))
			}
		})
	} else {
		key := strings.Join(path, ".")
		if f.ColorOutput && f.ColorScheme != nil {
			output.WriteString(" ")
			output.WriteString(f.formatKeyValue(key, fmt.Sprintf("%+v", value), false))
		} else {
			output.WriteString(fmt.Sprintf(" %s=%+v", key, value))
		}
	}
}

func (f *KeyValueFormatter) shouldExpandStruct(value interface{}) bool {
	if value == nil {
		return false
	}
	
	val := reflect.ValueOf(value)
	// Handle pointers to structs
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	
	return val.Kind() == reflect.Struct
}


func (f *KeyValueFormatter) formatKeyValue(key string, value string, newlinePrefix bool) string {
	if !f.ColorOutput || f.ColorScheme == nil {
		if newlinePrefix {
			return fmt.Sprintf("\n%s=%s", key, value)
		}
		return fmt.Sprintf("%s=%s", key, value)
	}

	// Get the appropriate color for this key
	color := f.getKeyColor(key)

	// Create dimmed version of the key
	dimmedKey := f.dimColor(key, color)

	// Apply full color to the value
	coloredValue := f.applyColor(value, color)

	if newlinePrefix {
		return fmt.Sprintf("\n%s=%s", dimmedKey, coloredValue)
	}
	return fmt.Sprintf("%s=%s", dimmedKey, coloredValue)
}

func (f *KeyValueFormatter) dimColor(text string, color string) string {
	if color == "" {
		return fmt.Sprintf("\033[2m%s\033[0m", text) // Default dim
	}
	// Extract color code and add dim attribute (2)
	if strings.HasPrefix(color, "\033[") && strings.HasSuffix(color, "m") {
		colorCode := color[2 : len(color)-1] // Remove \033[ and m
		return fmt.Sprintf("\033[2;%sm%s\033[0m", colorCode, text)
	}
	return fmt.Sprintf("\033[2m%s\033[0m", text)
}

func (f *KeyValueFormatter) getKeyColor(keyPath string) string {
	if f.ColorScheme == nil {
		return ""
	}

	// Check for exact match in custom mappings
	if color, exists := f.ColorScheme.KeyMappings[keyPath]; exists {
		return color
	}

	// Check for partial matches (e.g., "user" matches "user.profile.name")
	for mappedKey, color := range f.ColorScheme.KeyMappings {
		if strings.HasPrefix(keyPath, mappedKey+".") || strings.HasSuffix(keyPath, "."+mappedKey) {
			return color
		}
	}

	// Use default key color
	return f.ColorScheme.Keys
}

func (f *KeyValueFormatter) applyColor(text string, color string) string {
	if color == "" {
		return text
	}
	return fmt.Sprintf("%s%s\033[0m", color, text)
}

func (f *KeyValueFormatter) ContentType() string {
	return "text/plain"
}

func (f *KeyValueFormatter) levelString(level Level) string {
	return levelToString(level)
}

func (f *KeyValueFormatter) getFrame(pc uintptr) (runtime.Frame, bool) {
	return getFrame(pc)
}

// Optimized JSON helper functions to avoid reflection and allocations

// writeJSONEscapedString writes a JSON-escaped string directly to the buffer
func (f *JSONFormatter) writeJSONEscapedString(buf *bytes.Buffer, s string) {
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r < 32 {
				// Control characters - use Unicode escape
				buf.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				buf.WriteRune(r)
			}
		}
	}
}

// writeJSONInt writes an integer directly without string conversion overhead
func (f *JSONFormatter) writeJSONInt(buf *bytes.Buffer, i int) {
	if i == 0 {
		buf.WriteByte('0')
		return
	}
	
	if i < 0 {
		buf.WriteByte('-')
		i = -i
	}
	
	// Calculate number of digits to avoid allocations
	var digits [20]byte // enough for 64-bit int
	pos := 19
	
	for i > 0 {
		digits[pos] = byte('0' + i%10)
		i /= 10
		pos--
	}
	
	buf.Write(digits[pos+1:])
}
