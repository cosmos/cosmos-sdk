package internal

import (
	"os"
	"unsafe"
)

// StructWriter appends fixed-size structs to a file using unsafe pointer casting.
// It's the write counterpart to StructMmap — data written by StructWriter is later read
// back via StructMmap (memory-mapped). The struct is written as raw bytes with no encoding
// overhead, so the on-disk format is the same as the in-memory layout.
// Used for LeafLayout, BranchLayout, CheckpointInfo, and OrphanEntry.
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

func NewStructWriterSize[T any](file *os.File, bufSize int) *StructWriter[T] {
	fw := NewFileWriterSize(file, bufSize)

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
