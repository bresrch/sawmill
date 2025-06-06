package sawmill

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// FlatAttributes represents a high-performance flat map for log attributes
// This replaces RecursiveMap with a much more efficient implementation
type FlatAttributes struct {
	data map[string]interface{}
	mu   sync.RWMutex

	// Fast path for small attribute counts - avoid map allocations
	smallData [8]struct {
		key   string
		value interface{}
	}
	smallCount int
}

// NewFlatAttributes creates a new FlatAttributes instance
func NewFlatAttributes() *FlatAttributes {
	return &FlatAttributes{
		data: make(map[string]interface{}, 16), // Pre-size more aggressively for common case
	}
}

// Set sets a value at the given key path (converted to dot notation)
func (f *FlatAttributes) Set(keyPath []string, value interface{}) {
	if len(keyPath) == 0 {
		return
	}

	key := strings.Join(keyPath, ".")
	f.SetByDotNotation(key, value)
}

// SetByDotNotation sets a value using dot notation key
func (f *FlatAttributes) SetByDotNotation(dotPath string, value interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.data == nil {
		f.data = make(map[string]interface{}, 16)
	}
	f.data[dotPath] = value
}

// SetFast is an optimized version for single-level keys (no locking for performance)
func (f *FlatAttributes) SetFast(key string, value interface{}) {
	// Fast path: use small array for few attributes to avoid map allocation
	if f.data == nil && f.smallCount < len(f.smallData) {
		// Check if key already exists in small data
		for i := 0; i < f.smallCount; i++ {
			if f.smallData[i].key == key {
				f.smallData[i].value = value
				return
			}
		}
		// Add new key-value pair to small data
		f.smallData[f.smallCount].key = key
		f.smallData[f.smallCount].value = value
		f.smallCount++
		return
	}

	// Slow path: migrate to map if needed
	if f.data == nil {
		f.data = make(map[string]interface{}, 16) // Pre-size more aggressively
		// Migrate small data to map
		for i := 0; i < f.smallCount; i++ {
			f.data[f.smallData[i].key] = f.smallData[i].value
		}
		f.smallCount = 0 // Clear small data after migration
	}
	f.data[key] = value
}

// Get retrieves a value at the given key path
func (f *FlatAttributes) Get(keyPath []string) (interface{}, bool) {
	key := strings.Join(keyPath, ".")
	return f.GetByDotNotation(key)
}

// GetByDotNotation retrieves a value using dot notation
func (f *FlatAttributes) GetByDotNotation(dotPath string) (interface{}, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Check small data first
	for i := 0; i < f.smallCount; i++ {
		if f.smallData[i].key == dotPath {
			return f.smallData[i].value, true
		}
	}

	if f.data == nil {
		return nil, false
	}
	value, exists := f.data[dotPath]
	return value, exists
}

// Has checks if a key path exists
func (f *FlatAttributes) Has(keyPath []string) bool {
	_, exists := f.Get(keyPath)
	return exists
}

// HasByDotNotation checks if a path exists using dot notation
func (f *FlatAttributes) HasByDotNotation(dotPath string) bool {
	_, exists := f.GetByDotNotation(dotPath)
	return exists
}

// Delete removes a value at the given key path
func (f *FlatAttributes) Delete(keyPath []string) bool {
	key := strings.Join(keyPath, ".")
	return f.DeleteByDotNotation(key)
}

// DeleteByDotNotation removes a value using dot notation
func (f *FlatAttributes) DeleteByDotNotation(dotPath string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.data == nil {
		return false
	}

	if _, exists := f.data[dotPath]; exists {
		delete(f.data, dotPath)
		return true
	}
	return false
}

// Keys returns all keys in the flat map
func (f *FlatAttributes) Keys() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil {
		return nil
	}

	keys := make([]string, 0, len(f.data))
	for key := range f.data {
		keys = append(keys, key)
	}
	return keys
}

// AllPaths returns all paths that have values (same as Keys for flat structure)
func (f *FlatAttributes) AllPaths() [][]string {
	keys := f.Keys()
	paths := make([][]string, len(keys))
	for i, key := range keys {
		paths[i] = strings.Split(key, ".")
	}
	return paths
}

// Size returns the total number of key-value pairs
func (f *FlatAttributes) Size() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil {
		return f.smallCount
	}
	return len(f.data) + f.smallCount
}

// IsEmpty checks if the map is empty
func (f *FlatAttributes) IsEmpty() bool {
	return f.Size() == 0
}

// Clone creates a shallow copy of the attributes
func (f *FlatAttributes) Clone() *FlatAttributes {
	f.mu.RLock()
	defer f.mu.RUnlock()

	clone := NewFlatAttributes()
	if f.data != nil {
		for key, value := range f.data {
			clone.data[key] = value
		}
	}
	return clone
}

// CloneFromPool creates a clone using the pool system
func (f *FlatAttributes) CloneFromPool() *FlatAttributes {
	f.mu.RLock()
	defer f.mu.RUnlock()

	clone := NewFlatAttributesFromPool()
	if f.data != nil {
		for key, value := range f.data {
			clone.data[key] = value
		}
	}
	return clone
}

// Merge combines another FlatAttributes into this one
func (f *FlatAttributes) Merge(other *FlatAttributes) {
	if other == nil || other.IsEmpty() {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if f.data == nil {
		f.data = make(map[string]interface{}, len(other.data))
	}

	for key, value := range other.data {
		f.data[key] = value
	}
}

// Walk traverses all key-value pairs and calls the provided function
func (f *FlatAttributes) Walk(fn func(path []string, value interface{})) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Walk small data first
	for i := 0; i < f.smallCount; i++ {
		path := strings.Split(f.smallData[i].key, ".")
		fn(path, f.smallData[i].value)
	}

	if f.data == nil {
		return
	}

	for key, value := range f.data {
		path := strings.Split(key, ".")
		fn(path, value)
	}
}

