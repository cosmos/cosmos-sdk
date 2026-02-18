package internal

import (
	"os"
	"unsafe"
)

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
