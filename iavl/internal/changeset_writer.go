package internal

import (
	"errors"
	"fmt"
)

type ChangesetWriter struct {
	stagedLayer uint32

	files *ChangesetFiles

	kvlog        *KVDataWriter
	branchesData *StructWriter[BranchLayout]
	leavesData   *StructWriter[LeafLayout]
	layersData   *StructWriter[LayerInfo]

	reader *Changeset

	lastBranchIdx, lastLeafIdx uint32
}

func NewChangesetWriter(treeDir string, stagedLayer uint32, treeStore *TreeStore) (*ChangesetWriter, error) {
	files, err := CreateChangesetFiles(treeDir, stagedLayer, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	cs := &ChangesetWriter{
		stagedLayer:  stagedLayer,
		files:        files,
		kvlog:        NewKVDataWriter(files.KVDataFile()),
		branchesData: NewStructWriter[BranchLayout](files.BranchesFile()),
		leavesData:   NewStructWriter[LeafLayout](files.LeavesFile()),
		layersData:   NewStructWriter[LayerInfo](files.LayersFile()),
		reader:       NewChangeset(treeStore),
	}
	return cs, nil
}

func (cs *ChangesetWriter) SaveLayer(root *NodePointer) (layer uint32, err error) {
	if root == nil {
		return 0, fmt.Errorf("cannot save nil root node")
	}

	cs.lastBranchIdx = 0
	cs.lastLeafIdx = 0

	rootVersion, err := cs.writeNode(root)
	if err != nil {
		return 0, err
	}

	layer = cs.stagedLayer

	var layerInfo LayerInfo
	layerInfo.Branches.StartOffset = uint32(cs.branchesData.Count())
	layerInfo.Leaves.StartOffset = uint32(cs.leavesData.Count())
	layerInfo.RootID = root.id
	layerInfo.Layer = layer
	layerInfo.Version = rootVersion
	totalBranches := cs.lastBranchIdx
	if totalBranches > 0 {
		layerInfo.Branches.StartIndex = 1
		layerInfo.Branches.Count = totalBranches
		layerInfo.Branches.EndIndex = totalBranches
	}
	totalLeaves := cs.lastLeafIdx
	if totalLeaves > 0 {
		layerInfo.Leaves.StartIndex = 1
		layerInfo.Leaves.Count = totalLeaves
		layerInfo.Leaves.EndIndex = totalLeaves
	}

	// commit version info
	err = cs.layersData.Append(&layerInfo)
	if err != nil {
		return 0, fmt.Errorf("failed to write version info: %w", err)
	}

	// Set start version on first successful save
	info := cs.files.info
	if info.StartLayer == 0 {
		info.StartLayer = layer
		info.StartVersion = rootVersion
	}
	info.EndLayer = layer
	info.EndVersion = rootVersion

	cs.stagedLayer++

	return layer, nil
}

func (cs *ChangesetWriter) writeNode(np *NodePointer) (nodeVersion uint32, err error) {
	memNode := np.mem.Load()
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

	cs.lastBranchIdx++
	node.nodeId = NewNodeID(false, cs.stagedLayer, cs.lastBranchIdx)

	keyOffset, err := cs.kvlog.WriteKeyBlob(node.key)
	if err != nil {
		return fmt.Errorf("failed to write key data: %w", err)
	}

	leftVersion := node.left.id.Layer()
	rightVersion := node.right.id.Layer()

	var leftOffset uint32
	var rightOffset uint32

	// If the child node is in the same changeset, store its 1-based file offset.
	// fileIdx is already 1-based (set to Count() after append), and 0 means no offset.
	if leftVersion >= cs.StartLayer() {
		leftOffset = node.left.fileIdx
	}
	if rightVersion >= cs.StartLayer() {
		rightOffset = node.right.fileIdx
	}

	layout := BranchLayout{
		ID:    np.id,
		Left:  node.left.id,
		Right: node.right.id,
		// TODO remove these and overload Left/Right to be offsets when in same changeset
		LeftOffset:  leftOffset,
		RightOffset: rightOffset,
		KeyOffset:   keyOffset,
		// TODO add inline key prefix
		Height: node.height,
		Size:   NewUint40(uint64(node.size)),
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err = cs.branchesData.Append(&layout) // TODO check error
	if err != nil {
		return fmt.Errorf("failed to write branch node: %w", err)
	}

	np.fileIdx = uint32(cs.branchesData.Count())
	np.changeset = cs.reader

	return nil
}

func (cs *ChangesetWriter) writeLeaf(np *NodePointer, node *MemNode) error {
	keyOffset := node.keyOffset
	if node.keyOffset == 0 || node.valueOffset == 0 {
		return fmt.Errorf("leaf node missing key or value offset")
	}

	layout := LeafLayout{
		ID:        np.id,
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
	np.changeset = cs.reader

	return nil
}

//func (cs *ChangesetWriter) TotalBytes() int {
//	return cs.leavesData.Size() +
//		cs.branchesData.Size() +
//		cs.versionsData.Size() +
//		cs.kvlog.Size()
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
//	cs.kvlog = nil
//	cs.keyCache = nil
//	reader := cs.reader
//	cs.reader = nil
//
//	return reader, nil
//}

func (cs *ChangesetWriter) StartLayer() uint32 {
	return cs.files.StartLayer()
}

func (cs *ChangesetWriter) CreatedSharedReader() (*Changeset, error) {
	err := cs.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush data before creating shared reader: %w", err)
	}

	// TODO
	//err = cs.reader.InitShared(cs.files)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to initialize shared changeset reader: %w", err)
	//}

	reader := cs.reader
	cs.reader = NewChangeset(reader.treeStore)
	return reader, nil
}

func (cs *ChangesetWriter) Flush() error {
	return errors.Join(
		cs.files.RewriteInfo(),
		cs.leavesData.Flush(),
		cs.branchesData.Flush(),
		cs.kvlog.Flush(),
		cs.layersData.Flush(),
	)
}
