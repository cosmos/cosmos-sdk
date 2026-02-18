package internal

import (
	"fmt"
	"os"
	"sort"
	"unsafe"
)

type StructMmap[T any] struct {
	items []T
	file  *Mmap
	size  int
}

func NewStructMmap[T any](file *os.File) (*StructMmap[T], error) {
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
		return nil, fmt.Errorf("input buffer size is not a multiple of struct size: %d %% %d != 0", len(buf), size)
	}
	data := unsafe.Slice((*T)(p), len(buf)/size)
	df.items = data

	return df, nil
}

func (df *StructMmap[T]) UnsafeItem(i uint32) *T {
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
