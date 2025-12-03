package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type FileWriter struct {
	writer  *bufio.Writer
	written int
}

func NewFileWriter(file *os.File) *FileWriter {
	return &FileWriter{
		writer: bufio.NewWriterSize(file, 512*1024 /* 512kb */), // TODO: maybe we can have this as a config option?
	}
}

func (f *FileWriter) Write(p []byte) (n int, err error) {
	n, err = f.writer.Write(p)
	f.written += n
	return n, err
}

func (f *FileWriter) Flush() error {
	if err := f.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

func (f *FileWriter) Size() int {
	return f.written
}

var _ io.Writer = (*FileWriter)(nil)
