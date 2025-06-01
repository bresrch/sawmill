package sawmill

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	formatter Formatter
	buffer    Buffer
	level     Level
	attrs     *FlatAttributes
	groups    []string
	mu        sync.RWMutex
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(formatter Formatter, buffer Buffer, level Level) *BaseHandler {
	return &BaseHandler{
		formatter: formatter,
		buffer:    buffer,
		level:     level,
		attrs:     NewFlatAttributes(),
		groups:    make([]string, 0),
	}
}

func (h *BaseHandler) Handle(ctx context.Context, record *Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	h.mu.RLock()

	// Fast path: if no handler attributes, format directly without cloning
	if h.attrs.IsEmpty() {
		h.mu.RUnlock()
		data, err := h.formatter.Format(record)
		if err != nil {
			return err
		}
		_, err = h.buffer.Write(data)
		return err
	}

	// Slow path: clone and merge when handler has attributes
	recordCopy := &Record{
		Time:       record.Time,
		Level:      record.Level,
		Message:    record.Message,
		Attributes: record.Attributes.Clone(),
		Context:    record.Context,
		PC:         record.PC,
	}

	// Add handler attributes
	recordCopy.Attributes.Merge(h.attrs)

	h.mu.RUnlock()

	// Format the record
	data, err := h.formatter.Format(recordCopy)
	if err != nil {
		return err
	}

	// Write to buffer
	_, err = h.buffer.Write(data)
	return err
}

func (h *BaseHandler) WithAttrs(attrs []slog.Attr) Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newHandler := &BaseHandler{
		formatter: h.formatter,
		buffer:    h.buffer,
		level:     h.level,
		attrs:     h.attrs.Clone(),
		groups:    make([]string, len(h.groups)),
	}
	copy(newHandler.groups, h.groups)

	for _, attr := range attrs {
		keyPath := make([]string, len(h.groups))
		copy(keyPath, h.groups)
		keyPath = append(keyPath, attr.Key)
		newHandler.attrs.Set(keyPath, attr.Value.Any())
	}

	return newHandler
}

func (h *BaseHandler) WithGroup(name string) Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &BaseHandler{
		formatter: h.formatter,
		buffer:    h.buffer,
		level:     h.level,
		attrs:     h.attrs.Clone(),
		groups:    newGroups,
	}
}

func (h *BaseHandler) Enabled(ctx context.Context, level Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return level >= h.level
}

// NeedsSource indicates if this handler needs source information
func (h *BaseHandler) NeedsSource() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Check if formatter is configured to include source
	switch f := h.formatter.(type) {
	case *JSONFormatter:
		return f.IncludeSource
	case *TextFormatter:
		return f.IncludeSource
	case *XMLFormatter:
		return f.IncludeSource
	case *YAMLFormatter:
		return f.IncludeSource
	case *KeyValueFormatter:
		return f.IncludeSource
	default:
		return true // Safe default
	}
}

// TextHandler implements Handler for text output
type TextHandler struct {
	*BaseHandler
}

// NewTextHandler creates a new text handler with the given options
func NewTextHandler(options ...HandlerOption) *TextHandler {
	opts := NewHandlerOptions(options...)

	buffer := createBuffer(opts)
	level := determineLevel(opts)
	formatter := createTextFormatter(opts)

	return &TextHandler{
		BaseHandler: NewBaseHandler(formatter, buffer, level),
	}
}

// NewTextHandlerWithDefaults creates a text handler with default options
func NewTextHandlerWithDefaults() *TextHandler {
	return NewTextHandler()
}

// JSONHandler implements Handler for JSON output
type JSONHandler struct {
	*BaseHandler
}

// NewJSONHandler creates a new JSON handler with the given options
func NewJSONHandler(options ...HandlerOption) *JSONHandler {
	opts := NewHandlerOptions(options...)

	buffer := createBuffer(opts)
	level := determineLevel(opts)
	formatter := createJSONFormatter(opts)

	return &JSONHandler{
		BaseHandler: NewBaseHandler(formatter, buffer, level),
	}
}

