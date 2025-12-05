package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileWriter is a buffered writer that tracks the number of bytes written.
type FileWriter struct {
	writer  *bufio.Writer
	written int
}

// NewFileWriter creates a new FileWriter.
// Currently, it uses a buffer size of 512kb.
// If we want to make that configurable, we can add a constructor with a buffer size parameter in the future.
func NewFileWriter(file *os.File) *FileWriter {
	const defaultBufferSize = 512 * 1024 // 512kb
	return &FileWriter{
		writer: bufio.NewWriterSize(file, defaultBufferSize),
	}
}

// Write writes data to the underlying buffered writer and updates the written byte count.
func (f *FileWriter) Write(p []byte) (n int, err error) {
	n, err = f.writer.Write(p)
	f.written += n
	return n, err
}

// Flush flushes the underlying buffered writer.
func (f *FileWriter) Flush() error {
	if err := f.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

// Size returns the total number of bytes written so far.
func (f *FileWriter) Size() int {
	return f.written
}

var _ io.Writer = (*FileWriter)(nil)
