package iavlx

import (
	"fmt"
	"sync/atomic"

	"cosmossdk.io/log"
	"github.com/tidwall/btree"

	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
)

type Importer struct {
	// all of these are indexed by version
	treeStore    *TreeStore
	writers      map[uint32]*ChangesetWriter
	leafCounts   map[uint32]uint32
	branchCounts map[uint32]uint32
	childStack   []*importedNode
}

type importedNode struct {
	id     NodeID
	height int32
	size   uint32
	key    []byte
	value  []byte
	hash   []byte
}

func (i importedNode) Height() uint8 { return uint8(i.height) }

func (i importedNode) Size() int64 { return int64(i.size) }

func (i importedNode) Version() uint32 { return uint32(i.id.Version()) }

func (i importedNode) Key() ([]byte, error) { return i.key, nil }

func (i importedNode) Value() ([]byte, error) { return i.value, nil }

func (i importedNode) IsLeaf() bool { return i.height == 0 }

var _ hashableNode = (*importedNode)(nil)

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

	var curImportNode *importedNode
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

		curImportNode = &importedNode{
			id:     id,
			height: 0,
			size:   1,
			key:    key,
			value:  value,
		}
		hash, err := computeHash(curImportNode, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}
		curImportNode.hash = hash

		leaf := &LeafLayout{
			Id:        id,
			KeyOffset: keyOffset,
		}
		copy(leaf.Hash[:], hash)
		err = writer.leavesData.Append(leaf)
		if err != nil {
			return fmt.Errorf("failed to write leaf node: %w", err)
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

		size := left.size + right.size + 1
		curImportNode = &importedNode{
			id:     id,
			height: height,
			size:   size,
			key:    key,
		}
		hash, err := computeHash(curImportNode, left.hash, right.hash)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}
		curImportNode.hash = hash

		branch := &BranchLayout{
			Id:            id,
			Left:          NodeRef(left.id),
			Right:         NodeRef(right.id),
			KeyOffset:     keyOffset,
			Height:        uint8(height),
			Size:          left.size + right.size,
			OrphanVersion: 0,
		}
		copy(branch.Hash[:], hash)
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

	i.treeStore.savedVersion.Store(i.treeStore.stagedVersion)
	i.treeStore.stagedVersion++
	err := i.treeStore.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tree store: %w", err)
	}

	ct := &CommitTree{
		latest:           atomic.Pointer[NodePointer]{},
		root:             nil,
		version:          i.treeStore.savedVersion.Load(),
		store:            i.treeStore,
		zeroCopy:         i.treeStore.opts.ZeroCopy,
		evictionDepth:    i.treeStore.opts.EvictDepth,
		evictorRunning:   false,
		lastEvictVersion: 0,
		writeWal:         i.treeStore.opts.WriteWAL,
		walChan:          nil, // TODO
		walDone:          nil, // TODO
		pendingOrphans:   nil,
		logger:           i.treeStore.logger,
		lastCommitId:     storetypes.CommitID{},
		commitCtx:        nil,
	}
	return ct, nil
}