// NewJSONHandlerWithDefaults creates a JSON handler with default options
func NewJSONHandlerWithDefaults() *JSONHandler {
	return NewJSONHandler()
}

// XMLHandler implements Handler for XML output
type XMLHandler struct {
	*BaseHandler
}

// NewXMLHandler creates a new XML handler with the given options
func NewXMLHandler(options ...HandlerOption) *XMLHandler {
	opts := NewHandlerOptions(options...)

	buffer := createBuffer(opts)
	level := determineLevel(opts)
	formatter := createXMLFormatter(opts)

	return &XMLHandler{
		BaseHandler: NewBaseHandler(formatter, buffer, level),
	}
}

// NewXMLHandlerWithDefaults creates an XML handler with default options
func NewXMLHandlerWithDefaults() *XMLHandler {
	return NewXMLHandler()
}

// YAMLHandler implements Handler for YAML output
type YAMLHandler struct {
	*BaseHandler
}

// NewYAMLHandler creates a new YAML handler with the given options
func NewYAMLHandler(options ...HandlerOption) *YAMLHandler {
	opts := NewHandlerOptions(options...)

	buffer := createBuffer(opts)
	level := determineLevel(opts)
	formatter := createYAMLFormatter(opts)

	return &YAMLHandler{
		BaseHandler: NewBaseHandler(formatter, buffer, level),
	}
}

// NewYAMLHandlerWithDefaults creates a YAML handler with default options
func NewYAMLHandlerWithDefaults() *YAMLHandler {
	return NewYAMLHandler()
}

// KeyValueHandler implements Handler for key=value output
type KeyValueHandler struct {
	*BaseHandler
}

// NewKeyValueHandler creates a new key-value handler with the given options
func NewKeyValueHandler(options ...HandlerOption) *KeyValueHandler {
	opts := NewHandlerOptions(options...)

	buffer := createBuffer(opts)
	level := determineLevel(opts)
	formatter := createKeyValueFormatter(opts)

	return &KeyValueHandler{
		BaseHandler: NewBaseHandler(formatter, buffer, level),
	}
}

// NewKeyValueHandlerWithDefaults creates a key-value handler with default options
func NewKeyValueHandlerWithDefaults() *KeyValueHandler {
	return NewKeyValueHandler()
}

// MultiHandler allows writing to multiple handlers simultaneously
type MultiHandler struct {
	handlers []Handler
	mu       sync.RWMutex
}

// NewMultiHandler creates a new multi-handler
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

