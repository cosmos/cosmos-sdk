package iavlx

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

type ChangesetInfo struct {
	StartVersion             uint32
	EndVersion               uint32
	LeafOrphans              uint32
	BranchOrphans            uint32
	LeafOrphanVersionTotal   uint64
	BranchOrphanVersionTotal uint64
}

// RewriteChangesetInfo truncates and rewrites the info file with the given changeset info.
func RewriteChangesetInfo(file *os.File, info *ChangesetInfo) error {
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate info file: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek info file: %w", err)
	}

	size := int(unsafe.Sizeof(*info))
	data := unsafe.Slice((*byte)(unsafe.Pointer(info)), size)
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write changeset info: %w", err)
	}

	return nil
}

// ReadChangesetInfo reads changeset info from a file. Returns an empty default struct if file is zero length.
func ReadChangesetInfo(file *os.File) (*ChangesetInfo, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat info file: %w", err)
	}

	if stat.Size() == 0 {
		return &ChangesetInfo{}, nil
	}

	var info ChangesetInfo
	size := int(unsafe.Sizeof(info))

	if stat.Size() != int64(size) {
		return nil, fmt.Errorf("info file has unexpected size: %d, expected %d", stat.Size(), size)
	}

	buf := make([]byte, size)
	if _, err := io.ReadFull(file, buf); err != nil {
		return nil, fmt.Errorf("failed to read changeset info: %w", err)
	}

	return (*ChangesetInfo)(unsafe.Pointer(&buf[0])), nil
}
