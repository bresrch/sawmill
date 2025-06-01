package sawmill

import (
	"bytes"
	"sync"
	"time"
)

// Object pools for performance optimization
var (
	// RecursiveMap pool to reduce allocations with pre-warmed instances (legacy)
	recursiveMapPool = sync.Pool{
		New: newRecursiveMapPooled,
	}

	// FlatAttributes pool for new high-performance attribute system
	flatAttributesPool = sync.Pool{
		New: func() interface{} {
			return NewFlatAttributes()
		},
	}

	// Record pool to reduce allocations
	recordPool = sync.Pool{
		New: func() interface{} {
			return &Record{
				Attributes: NewFlatAttributesFromPool(),
			}
		},
	}

	// Buffer pool for JSON formatting
	bufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 2048)) // Larger initial size for better performance
		},
	}

	// Small buffer pool for keys and small strings
	smallBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 128))
		},
	}
)

// newRecursiveMapPooled creates a new RecursiveMap for the pool
func newRecursiveMapPooled() interface{} {
	return &RecursiveMap{
		children: make(map[string]*RecursiveMap, 8), // Pre-size for common case
		hasValue: false,
	}
}

// NewRecursiveMapFromPool creates a RecursiveMap from the pool
func NewRecursiveMapFromPool() *RecursiveMap {
	rm := recursiveMapPool.Get().(*RecursiveMap)
	rm.reset() // Ensure clean state
	return rm
}

// ReturnRecursiveMapToPool returns a RecursiveMap to the pool
func ReturnRecursiveMapToPool(rm *RecursiveMap) {
	if rm == nil {
		return
	}
	rm.reset() // Clean before returning
	recursiveMapPool.Put(rm)
}

// NewRecordFromPool creates a Record from the pool
func NewRecordFromPool(level Level, msg string) *Record {
	record := recordPool.Get().(*Record)
	record.Level = level
	record.Message = msg
	record.Time = time.Now()
	record.Context = nil
	record.PC = 0
	record.Attributes.reset() // Ensure clean attributes
	return record
}

// NewFlatAttributesFromPool creates a FlatAttributes from the pool
func NewFlatAttributesFromPool() *FlatAttributes {
	attrs := flatAttributesPool.Get().(*FlatAttributes)
	attrs.reset() // Ensure clean state
	return attrs
}

// ReturnFlatAttributesToPool returns a FlatAttributes to the pool
func ReturnFlatAttributesToPool(attrs *FlatAttributes) {
	if attrs == nil {
		return
	}
	attrs.reset() // Clean before returning
	flatAttributesPool.Put(attrs)
}

// ReturnRecordToPool returns a Record to the pool
func ReturnRecordToPool(record *Record) {
	if record == nil {
		return
	}
	// Don't return attributes to pool here - they might be referenced elsewhere
	recordPool.Put(record)
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// ReturnBuffer returns a buffer to the pool
func ReturnBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	bufferPool.Put(buf)
}

// GetSmallBuffer gets a small buffer from the pool
func GetSmallBuffer() *bytes.Buffer {
	return smallBufferPool.Get().(*bytes.Buffer)
}

// ReturnSmallBuffer returns a small buffer to the pool
func ReturnSmallBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	smallBufferPool.Put(buf)
}

// reset clears a RecursiveMap for reuse
func (rm *RecursiveMap) reset() {
	// Return child maps to pool if they exist
	for key, child := range rm.children {
		ReturnRecursiveMapToPool(child)
		delete(rm.children, key)
	}
	rm.value = nil
	rm.hasValue = false
}

// Pool-aware clone method
func (rm *RecursiveMap) CloneFromPool() *RecursiveMap {
	clone := NewRecursiveMapFromPool()
	clone.hasValue = rm.hasValue
	clone.value = rm.value

	for key, child := range rm.children {
		clone.children[key] = child.CloneFromPool()
	}

	return clone
}

// Pool management for graceful shutdown
func DrainPools() {
	// Clear all pools to prevent memory leaks during shutdown
	recursiveMapPool = sync.Pool{New: recursiveMapPool.New}
	flatAttributesPool = sync.Pool{New: flatAttributesPool.New}
	recordPool = sync.Pool{New: recordPool.New}
	bufferPool = sync.Pool{New: bufferPool.New}
	smallBufferPool = sync.Pool{New: smallBufferPool.New}
}
