package internal

import (
	"fmt"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

// Mmap represents a read-only memory map into a file.
type Mmap struct {
	handle mmap.MMap
}

// NewMmap creates a new read-only Mmap for the given file.
func NewMmap(file *os.File) (*Mmap, error) {
	// Check file size
	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Empty files are valid - just don't mmap them
	if fi.Size() == 0 {
		return &Mmap{}, nil
	}

	handle, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to mmap file: %w", err)
	}

	return &Mmap{handle: handle}, nil
}

// At returns the byte at the given index in the mmap-ed data.
// If the index is out of bounds, it panics.
func (m Mmap) At(i int) byte {
	return m.handle[i]
}

// UnsafeSlice returns a byte slice pointing to the mmap-ed data at the given offset and size.
// If the offset and size exceed the mapped data, an error is returned.
// WARNING: The returned byte slice is unsafe and should not be used after the mmap is closed.
func (m Mmap) UnsafeSlice(offset, size int) ([]byte, error) {
	if offset+size > len(m.handle) {
		return nil, fmt.Errorf("trying to read beyond mapped data: %d + %d >= %d", offset, size, len(m.handle))
	}
	bz := m.handle[offset : offset+size]
	return bz, nil
}

// UnsafeSliceVar returns a byte slice pointing to the mmap-ed data at the given offset with a maximum size.
// If the offset exceeds the mapped data, an error is returned.
// If the requested size exceeds the mapped data, it is truncated to fit within the mapped data.
// The number of bytes read is also returned.
// WARNING: The returned byte slice is unsafe and should not be used after the mmap is closed.
func (m Mmap) UnsafeSliceVar(offset, maxSize int) (int, []byte, error) {
	if offset >= len(m.handle) {
		return 0, nil, fmt.Errorf("trying to read beyond mapped data: %d >= %d", offset, len(m.handle))
	}
	if offset+maxSize > len(m.handle) {
		maxSize = len(m.handle) - offset
	}
	data := m.handle[offset : offset+maxSize]
	return maxSize, data, nil
}

// Len returns the length of the mmap-ed data.
func (m Mmap) Len() int {
	return len(m.handle)
}

// Close unmaps the memory-mapped file but does not close the underlying file.
func (m Mmap) Close() error {
	if m.handle == nil {
		return nil
	}
	return m.handle.Unmap()
}

var _ io.Closer = (*Mmap)(nil)
