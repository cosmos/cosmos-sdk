package iavlx

import (
	"errors"
	"fmt"
)

type ChangesetWriter struct {
	stagedVersion uint32

	files *ChangesetFiles

	kvlog        *KVLogWriter
	branchesData *StructWriter[BranchLayout]
	leavesData   *StructWriter[LeafLayout]
	versionsData *StructWriter[VersionInfo]

	reader *Changeset

	keyCache map[string]uint32
}

func NewChangesetWriter(treeDir string, startVersion uint32, treeStore *TreeStore) (*ChangesetWriter, error) {
	files, err := OpenChangesetFiles(treeDir, startVersion, 0, "")
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}

	cs := &ChangesetWriter{
		stagedVersion: startVersion,
		files:         files,
		kvlog:         NewKVDataWriter(files.kvlogFile),
		branchesData:  NewStructWriter[BranchLayout](files.branchesFile),
		leavesData:    NewStructWriter[LeafLayout](files.leavesFile),
		versionsData:  NewStructWriter[VersionInfo](files.versionsFile),
		reader:        NewChangeset(treeStore),
		keyCache:      make(map[string]uint32),
	}
	return cs, nil
}

func (cs *ChangesetWriter) WriteWALUpdates(updates []KVUpdate) error {
	return cs.kvlog.WriteUpdates(updates)
}

func (cs *ChangesetWriter) WriteWALCommit(version uint32) error {
	return cs.kvlog.WriteCommit(version)
}

func (cs *ChangesetWriter) SaveRoot(root *NodePointer, version, totalLeaves, totalBranches uint32) error {
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

func (cs *ChangesetWriter) CreatedSharedReader() (*Changeset, error) {
	err := cs.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush data before creating shared reader: %w", err)
	}

	err = cs.reader.InitShared(cs.files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shared changeset reader: %w", err)
	}

	reader := cs.reader
	cs.reader = NewChangeset(reader.treeStore)
	return reader, nil
}

func (cs *ChangesetWriter) Flush() error {
	return errors.Join(
		cs.files.infoMmap.Flush(),
		cs.leavesData.Flush(),
		cs.branchesData.Flush(),
		cs.kvlog.Flush(),
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
	keyOffset, ok := cs.keyCache[string(node.key)]
	if !ok {
		var err error
		keyOffset, err = cs.kvlog.WriteK(node.key)
		if err != nil {
			return fmt.Errorf("failed to write key data: %w", err)
		}
	}

	// now write parent
	parentIdx := int64(cs.branchesData.Count() + 1) // +1 to account for the node being written
	leftRef := cs.createNodeRef(parentIdx, node.left)
	rightRef := cs.createNodeRef(parentIdx, node.right)

	layout := BranchLayout{
		Id:            np.id,
		Left:          leftRef,
		Right:         rightRef,
		KeyOffset:     keyOffset,
		Height:        node.height,
		Size:          uint32(node.size), // TODO check overflow
		OrphanVersion: 0,
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err = cs.branchesData.Append(&layout) // TODO check error
	if err != nil {
		return fmt.Errorf("failed to write branch node: %w", err)
	}

	np.fileIdx = uint32(cs.branchesData.Count())
	np.store = cs.reader

	return nil
}

func (cs *ChangesetWriter) writeLeaf(np *NodePointer, node *MemNode) error {
	keyOffset := node.keyOffset
	if keyOffset == 0 {
		var err error
		keyOffset, err = cs.kvlog.WriteKV(node.key, node.value)
		if err != nil {
			return fmt.Errorf("failed to write key-value data: %w", err)
		}
	}

	layout := LeafLayout{
		Id:            np.id,
		KeyOffset:     keyOffset,
		OrphanVersion: 0,
	}
	copy(layout.Hash[:], node.hash) // TODO check length

	err := cs.leavesData.Append(&layout)
	if err != nil {
		return fmt.Errorf("failed to write leaf node: %w", err)
	}

	np.fileIdx = uint32(cs.leavesData.Count())
	np.store = cs.reader

	cs.keyCache[string(node.key)] = keyOffset

	return nil
}

func (cs *ChangesetWriter) createNodeRef(parentIdx int64, np *NodePointer) NodeRef {
	if np.store == cs.reader {
		if np.id.IsLeaf() {
			// return NodeRef(np.id)
			return NodeRef(NewNodeRelativePointer(true, int64(np.fileIdx)))
		} else {
			// for branch nodes the relative offset is the difference between the parent ID index and the branch ID index
			relOffset := int64(np.fileIdx) - parentIdx
			return NodeRef(NewNodeRelativePointer(false, relOffset))
		}
	} else {
		return NodeRef(np.id)
	}
}

func (cs *ChangesetWriter) TotalBytes() int {
	return cs.leavesData.Size() +
		cs.branchesData.Size() +
		cs.versionsData.Size() +
		cs.kvlog.Size()
}

func (cs *ChangesetWriter) Seal() (*Changeset, error) {
	err := cs.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush changeset data: %w", err)
	}

	err = cs.reader.InitOwned(cs.files)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize owned changeset reader: %w", err)
	}
	cs.files = nil
	cs.leavesData = nil
	cs.branchesData = nil
	cs.versionsData = nil
	cs.kvlog = nil
	cs.keyCache = nil
	reader := cs.reader
	cs.reader = nil

	return reader, nil
}

func (cs *ChangesetWriter) StartVersion() uint32 {
	return cs.files.StartVersion()
}
