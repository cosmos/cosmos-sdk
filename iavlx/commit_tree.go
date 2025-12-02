package iavlx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
)

type CommitTree struct {
	latest     atomic.Pointer[NodePointer]
	root       *NodePointer
	writeMutex sync.Mutex
	store      *TreeStore
	zeroCopy   bool

	walQueue *NonBlockingQueue[[]KVUpdate]
	walDone  <-chan error

	pendingOrphans [][]NodeID

	logger *slog.Logger

	lastCommitId storetypes.CommitID
	commitCtx    *commitContext
}

func NewCommitTree(ctx context.Context, dir string, opts Options, logger *slog.Logger, memMonitor *memoryMonitor) (*CommitTree, error) {
	ts, err := newTreeStore(ctx, dir, opts, logger, memMonitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree store: %w", err)
	}

	var root *NodePointer
	var lastCommitId storetypes.CommitID
	savedVersion := ts.SavedVersion()
	if savedVersion > 0 {
		root, err = ts.ResolveRoot(savedVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve root for saved version %d: %w", savedVersion, err)
		}
		if root != nil {
			rootNode, err := root.Resolve()
			if err != nil {
				return nil, fmt.Errorf("failed to resolve root node for saved version %d: %w", savedVersion, err)
			}
			hash := rootNode.Hash()
			lastCommitId = storetypes.CommitID{
				Version: int64(savedVersion),
				Hash:    hash,
			}
		}
	}

	tree := &CommitTree{
		store:        ts,
		root:         root,
		lastCommitId: lastCommitId,
		zeroCopy:     opts.ZeroCopy,
		logger:       logger,
	}
	tree.latest.Store(root)
	tree.reinitWalProc()

	return tree, nil
}

func (c *CommitTree) WorkingHash() []byte {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	return c.workingHash()
}

func (c *CommitTree) workingHash() []byte {
	// IMPORTANT: this function assumes the write lock is held

	// if we have no root, return empty hash
	if c.root == nil {
		c.commitCtx = nil
		return emptyHash
	}

	root := c.root.mem.Load()
	if root != nil && root.hash != nil {
		// already computed working hash
		return root.hash
	}

	savedVersion := c.store.SavedVersion()
	stagedVersion := c.store.stagedVersion()
	c.commitCtx = &commitContext{
		version:      stagedVersion,
		savedVersion: savedVersion,
	}

	// compute hash and assign node IDs
	hash, err := commitTraverse(c.commitCtx, c.root, 0)
	if err != nil {
		panic(fmt.Sprintf("failed to compute working hash: %v", err))
	}
	return hash
}

func (c *CommitTree) Commit() storetypes.CommitID {
	commitId, err := c.commit()
	if err != nil {
		panic(fmt.Sprintf("failed to commit: %v", err))
	}
	return commitId
}

func (c *CommitTree) commit() (storetypes.CommitID, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	c.walQueue.Close()

	// compute hash and assign node IDs
	hash := c.workingHash()

	stagedVersion := c.store.stagedVersion()
	// wait for WAL write to complete
	err := <-c.walDone
	if err != nil {
		return storetypes.CommitID{}, err
	}

	_, err = c.store.WriteWALCommit(stagedVersion)
	if err != nil {
		return storetypes.CommitID{}, err
	}

	c.reinitWalProc()

	commitCtx := c.commitCtx
	if commitCtx == nil {
		// make sure we have a non-nil commit context
		commitCtx = &commitContext{}
	}
	err = c.store.SaveRoot(c.root, VersionStats{
		TotalLeaves:   commitCtx.leafNodeIdx,
		TotalBranches: commitCtx.branchNodeIdx,
		KVDataSize:    commitCtx.kvDataBytes,
	})
	if err != nil {
		return storetypes.CommitID{}, err
	}

	c.store.MarkOrphans(stagedVersion, c.pendingOrphans)
	c.pendingOrphans = nil

	// cache the committed tree as the latest version
	c.latest.Store(c.root)
	commitId := storetypes.CommitID{
		Version: int64(stagedVersion),
		Hash:    hash,
	}
	c.lastCommitId = commitId
	c.commitCtx = nil

	return commitId, nil
}

func (c *CommitTree) LastCommitID() storetypes.CommitID {
	return c.lastCommitId
}

func (c *CommitTree) SetPruning(pruningtypes.PruningOptions) {}

func (c *CommitTree) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningDefault)
}

