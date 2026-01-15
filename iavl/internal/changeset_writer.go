package internal

import (
	"errors"
	"fmt"
)

type ChangesetWriter struct {
	layer uint32

	files *ChangesetFiles

	walWriter    *WALWriter
	kvWriter     *KVDataWriter
	branchesData *StructWriter[BranchLayout]
	leavesData   *StructWriter[LeafLayout]
	layersData   *StructWriter[LayerInfo]

	changeset *Changeset

	lastBranchIdx, lastLeafIdx uint32
}

func NewChangesetWriter(treeDir string, stagedLayer, stagedVersion uint32, treeStore *TreeStore) (*ChangesetWriter, error) {
	files, err := CreateChangesetFiles(treeDir, stagedVersion, 0, true)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	walWriter, err := NewWALWriter(files.WALFile(), uint64(stagedVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL currentWriter: %w", err)
	}

	cs := &ChangesetWriter{
		layer:        stagedLayer,
		files:        files,
		walWriter:    walWriter,
		kvWriter:     NewKVDataWriter(files.KVDataFile()),
		branchesData: NewStructWriter[BranchLayout](files.BranchesFile()),
		leavesData:   NewStructWriter[LeafLayout](files.LeavesFile()),
		layersData:   NewStructWriter[LayerInfo](files.LayersFile()),
		changeset:    NewChangeset(treeStore),
	}
	return cs, nil
}

func (cs *ChangesetWriter) Changeset() *Changeset {
	return cs.changeset
}

func (cs *ChangesetWriter) WALWriter() *WALWriter {
	return cs.walWriter
}

func (cs *ChangesetWriter) SaveLayer(layer uint32, root *NodePointer) error {
	if root == nil {
		return fmt.Errorf("cannot save nil root node")
	}

	cs.lastBranchIdx = 0
	cs.lastLeafIdx = 0

	rootVersion, err := cs.writeNode(root)
	if err != nil {
		return err
	}

	if cs.layer != 0 {
		if layer != cs.layer {
			return fmt.Errorf("invalid layer %d, expected %d", layer, cs.layer)
		}
	}

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
		return fmt.Errorf("failed to write version info: %w", err)
	}

	// Set start version on first successful save
	info := cs.files.info
	if info.StartLayer == 0 {
		info.StartLayer = layer
		info.StartVersion = rootVersion
	}
	info.EndLayer = layer
	info.EndVersion = rootVersion

	cs.layer = layer + 1

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

	node.nodeId = NewNodeID(false, cs.layer, cs.lastBranchIdx+1)
	cs.lastBranchIdx++

	keyOffset, found := cs.walWriter.LookupKeyOffset(node.key)
	var keyInfo BranchKeyInfo
	if !found {
		var err error
		keyOffset, err = cs.kvWriter.WriteKeyBlob(node.key)
		if err != nil {
			return fmt.Errorf("failed to write key data: %w", err)
		}
		keyInfo = keyInfo.SetIsInKVData(true)
	}
	node.keyOffset = keyOffset

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
		ID:      np.id,
		Version: node.version,
		Left:    node.left.id,
		Right:   node.right.id,
		// TODO remove these and overload Left/Right to be offsets when in same changeset
		LeftOffset:  leftOffset,
		RightOffset: rightOffset,
		KeyOffset:   keyOffset,
		KeyInfo:     keyInfo,
		// TODO add inline key prefix
		Height: node.height,
		Size:   NewUint40(uint64(node.size)),
	}
	copy(layout.Hash[:], node.hash) // TODO check length

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
	node.nodeId = NewNodeID(true, cs.layer, cs.lastLeafIdx+1)
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

func (cs *ChangesetWriter) StartLayer() uint32 {
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
		cs.layersData.Flush(),
	)
}

func (cs *ChangesetWriter) Seal() error {
	// TODO
	return nil
}
