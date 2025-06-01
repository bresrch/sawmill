package sawmill

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// RecursiveMap represents a recursive key/value map with infinite depth
type RecursiveMap struct {
	children map[string]*RecursiveMap
	value    interface{} // Can hold any type
	hasValue bool        // Indicates if this node has a terminal value
}

// NewRecursiveMap creates a new recursive map
func NewRecursiveMap() *RecursiveMap {
	return &RecursiveMap{
		children: make(map[string]*RecursiveMap),
		hasValue: false,
	}
}

// Set sets a value at the given key path
func (rm *RecursiveMap) Set(keyPath []string, value interface{}) {
	if len(keyPath) == 0 {
		rm.value = value
		rm.hasValue = true
		return
	}

	key := keyPath[0]
	if rm.children[key] == nil {
		rm.children[key] = NewRecursiveMap()
	}

	rm.children[key].Set(keyPath[1:], value)
}

// Get retrieves a value at the given key path
func (rm *RecursiveMap) Get(keyPath []string) (interface{}, bool) {
	if len(keyPath) == 0 {
		return rm.value, rm.hasValue
	}

	key := keyPath[0]
	child := rm.children[key]
	if child == nil {
		return nil, false
	}

	return child.Get(keyPath[1:])
}

// GetNode returns the RecursiveMap node at the given key path
func (rm *RecursiveMap) GetNode(keyPath []string) (*RecursiveMap, bool) {
	if len(keyPath) == 0 {
		return rm, true
	}

	key := keyPath[0]
	child := rm.children[key]
	if child == nil {
		return nil, false
	}

	return child.GetNode(keyPath[1:])
}

// Has checks if a key path exists
func (rm *RecursiveMap) Has(keyPath []string) bool {
	_, exists := rm.Get(keyPath)
	return exists
}

// Delete removes a value at the given key path
func (rm *RecursiveMap) Delete(keyPath []string) bool {
	if len(keyPath) == 0 {
		if rm.hasValue {
			rm.value = nil
			rm.hasValue = false
			return true
		}
		return false
	}

	key := keyPath[0]
	child := rm.children[key]
	if child == nil {
		return false
	}

	if len(keyPath) == 1 {
		// If this is the last key and the child has no children, remove it entirely
		if len(child.children) == 0 {
			delete(rm.children, key)
			return true
		}
		// Otherwise just remove the value
		return child.Delete([]string{})
	}

	return child.Delete(keyPath[1:])
}

// Keys returns all immediate child keys
func (rm *RecursiveMap) Keys() []string {
	keys := make([]string, 0, len(rm.children))
	for key := range rm.children {
		keys = append(keys, key)
	}
	return keys
}

// AllPaths returns all paths that have values
func (rm *RecursiveMap) AllPaths() [][]string {
	var paths [][]string
	rm.collectPaths([]string{}, &paths)
	return paths
}

func (rm *RecursiveMap) collectPaths(currentPath []string, paths *[][]string) {
	if rm.hasValue {
		// Make a copy of the current path
		pathCopy := make([]string, len(currentPath))
		copy(pathCopy, currentPath)
		*paths = append(*paths, pathCopy)
	}

	for key, child := range rm.children {
		newPath := append(currentPath, key)
		child.collectPaths(newPath, paths)
	}
}

// String returns a string representation of the map
func (rm *RecursiveMap) String() string {
	return rm.stringWithIndent(0)
}

func (rm *RecursiveMap) stringWithIndent(indent int) string {
	result := ""
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	if rm.hasValue {
		result += fmt.Sprintf("%svalue: %v (%s)\n", indentStr, rm.value, reflect.TypeOf(rm.value))
	}

	for key, child := range rm.children {
		result += fmt.Sprintf("%s%s:\n", indentStr, key)
		result += child.stringWithIndent(indent + 1)
	}

	return result
}

// Size returns the total number of values stored in the map
func (rm *RecursiveMap) Size() int {
	count := 0
	if rm.hasValue {
		count = 1
	}

	for _, child := range rm.children {
		count += child.Size()
	}

	return count
}

// IsEmpty checks if the map is completely empty
func (rm *RecursiveMap) IsEmpty() bool {
	return !rm.hasValue && len(rm.children) == 0
}

// Clone creates a deep copy of the recursive map
func (rm *RecursiveMap) Clone() *RecursiveMap {
	clone := NewRecursiveMap()
	clone.hasValue = rm.hasValue
	clone.value = rm.value // Note: this is a shallow copy of the value

	for key, child := range rm.children {
		clone.children[key] = child.Clone()
	}

	return clone
}

// SetByDotNotation sets a value using dot notation (e.g., "key1.key2.key3")
func (rm *RecursiveMap) SetByDotNotation(dotPath string, value interface{}) {
	keys := strings.Split(dotPath, ".")
	rm.Set(keys, value)
}

