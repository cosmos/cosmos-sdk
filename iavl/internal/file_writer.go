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

// NewFileWriter creates a new FileWriter with the default buffer size of 512kb
// and initializes the written byte count based on the current file size.
func NewFileWriter(file *os.File) *FileWriter {
	const defaultBufferSize = 512 * 1024 // 512kb
	return NewFileWriterSize(file, defaultBufferSize)
}

// NewFileWriterSize creates a new FileWriter with the specified buffer size
// and initializes the written byte count based on the current file size.
func NewFileWriterSize(file *os.File, size int) *FileWriter {
	var written int
	if info, err := file.Stat(); err == nil {
		written = int(info.Size())
	}
	return &FileWriter{
		file:    file,
		writer:  bufio.NewWriterSize(file, size),
		written: written,
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
	// TODO consider fsyncing the directory as well for extra durability guarantees when the file is newly created

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