func (c *CommitTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (c *CommitTree) CacheWrap() storetypes.CacheWrap {
	return NewCacheTree(c)
}

func (c *CommitTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO support tracing
	return c.CacheWrap()
}

func (c *CommitTree) Get(key []byte) []byte {
	if c.root == nil {
		return nil
	}

	root, err := c.root.Resolve()
	if err != nil {
		panic(err)
	}

	value, _, err := root.Get(key)
	if err != nil {
		panic(err)
	}

	return value
}

func (c *CommitTree) Has(key []byte) bool {
	return c.Get(key) != nil
}

func (c *CommitTree) Set(key, value []byte) {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	stagedVersion := c.store.stagedVersion()
	leafNode := &MemNode{
		height:  0,
		size:    1,
		version: stagedVersion,
		key:     key,
		value:   value,
	}

	// start writing this to the WAL asynchronously before we even mutate the tree
	c.walQueue.Send([]KVUpdate{{SetNode: leafNode}})

	ctx := &MutationContext{Version: stagedVersion}
	newRoot, _, err := setRecursive(c.root, leafNode, ctx)
	if err != nil {
		panic(err)
	}

	c.root = newRoot
	c.pendingOrphans = append(c.pendingOrphans, ctx.Orphans)
}

func (c *CommitTree) Delete(key []byte) {
	storetypes.AssertValidKey(key)

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// start writing this to the WAL asynchronously before we even mutate the tree
	c.walQueue.Send([]KVUpdate{{DeleteKey: key}})

	ctx := &MutationContext{Version: c.store.stagedVersion()}
	_, newRoot, _, err := removeRecursive(c.root, key, ctx)
	if err != nil {
		panic(err)
	}
	c.root = newRoot
	c.pendingOrphans = append(c.pendingOrphans, ctx.Orphans)
}

func (c *CommitTree) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, c.root, c.zeroCopy)
}

func (c *CommitTree) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, c.root, c.zeroCopy)
}

func (c *CommitTree) reinitWalProc() {
	walQueue := NewNonBlockingQueue[[]KVUpdate]()
	walDone := make(chan error, 1)
	c.walQueue = walQueue
	c.walDone = walDone

	go func() {
		for {
			batch := walQueue.Receive()
			if batch == nil {
				close(walDone)
				return
			}
			for _, updates := range batch {
				err := c.store.WriteWALUpdates(updates)
				if err != nil {
					walDone <- err
					return
				}
			}
		}
	}()
}

func (c *CommitTree) GetImmutable(version int64) (storetypes.KVStore, error) {
	var rootPtr *NodePointer
	if version == c.lastCommitId.Version {
		rootPtr = c.root
	} else {
		var err error
		rootPtr, err = c.store.ResolveRoot(uint32(version))
		if err != nil {
			return nil, err
		}
	}
	return NewImmutableTree(rootPtr), nil
}

func (c *CommitTree) ResolveRoot(version uint32) (*NodePointer, error) {
	if version == 0 {
		version = c.store.latestVersion.Load()
	}
	return c.store.ResolveRoot(version)
}

func (c *CommitTree) Version() uint32 {
	return c.store.latestVersion.Load()
}

func (c *CommitTree) Close() error {
	if c.walQueue != nil {
		c.walQueue.Close()
		// TODO do we need to wait for WAL done??
	}
	return c.store.Close()
}

type commitContext struct {
	version       uint32
	savedVersion  uint32
	branchNodeIdx uint32
	leafNodeIdx   uint32
	kvDataBytes   uint64
}

// commitTraverse performs a post-order traversal of the tree to compute hashes and assign node IDs.
// if it is run multiple times and the tree has been mutated before being committed, node IDs will be reassigned.
func commitTraverse(ctx *commitContext, np *NodePointer, depth uint8) (hash []byte, err error) {
	memNode := np.mem.Load()
	if memNode == nil {
		node, err := np.Resolve()
		if err != nil {
			return nil, err
		}
		return node.Hash(), nil
	}

	if memNode.version != ctx.version {
		return memNode.hash, nil
	}

	var leftHash, rightHash []byte
	var id NodeID
	if memNode.IsLeaf() {
		ctx.leafNodeIdx++
		id = NewNodeID(true, uint64(ctx.version), ctx.leafNodeIdx)
		ctx.kvDataBytes += uint64(len(memNode.key) + len(memNode.value))
	} else {
		// post-order traversal
		leftHash, err = commitTraverse(ctx, memNode.left, depth+1)
		if err != nil {
			return nil, err
		}
		rightHash, err = commitTraverse(ctx, memNode.right, depth+1)
		if err != nil {
			return nil, err
		}

		ctx.branchNodeIdx++
		ctx.kvDataBytes += uint64(len(memNode.key))
		id = NewNodeID(false, uint64(ctx.version), ctx.branchNodeIdx)
	}
	np.id = id
	memNode.nodeId = id

	if memNode.hash != nil {
		// hash previously computed node
		return memNode.hash, nil
	}

	return computeAndSetHash(memNode, leftHash, rightHash)
}

var (
	_ storetypes.CommitStore = &CommitTree{}
)
