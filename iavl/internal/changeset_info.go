package internal

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

// ChangesetInfo holds metadata about a changeset.
// This mainly tracks the start and end version of the changeset and also contains statistics about orphans in the
// changeset so that compaction can be efficiently scheduled.
// Currently, the orphan statistics track how many total leaf and branch orphans there are as well as the total sum
// of their orphan versions.
// This should give us some heuristics as to what percentage of the changeset is composed of
// orphans and roughly how long ago they were orphaned.
type ChangesetInfo struct {
	// StartVersion is the first version included in the changeset.
	StartVersion uint32
	// EndVersion is the last version included in the changeset.
	EndVersion uint32

	// LeafOrphans is the number of leaf orphan nodes in the changeset.
	LeafOrphans uint32
	// BranchOrphans is the number of branch orphan nodes in the changeset.
	BranchOrphans uint32

	// LeafOrphanVersionTotal is the sum of the orphan versions of all orphaned leaf nodes in the changeset.
	LeafOrphanVersionTotal uint64
	// BranchOrphanVersionTotal is the sum of the orphan versions of all orphaned branch nodes in the changeset.
	BranchOrphanVersionTotal uint64
}

// RewriteChangesetInfo truncates and rewrites the info file with the given changeset info.
// If the file does not exist, it will be created.
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

// ReadChangesetInfo reads changeset info from a file. Returns an empty default struct if file is zero length
// or doesn't exist.
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

	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek info file: %w", err)
	}

	// Read directly into the struct
	data := unsafe.Slice((*byte)(unsafe.Pointer(&info)), size)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("failed to read changeset info: %w", err)
	}

	return &info, nil
}
