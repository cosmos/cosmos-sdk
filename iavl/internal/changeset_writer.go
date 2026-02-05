package internal

import (
	"errors"
	"fmt"
)

type ChangesetWriter struct {
	// checkpoint is the current checkpoint version being written (checkpoint == version)
	checkpoint uint32

	files *ChangesetFiles

	walWriter       *WALWriter
	kvWriter        *KVDataWriter
	branchesData    *StructWriter[BranchLayout]
	leavesData      *StructWriter[LeafLayout]
	checkpointsData *StructWriter[CheckpointInfo]

	changeset *Changeset

	lastBranchIdx, lastLeafIdx uint32
	memUsage                   int64
}

func NewChangesetWriter(treeDir string, stagedVersion uint32, treeStore *TreeStore) (*ChangesetWriter, error) {
	files, err := CreateChangesetFiles(treeDir, stagedVersion, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	cs := &ChangesetWriter{
		files:           files,
		walWriter:       NewWALWriter(files.WALFile()),
		kvWriter:        NewKVDataWriter(files.KVDataFile()),
		branchesData:    NewStructWriter[BranchLayout](files.BranchesFile()),
		leavesData:      NewStructWriter[LeafLayout](files.LeavesFile()),
		checkpointsData: NewStructWriter[CheckpointInfo](files.CheckpointsFile()),
		changeset:       NewChangeset(treeStore, files),
	}
	// Create an initial reader so the changeset is immediately readable (with zero entries).
	// This avoids a race where the changeset is in the version map but has no reader
	// because the async checkpointer hasn't called CreateReader() yet.
	if err := cs.CreateReader(); err != nil {
		return nil, fmt.Errorf("failed to create initial changeset reader: %w", err)
	}
	return cs, nil
}

func (cs *ChangesetWriter) Changeset() *Changeset {
	return cs.changeset
}

func (cs *ChangesetWriter) WALWriter() *WALWriter {
	return cs.walWriter
}

// SaveCheckpoint persists the tree state at the given version.
func (cs *ChangesetWriter) SaveCheckpoint(checkpoint, version uint32, root *NodePointer) error {
	cs.lastBranchIdx = 0
	cs.lastLeafIdx = 0

	// set or validate checkpoint
	if cs.checkpoint != 0 {
		if checkpoint != cs.checkpoint+1 {
			return fmt.Errorf("invalid checkpoint %d, must be one greater than previous %d", checkpoint, cs.checkpoint)
		}
	}
	cs.checkpoint = checkpoint

	var cpInfo CheckpointInfo
	cpInfo.Branches.StartOffset = uint32(cs.branchesData.Count())
	cpInfo.Leaves.StartOffset = uint32(cs.leavesData.Count())

	if root != nil {
		// it is okay to have a nil root (empty tree)
		err := cs.writeNode(root)
		if err != nil {
			return err
		}
		cpInfo.RootID = root.id
	}

	cpInfo.Checkpoint = checkpoint
	cpInfo.Version = version
	totalBranches := cs.lastBranchIdx
	if totalBranches > 0 {
		cpInfo.Branches.StartIndex = 1
		cpInfo.Branches.Count = totalBranches
		cpInfo.Branches.EndIndex = totalBranches
	}
	totalLeaves := cs.lastLeafIdx
	if totalLeaves > 0 {
		cpInfo.Leaves.StartIndex = 1
		cpInfo.Leaves.Count = totalLeaves
		cpInfo.Leaves.EndIndex = totalLeaves
	}

	// file integrity check data
	cpInfo.KVEndOffset.Set(uint64(cs.kvWriter.Size()), true)
	cpInfo.SetCRC32()

	// commit checkpoint info
	err := cs.checkpointsData.Append(&cpInfo)
	if err != nil {
		return fmt.Errorf("failed to write checkpoint info: %w", err)
	}

	return nil
}

func (cs *ChangesetWriter) writeNode(np *NodePointer) error {
	memNode := np.Mem.Load()
	if memNode == nil || memNode.nodeId.Checkpoint() != cs.checkpoint {
		if np.changeset == nil {
			return fmt.Errorf("fatal logic error: trying to save checkpoint %d, but node %v from a different checkpoint is present and has no changeset", cs.checkpoint, np.id)
		}
		return nil // already persisted or nothing to write
	}
	// track memory usage of every node that gets saved to a checkpoint as these are nodes we can evict
	cs.memUsage += memNodeOverhead + int64(len(memNode.key)) + int64(len(memNode.hash))
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

	cs.lastBranchIdx++

	keyOffset, keyInWal := cs.walWriter.LookupKeyOffset(node.key)
	if !keyInWal {
		var err error
		keyOffset, err = cs.kvWriter.WriteKeyBlob(node.key)
		if err != nil {
			return fmt.Errorf("failed to write key data: %w", err)
		}
	}
	node.keyOffset.Set(keyOffset, !keyInWal)

	leftCheckpoint := node.left.id.Checkpoint()
	rightCheckpoint := node.right.id.Checkpoint()

	var leftOffset uint32
	var rightOffset uint32

	// If the child node is in the same changeset, store its 1-based file offset.
	// fileIdx is already 1-based (set to Count() after append), and 0 means no offset.
	if leftCheckpoint >= cs.StartVersion() {
		leftOffset = node.left.fileIdx
	}
	if rightCheckpoint >= cs.StartVersion() {
		rightOffset = node.right.fileIdx
	}

	layout := BranchLayout{
		ID:      np.id,
		Version: node.version,
		Left:    node.left.id,
		Right:   node.right.id,
		// TODO remove these and overload Left/Right to be offsets when in same changeset
		LeftOffset:  leftOffset,
		RightOffset: rightOffset,
		KeyOffset:   node.keyOffset,
		Height:      node.height,
		Size:        uint32(node.size),
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	keyLen := len(node.key)
	layout.InlineKeyLen = InlineKeyPrefixLen(keyLen)
	inlineCopyLen := InlineKeyCopyLen(keyLen)
	copy(layout.InlineKeyPrefix[:], node.key[:inlineCopyLen])

	err = cs.branchesData.Append(&layout) // TODO check error
	if err != nil {
		return fmt.Errorf("failed to write branch node: %w", err)
	}

	np.fileIdx = uint32(cs.branchesData.Count())
	np.changeset = cs.changeset

	return nil
}

func (cs *ChangesetWriter) writeLeaf(np *NodePointer, node *MemNode) error {
	cs.lastLeafIdx++

	// key and value offsets can be missing if we replayed from WAL
	keyOffset := node.keyOffset
	if node.keyOffset.IsZero() {
		offset, found := cs.walWriter.LookupKeyOffset(node.key)
		if found {
			keyOffset = NewKVOffset(offset, false)
		} else {
			offset, err := cs.kvWriter.WriteKeyBlob(node.key)
			if err != nil {
				return fmt.Errorf("failed to write key data: %w", err)
			}
			keyOffset = NewKVOffset(offset, true)
		}
	}

	if node.valueOffset.IsZero() {
		offset, err := cs.kvWriter.WriteValueBlob(node.value)
		if err != nil {
			return fmt.Errorf("failed to write value data: %w", err)
		}
		node.valueOffset = NewKVOffset(offset, true)
	}

	layout := LeafLayout{
		ID:          np.id,
		Version:     node.version,
		KeyOffset:   keyOffset,
		ValueOffset: node.valueOffset,
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

//func (cs *ChangesetWriter) TotalBytes() int {
//	return cs.leavesData.Size() +
//		cs.branchesData.Size() +
//		cs.versionsData.Size() +
//		cs.kvWriter.Size()
//}
//
//func (cs *ChangesetWriter) Seal() (*Changeset, error) {
//	err := cs.Flush()
//	if err != nil {
//		return nil, fmt.Errorf("failed to flush changeset data: %w", err)
//	}
//
//	err = cs.reader.InitOwned(cs.files)
//	if err != nil {
//		return nil, fmt.Errorf("failed to initialize owned changeset reader: %w", err)
//	}
//	cs.leavesData = nil
//	cs.branchesData = nil
//	cs.versionsData = nil
//	cs.kvWriter = nil
//	cs.keyCache = nil
//	reader := cs.reader
//	cs.reader = nil
//
//	return reader, nil
//}

// StartVersion returns the start version of this changeset.
func (cs *ChangesetWriter) StartVersion() uint32 {
	return cs.files.StartVersion()
}

func (cs *ChangesetWriter) CreateReader() error {
	err := cs.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush data before creating shared reader: %w", err)
	}

	rdr, err := NewChangesetReader(cs.changeset)
	if err != nil {
		return fmt.Errorf("failed to create shared changeset reader: %w", err)
	}

	cs.changeset.swapActiveReader(rdr)
	return nil
}

func (cs *ChangesetWriter) Flush() error {
	return errors.Join(
		// NOTE: we do not flush the WAL here as that is being done elsewhere
		cs.leavesData.Flush(),
		cs.branchesData.Flush(),
		cs.kvWriter.Flush(),
		cs.checkpointsData.Flush(),
	)
}

func (cs *ChangesetWriter) Seal() error {
	err := cs.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush changeset data: %w", err)
	}

	cs.changeset.sealed.Store(true)

	// defensively nil out writers to prevent further use
	cs.leavesData = nil
	cs.branchesData = nil
	cs.checkpointsData = nil
	cs.kvWriter = nil
	cs.walWriter = nil

	return nil
}
