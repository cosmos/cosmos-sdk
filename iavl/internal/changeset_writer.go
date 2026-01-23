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
		changeset:       NewChangeset(treeStore),
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
// The checkpoint identifier equals the version (checkpoint == version).
func (cs *ChangesetWriter) SaveCheckpoint(version uint32, root *NodePointer) error {
	cs.lastBranchIdx = 0
	cs.lastLeafIdx = 0

	// set or validate checkpoint
	if cs.checkpoint != 0 {
		if version <= cs.checkpoint {
			return fmt.Errorf("invalid checkpoint version %d, must be greater than previous %d", version, cs.checkpoint)
		}
	}
	cs.checkpoint = version

	var cpInfo CheckpointInfo
	cpInfo.Branches.StartOffset = uint32(cs.branchesData.Count())
	cpInfo.Leaves.StartOffset = uint32(cs.leavesData.Count())

	if root != nil {
		// it is okay to have a nil root (empty tree)
		_, err := cs.writeNode(root)
		if err != nil {
			return err
		}
		cpInfo.RootID = root.id
	}

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

	// commit checkpoint info
	err := cs.checkpointsData.Append(&cpInfo)
	if err != nil {
		return fmt.Errorf("failed to write checkpoint info: %w", err)
	}

	// Set first checkpoint on first successful save
	info := cs.files.info
	if info.FirstCheckpoint == 0 {
		info.FirstCheckpoint = version
	}
	info.LastCheckpoint = version

	return nil
}

func (cs *ChangesetWriter) writeNode(np *NodePointer) (nodeVersion uint32, err error) {
	memNode := np.Mem.Load()
	if memNode == nil || !memNode.nodeId.IsEmpty() {
		return 0, nil // already persisted
	}
	if memNode.IsLeaf() {
		return memNode.version, cs.writeLeaf(np, memNode)
	} else {
		return memNode.version, cs.writeBranch(np, memNode)
	}
}

func (cs *ChangesetWriter) writeBranch(np *NodePointer, node *MemNode) error {
	// recursively write children in post-order traversal
	_, err := cs.writeNode(node.left)
	if err != nil {
		return err
	}
	_, err = cs.writeNode(node.right)
	if err != nil {
		return err
	}

	node.nodeId = NewNodeID(false, cs.checkpoint, cs.lastBranchIdx+1)
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

	np.id = node.nodeId
	np.fileIdx = uint32(cs.branchesData.Count())
	np.changeset = cs.changeset

	return nil
}

func (cs *ChangesetWriter) writeLeaf(np *NodePointer, node *MemNode) error {
	node.nodeId = NewNodeID(true, cs.checkpoint, cs.lastLeafIdx+1)
	cs.lastLeafIdx++

	keyOffset := node.keyOffset
	if node.keyOffset.IsZero() || node.valueOffset.IsZero() {
		return fmt.Errorf("leaf node missing key or value offset")
	}

	layout := LeafLayout{
		ID:        np.id,
		Version:   node.version,
		KeyOffset: keyOffset,
		// TODO add inline key prefix
		ValueOffset: node.valueOffset,
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err := cs.leavesData.Append(&layout)
	if err != nil {
		return fmt.Errorf("failed to write leaf node: %w", err)
	}

	np.fileIdx = uint32(cs.leavesData.Count())
	np.changeset = cs.changeset
	np.id = node.nodeId

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

func (cs *ChangesetWriter) CreatedSharedReader() error {
	err := cs.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush data before creating shared reader: %w", err)
	}

	rdr, err := NewChangesetReader(cs.changeset, cs.files, false)
	if err != nil {
		return fmt.Errorf("failed to create shared changeset reader: %w", err)
	}

	cs.changeset.SwapActiveReader(rdr)
	return nil
}

func (cs *ChangesetWriter) Flush() error {
	return errors.Join(
		// NOTE: we do not flush the WAL here as that is being done elsewhere
		cs.files.RewriteInfo(),
		cs.leavesData.Flush(),
		cs.branchesData.Flush(),
		cs.kvWriter.Flush(),
		cs.checkpointsData.Flush(),
	)
}

func (cs *ChangesetWriter) Seal(endVersion uint32) error {
	cs.files.Info().WALEndVersion = endVersion
	err := cs.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush changeset data: %w", err)
	}

	rdr, err := NewChangesetReader(cs.changeset, cs.files, false)
	if err != nil {
		return fmt.Errorf("failed to create shared changeset reader: %w", err)
	}

	cs.changeset.SwapActiveReader(rdr)

	// defensively nil out writers to prevent further use
	cs.leavesData = nil
	cs.branchesData = nil
	cs.checkpointsData = nil
	cs.kvWriter = nil
	cs.walWriter = nil

	return nil
}
