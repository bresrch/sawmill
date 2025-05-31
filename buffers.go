package sawmill

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

// MemoryBuffer implements Buffer interface using in-memory storage
type MemoryBuffer struct {
	buf    *bytes.Buffer
	mu     sync.RWMutex
	maxSize int64
}

// NewMemoryBuffer creates a new memory buffer
func NewMemoryBuffer(maxSize int64) *MemoryBuffer {
	return &MemoryBuffer{
		buf:     &bytes.Buffer{},
		maxSize: maxSize,
	}
}

func (b *MemoryBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.maxSize > 0 && int64(b.buf.Len())+int64(len(p)) > b.maxSize {
		b.buf.Reset()
	}
	
	return b.buf.Write(p)
}

func (b *MemoryBuffer) Flush() error {
	return nil
}

func (b *MemoryBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
	return nil
}

func (b *MemoryBuffer) Size() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return int64(b.buf.Len())
}

func (b *MemoryBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}

func (b *MemoryBuffer) Bytes() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.Bytes()
}

// FileBuffer implements Buffer interface for file-based buffering
type FileBuffer struct {
	writer   *bufio.Writer
	file     *os.File
	mu       sync.RWMutex
	size     int64
	maxSize  int64
	autoSync bool
}

// NewFileBuffer creates a new file buffer
func NewFileBuffer(filename string, bufferSize int, maxSize int64, autoSync bool) (*FileBuffer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &FileBuffer{
		writer:   bufio.NewWriterSize(file, bufferSize),
		file:     file,
		size:     stat.Size(),
		maxSize:  maxSize,
		autoSync: autoSync,
	}, nil
}

func (b *FileBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.maxSize > 0 && b.size+int64(len(p)) > b.maxSize {
		b.file.Truncate(0)
		b.file.Seek(0, 0)
		b.size = 0
		b.writer.Reset(b.file)
	}

	n, err := b.writer.Write(p)
	if err != nil {
		return n, err
	}

	b.size += int64(n)

	if b.autoSync {
		b.writer.Flush()
		b.file.Sync()
	}

	return n, nil
}

func (b *FileBuffer) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.writer.Flush()
	if err != nil {
		return err
	}
	return b.file.Sync()
}

func (b *FileBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if err := b.writer.Flush(); err != nil {
		b.file.Close()
		return err
	}
	
	return b.file.Close()
}

func (b *FileBuffer) Size() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *FileBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.file.Truncate(0)
	b.file.Seek(0, 0)
	b.size = 0
	b.writer.Reset(b.file)
}

// RotatingFileBuffer implements Buffer with file rotation
type RotatingFileBuffer struct {
	basePath    string
	maxSize     int64
	maxFiles    int
	current     *FileBuffer
	mu          sync.RWMutex
	bufferSize  int
	rotateCount int
}

// NewRotatingFileBuffer creates a rotating file buffer
func NewRotatingFileBuffer(basePath string, maxSize int64, maxFiles int, bufferSize int) (*RotatingFileBuffer, error) {
	rb := &RotatingFileBuffer{
		basePath:   basePath,
		maxSize:    maxSize,
		maxFiles:   maxFiles,
		bufferSize: bufferSize,
	}

	err := rb.rotate()
	return rb, err
}

func (b *RotatingFileBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.current.Size()+int64(len(p)) > b.maxSize {
		if err := b.rotate(); err != nil {
			return 0, err
		}
	}

	return b.current.Write(p)
}

func (b *RotatingFileBuffer) Flush() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current.Flush()
}

func (b *RotatingFileBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.current.Close()
}

func (b *RotatingFileBuffer) Size() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current.Size()
}

func (b *RotatingFileBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current.Reset()
}

func (b *RotatingFileBuffer) rotate() error {
	if b.current != nil {
		b.current.Close()
	}

	b.rotateCount++
	
	// Remove old files if we exceed maxFiles
	if b.maxFiles > 0 && b.rotateCount > b.maxFiles {
		oldFile := b.getRotatedFilename(b.rotateCount - b.maxFiles)
		os.Remove(oldFile)
	}

	filename := b.getRotatedFilename(b.rotateCount)
	var err error
	b.current, err = NewFileBuffer(filename, b.bufferSize, 0, false)
	return err
}

func (b *RotatingFileBuffer) getRotatedFilename(count int) string {
	if count == 1 {
		return b.basePath
	}
	return b.basePath + "." + time.Now().Format("20060102-150405")
}

// WriterBuffer wraps any io.Writer as a Buffer
type WriterBuffer struct {
	writer io.Writer
	size   int64
	mu     sync.RWMutex
}

// NewWriterBuffer creates a buffer from an io.Writer
func NewWriterBuffer(writer io.Writer) *WriterBuffer {
	return &WriterBuffer{
		writer: writer,
	}
}

func (b *WriterBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.writer.Write(p)
	b.size += int64(n)
	return n, err
}

func (b *WriterBuffer) Flush() error {
	if flusher, ok := b.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

func (b *WriterBuffer) Close() error {
	if closer, ok := b.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (b *WriterBuffer) Size() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *WriterBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.size = 0
}