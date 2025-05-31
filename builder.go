package sawmill

import (
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
			rm.children[key] = child.Clone()
		} else {
			rm.children[key].Merge(child)
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