// ToMap converts to a regular Go map (flat structure)
func (f *FlatAttributes) ToMap() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{}, len(f.data))
	for key, value := range f.data {
		result[key] = value
	}
	return result
}

// ToNestedMap converts flat keys to nested map structure
func (f *FlatAttributes) ToNestedMap() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	for flatKey, value := range f.data {
		parts := strings.Split(flatKey, ".")
		current := result

		// Navigate to the parent container
		for i, part := range parts[:len(parts)-1] {
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}
			var ok bool
			current, ok = current[part].(map[string]interface{})
			if !ok {
				// Conflict: overwrite with map
				current = make(map[string]interface{})
				// Re-navigate from result
				current = result
				for j := 0; j <= i; j++ {
					if j == i {
						current[parts[j]] = make(map[string]interface{})
						current = current[parts[j]].(map[string]interface{})
					} else {
						current = current[parts[j]].(map[string]interface{})
					}
				}
			}
		}

		// Set the final value
		current[parts[len(parts)-1]] = value
	}

	return result
}

// MarshalJSON implements json.Marshaler for efficient JSON encoding
func (f *FlatAttributes) MarshalJSON() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil && f.smallCount == 0 {
		return []byte("{}"), nil
	}

	// Fast path for small data only
	if f.data == nil {
		smallMap := make(map[string]interface{}, f.smallCount)
		for i := 0; i < f.smallCount; i++ {
			smallMap[f.smallData[i].key] = f.smallData[i].value
		}
		return json.Marshal(smallMap)
	}

	// Combined path: merge small data with map
	if f.smallCount > 0 {
		combinedMap := make(map[string]interface{}, len(f.data)+f.smallCount)
		for k, v := range f.data {
			combinedMap[k] = v
		}
		for i := 0; i < f.smallCount; i++ {
			combinedMap[f.smallData[i].key] = f.smallData[i].value
		}
		return json.Marshal(combinedMap)
	}

	// Use direct JSON marshaling of the flat map for performance
	return json.Marshal(f.data)
}

// MarshalNestedJSON creates nested JSON structure from flat keys
func (f *FlatAttributes) MarshalNestedJSON() ([]byte, error) {
	nested := f.ToNestedMap()
	return json.Marshal(nested)
}

// String returns a string representation of the attributes
func (f *FlatAttributes) String() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil || len(f.data) == 0 {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{")
	first := true

	for key, value := range f.data {
		if !first {
			builder.WriteString(", ")
		}
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(fmt.Sprintf("%v", value))
		first = false
	}

	builder.WriteString("}")
	return builder.String()
}

// reset clears the attributes for pool reuse
func (f *FlatAttributes) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clear the map but keep the allocation
	for key := range f.data {
		delete(f.data, key)
	}

	// Clear small data
	for i := 0; i < f.smallCount; i++ {
		f.smallData[i].key = ""
		f.smallData[i].value = nil
	}
	f.smallCount = 0
}

// maskValue applies masking to a field value based on the sawmill tag
func (f *FlatAttributes) maskValue(value interface{}, maskTag string) interface{} {
	if maskTag == "" {
		return value
	}

	strValue := fmt.Sprintf("%v", value)
	if strValue == "" {
		return value
	}

	// Check for mask with number pattern: mask[n]
	re := regexp.MustCompile(`^mask\[(\d+)\]$`)
	if matches := re.FindStringSubmatch(maskTag); len(matches) > 1 {
		// Parse the number of characters to unmask
		unmaskCount, err := strconv.Atoi(matches[1])
		if err != nil || unmaskCount < 0 {
			// Invalid number, default to full masking
			return strings.Repeat("*", len(strValue))
		}

		if unmaskCount >= len(strValue) {
			// Don't mask if unmask count is greater than or equal to string length
			return value
		}

		// Show first n characters, mask the rest
		unmasked := strValue[:unmaskCount]
		masked := strings.Repeat("*", len(strValue)-unmaskCount)
		return unmasked + masked
	}

	// Simple "mask" tag - mask everything
	if maskTag == "mask" {
		return strings.Repeat("*", len(strValue))
	}

	// Unknown mask format, return original value
	return value
}

// ExpandStruct automatically expands struct fields into dot notation
func (f *FlatAttributes) ExpandStruct(prefix string, value interface{}) {
	if value == nil {
		return
	}

	val := reflect.ValueOf(value)
	typ := reflect.TypeOf(value)

	// Handle pointers to structs
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		// Not a struct, store as-is
		f.SetByDotNotation(prefix, value)
		return
	}

	// Expand struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := strings.ToLower(fieldType.Name)
		var fieldKey string
		if prefix != "" {
			fieldKey = prefix + "." + fieldName
		} else {
			fieldKey = fieldName
		}

		fieldValue := field.Interface()

		// Check for sawmill struct tag for masking
		sawmillTag := fieldType.Tag.Get("sawmill")
		
		// Recursively expand nested structs
		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct) {
			f.ExpandStruct(fieldKey, fieldValue)
		} else {
			// Apply masking if sawmill tag contains mask directive
			if strings.HasPrefix(sawmillTag, "mask") {
				fieldValue = f.maskValue(fieldValue, sawmillTag)
			}
			f.SetByDotNotation(fieldKey, fieldValue)
		}
	}
}
