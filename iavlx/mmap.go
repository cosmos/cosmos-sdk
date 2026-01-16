package iavlx

import (
	"fmt"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

type MmapFile struct {
	handle mmap.MMap
}

func NewMmapFile(file *os.File) (*MmapFile, error) {
	// Check file size
	fi, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	res := &MmapFile{}

	// Empty files are valid - just don't mmap them
	if fi.Size() == 0 {
		return res, nil
	}

	// maybe we can make read/write configurable? not sure if the OS optimizes read-only mapping
	handle, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to mmap file: %w", err)
	}

	res.handle = handle
	return res, nil
}

func (m *MmapFile) UnsafeSliceVar(offset, maxSize int) (int, []byte, error) {
	if offset >= len(m.handle) {
		return 0, nil, fmt.Errorf("trying to read beyond mapped data: %d >= %d", offset, len(m.handle))
	}
	if offset+maxSize > len(m.handle) {
		maxSize = len(m.handle) - offset
	}
	data := m.handle[offset : offset+maxSize]
	// make a copy of the data to avoid data being changed after remap
	return maxSize, data, nil
}

func (m *MmapFile) UnsafeSliceExact(offset, size int) ([]byte, error) {
	if offset+size > len(m.handle) {
		return nil, fmt.Errorf("trying to read beyond mapped data: %d + %d >= %d", offset, size, len(m.handle))
	}
	bz := m.handle[offset : offset+size]
	return bz, nil
}

func (m *MmapFile) Data() []byte {
	return m.handle
}

func (m *MmapFile) Close() error {
	if m.handle != nil {
		handle := m.handle
		m.handle = nil
		return handle.Unmap()
	}
	return nil
}

func (m *MmapFile) TotalBytes() int {
	return len(m.handle)
}

var _ io.Closer = &MmapFile{}
