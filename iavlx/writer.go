package iavlx

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unsafe"
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

type StructWriter[T any] struct {
	size int
	*FileWriter
}

func NewStructWriter[T any](file *os.File) *StructWriter[T] {
	fw := NewFileWriter(file)

	return &StructWriter[T]{
		size:       int(unsafe.Sizeof(*new(T))),
		FileWriter: fw,
	}
}

func (sw *StructWriter[T]) Append(x *T) error {
	_, err := sw.Write(unsafe.Slice((*byte)(unsafe.Pointer(x)), sw.size))
	return err
}

func (sw *StructWriter[T]) Count() int {
	return sw.written / sw.size
}
