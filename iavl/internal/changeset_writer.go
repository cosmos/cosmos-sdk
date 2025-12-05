package internal

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type ChangesetWriter struct {
	changeset *Changeset
	files     *ChangesetFiles

	stagedVersion uint32

	needsSync atomic.Bool

	kvDataWriter *KVDataWriter
	branchesData *StructWriter[BranchLayout]
	leavesData   *StructWriter[LeafLayout]
	versionsData *StructWriter[VersionInfo]
}

func NewChangesetWriter(changeset *Changeset, treeDir string, startVersion uint32) (*ChangesetWriter, error) {
	files, err := CreateChangesetFiles(treeDir, startVersion, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	cs := &ChangesetWriter{
		files:         files,
		changeset:     changeset,
		stagedVersion: startVersion,
		kvDataWriter:  NewKVDataWriter(files.KVDataFile()),
		branchesData:  NewStructWriter[BranchLayout](files.BranchesFile()),
		leavesData:    NewStructWriter[LeafLayout](files.LeavesFile()),
		versionsData:  NewStructWriter[VersionInfo](files.VersionsFile()),
	}
	return cs, nil
}

//	func (cs *ChangesetWriter) WriteWALUpdates(updates []KVUpdate) error {
//		return cs.kvDataWriter.WriteUpdates(updates)
//	}
//
//	func (cs *ChangesetWriter) WriteWALCommit(version uint32) error {
//		return cs.kvDataWriter.WriteCommit(version)
//	}
func (cs *ChangesetWriter) SaveRoot(root *NodePointer, version, totalLeaves, totalBranches uint32) error {
	cs.needsSync.Store(true)

	if version != cs.stagedVersion {
		return fmt.Errorf("version mismatch: expected %d, got %d", cs.stagedVersion, version)
	}

	var versionInfo VersionInfo
	versionInfo.Branches.StartOffset = uint32(cs.branchesData.Count())
	versionInfo.Leaves.StartOffset = uint32(cs.leavesData.Count())
	if totalBranches > 0 {
		versionInfo.Branches.StartIndex = 1
		versionInfo.Branches.Count = totalBranches
		versionInfo.Branches.EndIndex = totalBranches
	}
	if totalLeaves > 0 {
		versionInfo.Leaves.StartIndex = 1
		versionInfo.Leaves.Count = totalLeaves
		versionInfo.Leaves.EndIndex = totalLeaves
	}

	if root != nil {
		err := cs.writeNode(root)
		if err != nil {
			return err
		}

		versionInfo.RootID = root.id
	}

	// commit version info
	err := cs.versionsData.Append(&versionInfo)
	if err != nil {
		return fmt.Errorf("failed to write version info: %w", err)
	}

	// Set start version on first successful save
	info := cs.files.info
	if info.StartVersion == 0 {
		info.StartVersion = version
	}

	// Always update end version
	info.EndVersion = version

	cs.stagedVersion++

	return nil
}

func (cs *ChangesetWriter) CreatedSharedReader() (*ChangesetReader, error) {
	err := cs.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush data before creating shared reader: %w", err)
	}
	rdr := &ChangesetReader{changeset: cs.changeset}

	err = rdr.InitShared(cs.files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shared changeset reader: %w", err)
	}

	return rdr, nil
}

func (cs *ChangesetWriter) Flush() error {
	return errors.Join(
		cs.files.RewriteInfo(),
		cs.leavesData.Flush(),
		cs.branchesData.Flush(),
		cs.kvDataWriter.Flush(),
		cs.versionsData.Flush(),
	)
}

func (cs *ChangesetWriter) writeNode(np *NodePointer) error {
	memNode := np.mem.Load()
	if memNode == nil {
		return nil // already persisted
	}
	if memNode.version != cs.stagedVersion {
		return nil // not part of this version
	}
	if memNode.IsLeaf() {
		return cs.writeLeaf(np, memNode)
	} else {
		return cs.writeBranch(np, memNode)
	}
}

func (cs *ChangesetWriter) writeBranch(np *NodePointer, node *MemNode) error {
	// recursively write children in post-order traversal
	err := cs.writeNode(node.left)
	if err != nil {
		return err
	}
	err = cs.writeNode(node.right)
	if err != nil {
		return err
	}

	// TODO cache key offset in memory to avoid duplicate writes
	keyOffset, err := cs.kvDataWriter.WriteKeyBlob(node.key)
	if err != nil {
		return fmt.Errorf("failed to write key blob data: %w", err)
	}

	var leftOffset uint32
	var rightOffset uint32

	// If the child node is in the same changeset, store its 1-based file offset.
	// fileIdx is already 1-based (set to Count() after append), and 0 means no offset.
	csStartVersion := cs.StartVersion()
	if node.left.id.Version >= csStartVersion {
		leftOffset = node.left.fileIdx
	}
	if node.right.id.Version >= csStartVersion {
		rightOffset = node.right.fileIdx
	}

	layout := BranchLayout{
		ID:          np.id,
		Left:        node.left.id,
		Right:       node.right.id,
		LeftOffset:  leftOffset,
		RightOffset: rightOffset,
		KeyOffset:   keyOffset,
		Height:      node.height,
		Size:        NewUint40(uint64(node.size)),
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err = cs.branchesData.Append(&layout) // TODO check error
	if err != nil {
		return fmt.Errorf("failed to write branch node: %w", err)
	}

	np.fileIdx = uint32(cs.branchesData.Count())
	np.changeset = cs.changeset

	return nil
}

func (cs *ChangesetWriter) writeLeaf(np *NodePointer, node *MemNode) error {
	keyOffset := node.keyOffset
	valueOffset := node.valueOffset
	if keyOffset == 0 {
		var err error
		keyOffset, valueOffset, err := cs.kvDataWriter.WriteKeyValueBlobs(node.key, node.value)
		if err != nil {
			return fmt.Errorf("failed to write key-value data: %w", err)
		}
		node.keyOffset = keyOffset
		node.valueOffset = valueOffset
	}

	layout := LeafLayout{
		ID:          np.id,
		KeyOffset:   keyOffset,
		ValueOffset: valueOffset,
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err := cs.leavesData.Append(&layout)
	if err != nil {
		return fmt.Errorf("failed to write leaf node: %w", err)
	}

	np.fileIdx = uint32(cs.leavesData.Count())
	np.changeset = cs.changeset

	return nil
}

func (cs *ChangesetWriter) TotalBytes() int {
	return cs.leavesData.Size() +
		cs.branchesData.Size() +
		cs.versionsData.Size() +
		cs.kvDataWriter.Size()
}

func (cs *ChangesetWriter) Seal() (*ChangesetReader, error) {
	err := cs.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush changeset data: %w", err)
	}

	reader := &ChangesetReader{changeset: cs.changeset}
	err = reader.InitOwned(cs.files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize owned changeset reader: %w", err)
	}
	cs.leavesData = nil
	cs.branchesData = nil
	cs.versionsData = nil
	cs.kvDataWriter = nil
	return reader, nil
}

func (cs *ChangesetWriter) StartVersion() uint32 {
	return cs.files.StartVersion()
}

func (cs *ChangesetWriter) SyncWAL() error {
	if !cs.needsSync.CompareAndSwap(true, false) {
		return nil
	}
	return cs.files.kvDataFile.Sync()
}
