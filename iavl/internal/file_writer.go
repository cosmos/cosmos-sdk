package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// FileWriter is a buffered currentWriter that tracks the number of bytes written.
type FileWriter struct {
	file    *os.File
	writer  *bufio.Writer
	written int
}

// NewFileWriter creates a new FileWriter.
// Currently, it uses a buffer size of 512kb.
// If we want to make that configurable, we can add a constructor with a buffer size parameter in the future.
func NewFileWriter(file *os.File) *FileWriter {
	const defaultBufferSize = 512 * 1024 // 512kb
	// TODO should we make sure the file is at offset 0 and empty?
	return &FileWriter{
		file:   file,
		writer: bufio.NewWriterSize(file, defaultBufferSize),
	}
}

// Write writes data to the underlying buffered currentWriter and updates the written byte count.
func (f *FileWriter) Write(p []byte) (n int, err error) {
	n, err = f.writer.Write(p)
	f.written += n
	return n, err
}

// Flush flushes the underlying buffered currentWriter.
func (f *FileWriter) Flush() error {
	if err := f.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush currentWriter: %w", err)
	}
	return nil
}

func (f *FileWriter) Sync() error {
	err := f.Flush()
	if err != nil {
		return err
	}
	err = f.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	return nil
}

// Size returns the total number of bytes written so far.
func (f *FileWriter) Size() int {
	return f.written
}

var _ io.Writer = (*FileWriter)(nil)
