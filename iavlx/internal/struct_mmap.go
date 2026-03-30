package internal

import (
	"fmt"
	"os"
	"sort"
	"unsafe"
)

// StructMmap provides zero-copy read access to an array of fixed-size structs stored in a
// memory-mapped file. The mmap'd bytes are cast directly to a []T slice via unsafe, so reads
// are just pointer dereferences with no deserialization. This is the read counterpart to StructWriter.
//
// TODO: the on-disk format is the native memory layout, which means it is endian-dependent.
// In practice this is fine — all mainstream server hardware (x86, ARM, RISC-V) is little-endian,
// and big-endian systems are essentially non-existent in the Cosmos ecosystem. However, it would
// be unsafe to copy data files between systems of different endianness. We could add a magic
// number file to each tree's data directory (written on creation, checked on load) to detect
// endianness mismatches and produce a clear error instead of silent data corruption.
//
// Alignment is enforced at construction time — the mmap buffer must be aligned to T's alignment
// and its size must be a multiple of sizeof(T) (unless debugBadBuffers is set, which truncates
// trailing bytes for crash recovery scenarios).
type StructMmap[T any] struct {
	items []T
	file  *Mmap
	size  int
}

func NewStructMmap[T any](file *os.File) (*StructMmap[T], error) {
	return NewStructMmapDebug[T](file, false)
}

func NewStructMmapDebug[T any](file *os.File, debugBadBuffers bool) (*StructMmap[T], error) {
	mmap, err := NewMmap(file)
	if err != nil {
		return nil, err
	}

	var zero T
	df := &StructMmap[T]{
		file: mmap,
		size: int(unsafe.Sizeof(zero)),
	}

	buf := mmap.handle
	p := unsafe.Pointer(unsafe.SliceData(mmap.handle))
	align := unsafe.Alignof(zero)
	if uintptr(p)%align != 0 {
		return nil, fmt.Errorf("input buffer is not aligned: %p", p)
	}

	size := df.size
	if len(buf)%size != 0 {
		if debugBadBuffers {
			// update the buffer to be a multiple of the struct size, so that we can still read the valid items and ignore the trailing bytes
			buf = buf[:len(buf)-len(buf)%size]
			p = unsafe.Pointer(unsafe.SliceData(buf))
			// set an error to indicate that the buffer was not a multiple of the struct size, but we are ignoring the trailing bytes
			err = fmt.Errorf("input buffer size is not a multiple of struct size: %d %% %d != 0, ignoring trailing bytes", len(mmap.handle), size)
		} else {
			return nil, fmt.Errorf("input buffer size is not a multiple of struct size: %d %% %d != 0", len(buf), size)
		}
	}
	data := unsafe.Slice((*T)(p), len(buf)/size)
	df.items = data

	return df, err
}

func (df *StructMmap[T]) UnsafeItem(i int) *T {
	return &df.items[i]
}

func (df *StructMmap[T]) Count() int {
	return len(df.items)
}

// BinarySearch finds the smallest index i in [0, Count()) at which f(item) is true,
// assuming that f(item) is false for some prefix of the items and then true for the remainder.
// Returns Count() if no such index exists.
// This is a thin wrapper around sort.Search operating on the mmap'd items.
func (df *StructMmap[T]) BinarySearch(f func(*T) bool) int {
	return sort.Search(len(df.items), func(i int) bool {
		return f(&df.items[i])
	})
}

func (df *StructMmap[T]) TotalBytes() int {
	return df.file.Len()
}

func (df *StructMmap[T]) Close() error {
	return df.file.Close()
}
