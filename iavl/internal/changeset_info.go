package internal

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

const (
	sizeChangesetInfo = 32
)

func init() {
	// Verify the size of ChangesetInfo is what we expect it to be at runtime.
	if unsafe.Sizeof(ChangesetInfo{}) != sizeChangesetInfo {
		panic(fmt.Sprintf("invalid ChangesetInfo size: got %d, want %d", unsafe.Sizeof(ChangesetInfo{}), sizeChangesetInfo))
	}
}

// ChangesetInfo holds metadata about a changeset.
// This mainly tracks the start and end version of the changeset and also contains statistics about orphans in the
// changeset so that compaction can be efficiently scheduled.
// Currently, the orphan statistics track how many total leaf and branch orphans there are as well as the total sum
// of their orphan versions.
// This should give us some heuristics as to what percentage of the changeset is composed of
// orphans and roughly how long ago they were orphaned.
type ChangesetInfo struct {
	// StartVersion is the first version with a root in the changeset.
	StartVersion uint32
	// EndVersion is the last version with a root in the changeset.
	EndVersion uint32

	StartLayer uint32
	EndLayer   uint32

	// LeafOrphans is the number of leaf orphan nodes in the changeset.
	LeafOrphans uint32
	// BranchOrphans is the number of branch orphan nodes in the changeset.
	BranchOrphans uint32

	// LeafOrphanVersionTotal is the sum of the orphan versions of all orphaned leaf nodes in the changeset.
	LeafOrphanVersionTotal uint64
	// BranchOrphanVersionTotal is the sum of the orphan versions of all orphaned branch nodes in the changeset.
	BranchOrphanVersionTotal uint64
}

// RewriteChangesetInfo rewrites the info file with the given changeset info.
// This method is okay to call the first time the file is created as well.
func RewriteChangesetInfo(file *os.File, info *ChangesetInfo) error {
	data := unsafe.Slice((*byte)(unsafe.Pointer(info)), sizeChangesetInfo)
	if _, err := file.WriteAt(data, 0); err != nil {
		return fmt.Errorf("failed to write changeset info: %w", err)
	}

	return nil
}

// ReadChangesetInfo reads changeset info from a file. It returns an empty default struct if file is empty.
func ReadChangesetInfo(file *os.File) (*ChangesetInfo, error) {
	var info ChangesetInfo
	data := unsafe.Slice((*byte)(unsafe.Pointer(&info)), sizeChangesetInfo)

	n, err := file.ReadAt(data, 0)
	if err == io.EOF && n == 0 {
		return &ChangesetInfo{}, nil // empty file
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read changeset info: %w", err)
	}
	if n != sizeChangesetInfo {
		return nil, fmt.Errorf("info file has unexpected size: %d, expected %d", n, sizeChangesetInfo)
	}

	return &info, nil
}
