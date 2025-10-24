package iavlx

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
)

type CommitTree struct {
	latest     atomic.Pointer[NodePointer]
	root       *NodePointer
	version    uint32
	writeMutex sync.Mutex
	store      *TreeStore
	zeroCopy   bool

	evictionDepth    uint8
	evictorRunning   bool
	lastEvictVersion uint32

	writeWal bool
	walChan  chan<- []KVUpdate
	walDone  <-chan error

	pendingOrphans [][]NodeID

	logger log.Logger

	lastCommitId storetypes.CommitID
	commitCtx    *commitContext
}

func (c *CommitTree) getRoot() *NodePointer {
	return c.root
}

func (c *CommitTree) applyChangesToParent(origRoot, newRoot *NodePointer, updateBatch KVUpdateBatch) error {
	// TODO check channel errors
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if updateBatch.Version != c.stagedVersion() {
		return fmt.Errorf("tree version %d does not match staged version %d", updateBatch.Version, c.stagedVersion())
	}
	if origRoot != c.root {
		// TODO find a way to apply the changes incrementally when roots don't match
		return fmt.Errorf("tree original root does not match current root")
	}
	c.root = newRoot
	c.pendingOrphans = append(c.pendingOrphans, updateBatch.Orphans...)

	if c.writeWal {
		c.walChan <- updateBatch.Updates
	}

	return nil
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
	stagedVersion := c.stagedVersion()
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

	if c.writeWal {
		close(c.walChan)
	}

	// compute hash and assign node IDs
	hash := c.workingHash()

	stagedVersion := c.stagedVersion()
	if c.writeWal {
		// wait for WAL write to complete
		err := <-c.walDone
		if err != nil {
			return storetypes.CommitID{}, err
		}

		err = c.store.WriteWALCommit(stagedVersion)
		if err != nil {
			return storetypes.CommitID{}, err
		}

		c.reinitWalProc()
	}

	commitCtx := c.commitCtx
	if commitCtx == nil {
		// make sure we have a non-nil commit context
		commitCtx = &commitContext{}
	}
	err := c.store.SaveRoot(stagedVersion, c.root, commitCtx.leafNodeIdx, commitCtx.branchNodeIdx)
	if err != nil {
		return storetypes.CommitID{}, err
	}

	c.store.MarkOrphans(stagedVersion, c.pendingOrphans)
	c.pendingOrphans = nil

	// start eviction if needed
	c.startEvict(c.store.SavedVersion())

	// cache the committed tree as the latest version
	c.latest.Store(c.root)
	c.version++
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
	return NewTree(c, c.stagedVersion(), c.zeroCopy)
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
	tree := c.CacheWrap().(*Tree)
	tree.Set(key, value)
	tree.Write()
}

func (c *CommitTree) Delete(key []byte) {
	tree := c.CacheWrap().(*Tree)
	tree.Delete(key)
	tree.Write()
}

func (c *CommitTree) Iterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, true, c.root, c.zeroCopy)
}

func (c *CommitTree) ReverseIterator(start, end []byte) storetypes.Iterator {
	return NewIterator(start, end, false, c.root, c.zeroCopy)
}

func NewCommitTree(dir string, opts Options, logger log.Logger) (*CommitTree, error) {
	ts, err := NewTreeStore(dir, opts, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tree store: %w", err)
	}

	tree := &CommitTree{
		root:          nil,
		zeroCopy:      opts.ZeroCopy,
		version:       0,
		logger:        logger,
		store:         ts,
		evictionDepth: opts.EvictDepth,
		writeWal:      opts.WriteWAL,
	}
	tree.reinitWalProc()

	return tree, nil
}

func (c *CommitTree) stagedVersion() uint32 {
	return c.version + 1
}

func (c *CommitTree) reinitWalProc() {
	if !c.writeWal {
		return
	}

	walChan := make(chan []KVUpdate, 2048)
	walDone := make(chan error, 1)
	c.walChan = walChan
	c.walDone = walDone

	go func() {
		defer close(walDone)
		for updates := range walChan {
			err := c.store.WriteWALUpdates(updates)
			if err != nil {
				walDone <- err
				return
			}
		}
	}()
}

func (c *CommitTree) startEvict(evictVersion uint32) {
	if c.evictorRunning {
		// eviction in progress
		return
	}

	if evictVersion <= c.lastEvictVersion {
		// no new version to evict
		return
	}

	latest := c.latest.Load()
	if latest == nil {
		// nothing to evict
		return
	}

	c.logger.Debug("start eviction", "version", evictVersion, "depth", c.evictionDepth)
	c.evictorRunning = true
	go func() {
		evictedCount := evictTraverse(latest, 0, c.evictionDepth, evictVersion)
		c.logger.Debug("eviction completed", "version", evictVersion, "lastEvict", c.lastEvictVersion, "evictedNodes", evictedCount)
		c.lastEvictVersion = evictVersion
		c.evictorRunning = false
	}()
}

func (c *CommitTree) GetImmutable(version int64) (storetypes.CacheWrap, error) {
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
	return &Tree{
		root: rootPtr,
	}, nil
}

func (c *CommitTree) Close() error {
	if c.walChan != nil {
		close(c.walChan)
	}
	// close(c.commitChan)
	// return <-c.commitDone
	return c.store.Close()
}

type commitContext struct {
	version       uint32
	savedVersion  uint32
	branchNodeIdx uint32
	leafNodeIdx   uint32
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
	if memNode.IsLeaf() {
		ctx.leafNodeIdx++
		np.id = NewNodeID(true, uint64(ctx.version), ctx.leafNodeIdx)
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
		np.id = NewNodeID(false, uint64(ctx.version), ctx.branchNodeIdx)

	}

	if memNode.hash != nil {
		// hash previously computed node
		return memNode.hash, nil
	}

	return computeAndSetHash(memNode, leftHash, rightHash)
}

func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictVersion uint32) (count int) {
	memNode := np.mem.Load()
	if memNode == nil {
		return 0
	}

	// Evict nodes at or below the eviction depth
	if memNode.version <= evictVersion && depth >= evictionDepth {
		np.mem.Store(nil)
		count = 1
	}

	if memNode.IsLeaf() {
		return count
	}

	// Continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictVersion)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictVersion)
	return count
}

var (
	_ storetypes.CommitStore = &CommitTree{}
	_ parentTree             = &CommitTree{}
)
