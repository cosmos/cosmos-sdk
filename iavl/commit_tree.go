package iavlx

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
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

	logger *slog.Logger
}

func NewCommitTree(dir string, opts Options, logger *slog.Logger) (*CommitTree, error) {
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

func (c *CommitTree) Branch() *Tree {
	return NewTree(c.root, NewKVUpdateBatch(c.stagedVersion()), c.zeroCopy)
}

func (c *CommitTree) Apply(tree *Tree) error {
	// TODO check channel errors
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if tree.updateBatch.Version != c.stagedVersion() {
		return fmt.Errorf("tree version %d does not match staged version %d", tree.updateBatch.Version, c.stagedVersion())
	}
	if tree.origRoot != c.root {
		// TODO find a way to apply the changes incrementally when roots don't match
		return fmt.Errorf("tree original root does not match current root")
	}
	c.root = tree.root
	batch := tree.updateBatch
	c.pendingOrphans = append(c.pendingOrphans, batch.Orphans...)

	if c.writeWal {
		c.walChan <- batch.Updates
	}

	// TODO prevent further writes to the branch tree

	return nil
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

func (c *CommitTree) Commit() ([]byte, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if c.writeWal {
		close(c.walChan)
	}

	var hash []byte
	savedVersion := c.store.SavedVersion()
	stagedVersion := c.stagedVersion()
	commitCtx := &commitContext{
		version:      stagedVersion,
		savedVersion: savedVersion,
	}
	if c.root == nil {
		hash = emptyHash
	} else {
		// compute hash and assign node IDs
		var err error
		hash, err = commitTraverse(commitCtx, c.root, 0)
		if err != nil {
			return nil, err
		}
	}

	if c.writeWal {
		// wait for WAL write to complete
		err := <-c.walDone
		if err != nil {
			return nil, err
		}

		err = c.store.WriteWALCommit(stagedVersion)
		if err != nil {
			return nil, err
		}

		c.reinitWalProc()
	}

	err := c.store.SaveRoot(stagedVersion, c.root, commitCtx.leafNodeIdx, commitCtx.branchNodeIdx)
	if err != nil {
		return nil, err
	}

	c.store.MarkOrphans(stagedVersion, c.pendingOrphans)
	c.pendingOrphans = nil

	// start eviction if needed
	c.startEvict(savedVersion)

	// cache the committed tree as the latest version
	c.latest.Store(c.root)
	c.version++

	return hash, nil
}

func (c *CommitTree) Close() error {
	if c.walChan != nil {
		close(c.walChan)
	}
	//close(c.commitChan)
	//return <-c.commitDone
	return c.store.Close()
}

type commitContext struct {
	version       uint32
	savedVersion  uint32
	branchNodeIdx uint32
	leafNodeIdx   uint32
}

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
		// not sure when we would encounter this but if the hash is already computed, just return it
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
		return
	}

	// Continue traversing to find nodes to evict
	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictVersion)
	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictVersion)
	return
}
