package iavlx

import (
	"fmt"

	"cosmossdk.io/log"
	"github.com/tidwall/btree"

	snapshottypes "cosmossdk.io/store/snapshots/types"
)

type Importer struct {
	// all of these are indexed by version
	treeStore    *TreeStore
	writers      map[uint32]*ChangesetWriter
	leafCounts   map[uint32]uint32
	branchCounts map[uint32]uint32
	childStack   []importNode
}

type importNode struct {
	id     NodeID
	height int32
	size   uint32
}

func NewImporter(dir string, version uint32, opts Options, logger log.Logger) *Importer {
	ts := &TreeStore{
		dir:           dir,
		changesets:    &btree.Map[uint32, *changesetEntry]{},
		logger:        logger,
		opts:          opts,
		stagedVersion: version,
	}
	return &Importer{
		treeStore:    ts,
		writers:      make(map[uint32]*ChangesetWriter),
		leafCounts:   make(map[uint32]uint32),
		branchCounts: make(map[uint32]uint32),
	}
}

func (i *Importer) ImportNode(node *snapshottypes.SnapshotIAVLItem) error {
	version := uint32(node.Version)
	if version > i.treeStore.stagedVersion {
		return fmt.Errorf("cannot import node for future version %d (staged version %d)", version, i.treeStore.stagedVersion)
	}

	key := node.Key
	if key == nil {
		key = []byte{}
	}

	writer, ok := i.writers[version]
	if !ok {
		var err error
		writer, err = NewChangesetWriter(i.treeStore, version)
		if err != nil {
			return err
		}
		i.writers[version] = writer
	}

	var curImportNode importNode
	if node.Height == 0 {
		// leaf node
		value := node.Value
		if value == nil {
			value = []byte{}
		}

		keyOffset, err := writer.kvlog.WriteKV(key, value)
		if err != nil {
			return fmt.Errorf("failed to write key-value data: %w", err)
		}

		idx := i.leafCounts[version] + 1
		id := NewNodeID(true, uint64(version), idx)
		i.leafCounts[version] = idx

		// TODO compute hash
		leaf := &LeafLayout{
			Id:            id,
			KeyOffset:     keyOffset,
			OrphanVersion: 0,
			Hash:          [32]byte{},
		}
		err = writer.leavesData.Append(leaf)
		if err != nil {
			return fmt.Errorf("failed to write leaf node: %w", err)
		}

		curImportNode = importNode{
			id:     id,
			height: 0,
			size:   1,
		}
	} else {
		// branch node
		keyOffset, err := writer.kvlog.WriteK(key)
		if err != nil {
			return fmt.Errorf("failed to write key data: %w", err)
		}

		idx := i.branchCounts[version] + 1
		id := NewNodeID(false, uint64(version), idx)
		i.branchCounts[version] = idx

		// We build the tree from the bottom-left up. The stack is used to store unresolved left
		// children while constructing right children. When all children are built, the parent can
		// be constructed and the resolved children can be discarded from the stack. Using a stack
		// ensures that we can handle additional unresolved left children while building a right branch.
		//
		// We don't modify the stack until we've verified the built node, to avoid leaving the
		// importer in an inconsistent state when we return an error.
		height := node.Height
		stackSize := len(i.childStack)
		left := i.childStack[stackSize-2]
		right := i.childStack[stackSize-1]
		if stackSize < 2 || left.height != height-1 || right.height != height-1 {
			return fmt.Errorf("invalid child stack for branch node at height %d", height)
		}
		// pop the two children off the stack
		i.childStack = i.childStack[:stackSize-2]

		// TODO left and right IDs
		branch := &BranchLayout{
			Id:            id,
			Left:          NodeRef(left.id),
			Right:         NodeRef(right.id),
			KeyOffset:     keyOffset,
			Height:        uint8(height),
			Size:          left.size + right.size,
			OrphanVersion: 0,
			Hash:          [32]byte{}, // TODO
		}

		err = writer.branchesData.Append(branch)
		if err != nil {
			return fmt.Errorf("failed to write branch node: %w", err)
		}
	}

	i.childStack = append(i.childStack, curImportNode)

	return nil
}

func (i *Importer) Finish() (*CommitTree, error) {
	for version, writer := range i.writers {
		changeset, err := writer.Seal()
		if err != nil {
			return nil, fmt.Errorf("failed to seal changeset for version %d: %w", version, err)
		}
		entry := &changesetEntry{}
		entry.changeset.Store(changeset)
		i.treeStore.changesets.Set(version, entry)
	}

	panic("TODO")
}