func (h *MultiHandler) Handle(ctx context.Context, record *Record) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lastErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) Handler {
	h.mu.RLock()
	newHandlers := make([]Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	h.mu.RUnlock()

	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) Handler {
	h.mu.RLock()
	newHandlers := make([]Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	h.mu.RUnlock()

	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Helper functions

func getDestinationBuffer(dest Destination) Buffer {
	if dest == nil {
		return NewWriterBuffer(os.Stdout)
	}

	switch d := dest.(type) {
	case *FileDestination:
		buffer, err := NewFileBuffer(d.Path, 4096, d.MaxSize, true)
		if err != nil {
			return NewWriterBuffer(os.Stdout)
		}
		return buffer
	case *WriterDestination:
		return NewWriterBuffer(d.Writer)
	case *NetworkDestination:
		// Network destinations would require additional implementation
		return NewWriterBuffer(os.Stdout)
	default:
		return NewWriterBuffer(os.Stdout)
	}
}

func parseLevel(levelStr string) Level {
	switch levelStr {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "panic":
		return LevelPanic
	case "mark":
		return LevelMark
	default:
		return LevelInfo
	}
}

// NewWriterDestination creates a new writer destination
func NewWriterDestination(writer io.Writer) *WriterDestination {
	if writer == nil {
		writer = os.Stdout
	}
	return &WriterDestination{Writer: writer}
}

// NewFileDestination creates a new file destination
func NewFileDestination(path string, maxSize int64, maxAge int64, compress bool) *FileDestination {
	return &FileDestination{
		Path:     path,
		MaxSize:  maxSize,
		Compress: compress,
	}
}

func (d *WriterDestination) Write(data []byte) (int, error) {
	return d.Writer.Write(data)
}

func (d *WriterDestination) Close() error {
	if closer, ok := d.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (d *FileDestination) Write(data []byte) (int, error) {
	return 0, fmt.Errorf("FileDestination.Write not implemented - use with handler")
}

func (d *FileDestination) Close() error {
	return nil
}

func (d *NetworkDestination) Write(data []byte) (int, error) {
	return 0, fmt.Errorf("NetworkDestination.Write not implemented - use with handler")
}

func (d *NetworkDestination) Close() error {
	return nil
}

// NewJSONHandlerWithKey creates a JSON handler with a custom attributes key
// Deprecated: Use NewJSONHandler with functional options instead
func NewJSONHandlerWithKey(dest Destination, opts *SawmillOptions, attributesKey string) *JSONHandler {
	var options []HandlerOption
	options = append(options, WithDestination(dest))
	options = append(options, WithAttributesKey(attributesKey))

	if opts != nil {
		options = append(options, WithSawmillOptions(opts))
	}

	return NewJSONHandler(options...)
}

// NewXMLHandlerWithKey creates an XML handler with a custom attributes key
// Deprecated: Use NewXMLHandler with functional options instead
func NewXMLHandlerWithKey(dest Destination, opts *SawmillOptions, attributesKey string) *XMLHandler {
	var options []HandlerOption
	options = append(options, WithDestination(dest))
	options = append(options, WithAttributesKey(attributesKey))

	if opts != nil {
		options = append(options, WithSawmillOptions(opts))
	}

	return NewXMLHandler(options...)
}

// NewYAMLHandlerWithKey creates a YAML handler with a custom attributes key
// Deprecated: Use NewYAMLHandler with functional options instead
func NewYAMLHandlerWithKey(dest Destination, opts *SawmillOptions, attributesKey string) *YAMLHandler {
	var options []HandlerOption
	options = append(options, WithDestination(dest))
	options = append(options, WithAttributesKey(attributesKey))

	if opts != nil {
		options = append(options, WithSawmillOptions(opts))
	}

	return NewYAMLHandler(options...)
}

// NewTextHandlerWithColors creates a text handler with custom color mappings
// Deprecated: Use NewTextHandler with functional options instead
func NewTextHandlerWithColors(dest Destination, opts *SawmillOptions, colorMappings map[string]string, enableColors bool) *TextHandler {
	var options []HandlerOption
	options = append(options, WithDestination(dest))
	options = append(options, WithColorMappings(colorMappings))
	options = append(options, WithColorsEnabled(enableColors))

	if opts != nil {
		options = append(options, WithSawmillOptions(opts))
	}

	return NewTextHandler(options...)
}

// NewJSONHandlerWithColors creates a JSON handler with custom color mappings
// Deprecated: Use NewJSONHandler with functional options instead
func NewJSONHandlerWithColors(dest Destination, opts *SawmillOptions, colorMappings map[string]string, enableColors bool) *JSONHandler {
	var options []HandlerOption
	options = append(options, WithDestination(dest))
	options = append(options, WithColorMappings(colorMappings))
	options = append(options, WithColorsEnabled(enableColors))

	if opts != nil {
		options = append(options, WithSawmillOptions(opts))
	}

	return NewJSONHandler(options...)
}

// Helper functions for the options pattern

func createBuffer(options *HandlerOptions) Buffer {
	if options.sawmillOpts != nil && options.sawmillOpts.LogFile != "" {
		fileBuffer, err := NewFileBuffer(
			options.sawmillOpts.LogFile,
			4096,
			int64(options.sawmillOpts.MaxSize)*1024*1024,
			true,
		)
		if err != nil {
			return NewWriterBuffer(os.Stdout)
		}
		return fileBuffer
	}
	return getDestinationBuffer(options.destination)
}

func determineLevel(options *HandlerOptions) Level {
	if options.sawmillOpts != nil {
		return parseLevel(options.sawmillOpts.LogLevel)
	}
	return options.level
}

func createTextFormatter(options *HandlerOptions) *TextFormatter {
	formatter := NewTextFormatter()
	formatter.TimeFormat = options.timeFormat
	formatter.IncludeSource = options.includeSource
	formatter.IncludeLevel = options.includeLevel
	formatter.AttributeFormat = options.attrFormat
	formatter.ColorOutput = options.colorOutput
	formatter.AttributesKey = options.attributesKey

	if options.enableColors {
		formatter.ColorScheme = NewColorScheme(options.colorMappings)
		formatter.ColorOutput = true
	}

	return formatter
}

func createJSONFormatter(options *HandlerOptions) *JSONFormatter {
	formatter := NewJSONFormatter()
	formatter.TimeFormat = options.timeFormat
	formatter.PrettyPrint = options.prettyPrint
	formatter.IncludeSource = options.includeSource
	formatter.IncludeLevel = options.includeLevel
	formatter.AttributesKey = options.attributesKey
	formatter.ColorOutput = options.colorOutput

	if options.enableColors {
		formatter.ColorScheme = NewColorScheme(options.colorMappings)
		formatter.ColorOutput = true
	}

	return formatter
}

func createXMLFormatter(options *HandlerOptions) *XMLFormatter {
	formatter := NewXMLFormatter()
	formatter.TimeFormat = options.timeFormat
	formatter.IncludeSource = options.includeSource
	formatter.IncludeLevel = options.includeLevel
	formatter.AttributesKey = options.attributesKey

	return formatter
}

func createYAMLFormatter(options *HandlerOptions) *YAMLFormatter {
	formatter := NewYAMLFormatter()
	formatter.TimeFormat = options.timeFormat
	formatter.IncludeSource = options.includeSource
	formatter.IncludeLevel = options.includeLevel
	formatter.AttributesKey = options.attributesKey

	return formatter
}

func createKeyValueFormatter(options *HandlerOptions) *KeyValueFormatter {
	formatter := NewKeyValueFormatter()
	formatter.TimeFormat = options.timeFormat
	formatter.IncludeSource = options.includeSource
	formatter.IncludeLevel = options.includeLevel
	formatter.ColorOutput = options.colorOutput

	if options.enableColors {
		formatter.ColorScheme = NewColorScheme(options.colorMappings)
		formatter.ColorOutput = true
	}

	return formatter
}

// temporaryHandler wraps an existing handler to use a different formatter temporarily
type temporaryHandler struct {
	originalHandler Handler
	formatter       Formatter
}

func (h *temporaryHandler) Handle(ctx context.Context, record *Record) error {
	if !h.originalHandler.Enabled(ctx, record.Level) {
		return nil
	}

	// Format with our temporary formatter
	data, err := h.formatter.Format(record)
	if err != nil {
		return err
	}

	// Get the buffer from the original handler to write to
	var buffer Buffer
	
	switch originalHandler := h.originalHandler.(type) {
	case *TextHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	case *JSONHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	case *XMLHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	case *YAMLHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	case *KeyValueHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	case *BaseHandler:
		originalHandler.mu.RLock()
		buffer = originalHandler.buffer
		originalHandler.mu.RUnlock()
	default:
		// Fallback: use the original handler's Handle method
		return h.originalHandler.Handle(ctx, record)
	}

	// Write to the original handler's buffer
	_, err = buffer.Write(data)
	return err
}

func (h *temporaryHandler) WithAttrs(attrs []slog.Attr) Handler {
	return &temporaryHandler{
		originalHandler: h.originalHandler.WithAttrs(attrs),
		formatter:       h.formatter,
	}
}

func (h *temporaryHandler) WithGroup(name string) Handler {
	return &temporaryHandler{
		originalHandler: h.originalHandler.WithGroup(name),
		formatter:       h.formatter,
	}
}

func (h *temporaryHandler) Enabled(ctx context.Context, level Level) bool {
	return h.originalHandler.Enabled(ctx, level)
}