// SetFast is an optimized version for single-level keys
func (rm *RecursiveMap) SetFast(key string, value interface{}) {
	child := rm.children[key]
	if child == nil {
		child = NewRecursiveMapFromPool()
		rm.children[key] = child
	}
	child.value = value
	child.hasValue = true
}

// SetFastDirect sets a value directly without any checks (unsafe but fast)
func (rm *RecursiveMap) SetFastDirect(key string, value interface{}) {
	// Pre-condition: key must not exist in children
	child := NewRecursiveMapFromPool()
	child.value = value
	child.hasValue = true
	rm.children[key] = child
}

// GetByDotNotation retrieves a value using dot notation
func (rm *RecursiveMap) GetByDotNotation(dotPath string) (interface{}, bool) {
	keys := strings.Split(dotPath, ".")
	return rm.Get(keys)
}

// HasByDotNotation checks if a path exists using dot notation
func (rm *RecursiveMap) HasByDotNotation(dotPath string) bool {
	keys := strings.Split(dotPath, ".")
	return rm.Has(keys)
}

// DeleteByDotNotation removes a value using dot notation
func (rm *RecursiveMap) DeleteByDotNotation(dotPath string) bool {
	keys := strings.Split(dotPath, ".")
	return rm.Delete(keys)
}

// Merge combines another RecursiveMap into this one
func (rm *RecursiveMap) Merge(other *RecursiveMap) {
	if other.hasValue {
		rm.value = other.value
		rm.hasValue = true
	}

	for key, child := range other.children {
		if rm.children[key] == nil {
			rm.children[key] = child.CloneFromPool()
		} else {
			rm.children[key].Merge(child)
		}
	}
}

// MergeShallow performs a shallow merge without cloning for read-only operations
func (rm *RecursiveMap) MergeShallow(other *RecursiveMap) {
	if other.hasValue {
		rm.value = other.value
		rm.hasValue = true
	}

	for key, child := range other.children {
		if rm.children[key] == nil {
			rm.children[key] = child // Shallow copy - only safe for read-only use
		} else {
			rm.children[key].MergeShallow(child)
		}
	}
}

// ToMap converts the RecursiveMap to a regular Go map[string]interface{}
func (rm *RecursiveMap) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if rm.hasValue {
		result["_value"] = rm.value
	}

	for key, child := range rm.children {
		if child.IsLeaf() {
			result[key] = child.value
		} else {
			result[key] = child.ToMap()
		}
	}

	return result
}

// IsLeaf checks if this node is a leaf (has a value and no children)
func (rm *RecursiveMap) IsLeaf() bool {
	return rm.hasValue && len(rm.children) == 0
}

// Depth returns the maximum depth of the tree
func (rm *RecursiveMap) Depth() int {
	if len(rm.children) == 0 {
		return 1
	}

	maxDepth := 0
	for _, child := range rm.children {
		depth := child.Depth()
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth + 1
}

// Walk traverses the map and calls the provided function for each value
func (rm *RecursiveMap) Walk(fn func(path []string, value interface{})) {
	rm.walkInternal([]string{}, fn)
}

func (rm *RecursiveMap) walkInternal(currentPath []string, fn func(path []string, value interface{})) {
	if rm.hasValue {
		pathCopy := make([]string, len(currentPath))
		copy(pathCopy, currentPath)
		fn(pathCopy, rm.value)
	}

	for key, child := range rm.children {
		newPath := append(currentPath, key)
		child.walkInternal(newPath, fn)
	}
}

// FromMap creates a RecursiveMap from a regular Go map
func FromMap(data map[string]interface{}) *RecursiveMap {
	rm := NewRecursiveMap()

	for key, value := range data {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			rm.children[key] = FromMap(nestedMap)
		} else {
			rm.children[key] = NewRecursiveMap()
			rm.children[key].value = value
			rm.children[key].hasValue = true
		}
	}

	return rm
}

// MarshalJSON implements json.Marshaler for efficient JSON encoding
func (rm *RecursiveMap) MarshalJSON() ([]byte, error) {
	if rm.IsEmpty() {
		return []byte("{}"), nil
	}

	buf := GetBuffer()
	defer ReturnBuffer(buf)

	buf.WriteByte('{')
	first := true

	// Write the value if this node has one
	if rm.hasValue {
		if !first {
			buf.WriteByte(',')
		}
		buf.WriteString(`"_value":`)
		valueBytes, err := json.Marshal(rm.value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
		first = false
	}

	// Write all children
	for key, child := range rm.children {
		if !first {
			buf.WriteByte(',')
		}

		// Write key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Write value
		if child.IsLeaf() {
			valueBytes, err := json.Marshal(child.value)
			if err != nil {
				return nil, err
			}
			buf.Write(valueBytes)
		} else {
			childBytes, err := child.MarshalJSON()
			if err != nil {
				return nil, err
			}
			buf.Write(childBytes)
		}
		first = false
	}

	buf.WriteByte('}')

	// Copy buffer contents to return slice
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}
