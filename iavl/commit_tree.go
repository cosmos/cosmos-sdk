package iavl

import (
	"crypto/sha256"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"sync"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type CommitTree struct {
	writeMutex sync.Mutex

	opts Options

	treeStore *internal.TreeStore

	lastCommitId storetypes.CommitID

	walWriter *internal.KVDataWriter
}

func NewCommitTree(dir string, opts Options) (*CommitTree, error) {
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("creating tree directory: %w", err)
	}

	treeStore := &internal.TreeStore{}
	var walWriter *internal.KVDataWriter
	if dir != "" {
		walFile, err := os.OpenFile(filepath.Join(dir, "wal.dat"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("opening WAL file: %w", err)
		}
		walWriter = internal.NewKVDataWriter(walFile)
		err = walWriter.WriteStartWAL(uint64(treeStore.StagedVersion()))
		if err != nil {
			return nil, fmt.Errorf("writing WAL start: %w", err)
		}
	}
	return &CommitTree{
		opts:      opts,
		treeStore: treeStore,
		walWriter: walWriter,
	}, nil
}

//	func (c *CommitTree) WorkingHash() ([]byte, error) {
//		c.writeMutex.Lock()
//		defer c.writeMutex.Unlock()
//
//		return c.workingHash()
//	}
var emptyHash = sha256.New().Sum(nil)

func workingHash(rootPtr *internal.NodePointer) ([]byte, error) {
	// IMPORTANT: this function assumes the write lock is held

	// if we have no root, return empty hash
	if rootPtr == nil {
		return emptyHash, nil
	}

	root, pin, err := rootPtr.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, fmt.Errorf("resolving root node: %w", err)
	}

	hash, err := root.ComputeHash()
	if err != nil {
		return nil, fmt.Errorf("computing root hash: %w", err)
	}

	return hash.SafeCopy(), nil
}

func (c *CommitTree) Commit(updates iter.Seq[Update]) (storetypes.CommitID, error) {
	// TODO maybe support writing in batches, but for now let's assume a single batch
	// because that's actually what happens in the SDK due to cache wrapping and needing all keys in sorted
	// order before updating the tree
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	stagedVersion := c.treeStore.StagedVersion()

	// TODO pre-allocate a decent sized slice that we reuse and reset across commits
	// TODO we can also add a function param for the number of updates to pre-allocate
	var nodeUpdates []internal.KVUpdate
	for update := range updates {
		if update.Delete {
			nodeUpdates = append(nodeUpdates, internal.KVUpdate{DeleteKey: update.Key})
		} else {
			nodeUpdates = append(nodeUpdates, internal.KVUpdate{SetNode: internal.NewLeafNode(update.Key, update.Value, stagedVersion)})
		}
	}

	// TODO background start writing nodeUpdates to WAL immediately
	var walDone chan error
	if c.walWriter != nil {
		walDone = make(chan error, 1)
		go func() {
			defer close(walDone)
			err := c.walWriter.WriteWALUpdates(nodeUpdates)
			if err != nil {
				walDone <- err
			}

			err = c.walWriter.WriteWALCommit(uint64(stagedVersion))
			if err != nil {
				walDone <- err
			}

			if c.opts.Fsync {
				err = c.walWriter.Sync()
				if err != nil {
					walDone <- err
				}
			}
		}()
	}

	// start computing hashes for leaf nodes in parallel
	hashErr := make(chan error, 1)
	go func() {
		defer close(hashErr)
		for _, nu := range nodeUpdates {
			if setNode := nu.SetNode; setNode != nil {
				_, err := setNode.ComputeHash()
				if err != nil {
					hashErr <- err
					return
				}
			}
		}
	}()

	// process all the nodeUpdates against the current root to produce a new root
	root := c.treeStore.Latest()
	mutationCtx := internal.NewMutationContext(stagedVersion)
	for _, nu := range nodeUpdates {
		var err error
		if setNode := nu.SetNode; setNode != nil {
			root, _, err = internal.SetRecursive(root, setNode, mutationCtx)
			if err != nil {
				return storetypes.CommitID{}, err
			}
		} else {
			_, root, _, err = internal.RemoveRecursive(root, nu.DeleteKey, mutationCtx)
			if err != nil {
				return storetypes.CommitID{}, err
			}
		}
	}

	// wait for the leaf node hash queue to finish
	if err := <-hashErr; err != nil {
		return storetypes.CommitID{}, err
	}

	// compute the root hash
	hash, err := workingHash(root)
	if err != nil {
		return storetypes.CommitID{}, err
	}

	// wait for the WAL write to finish
	// background queue root to be stored, after that is done eviction can start

	//if c.writeWal {
	//	c.walQueue.Close()
	//}
	//
	// compute hash and assign node IDs
	//hash, err := c.workingHash()
	//if err != nil {
	//	return storetypes.CommitID{}, err
	//}

	//stagedVersion := c.store.stagedVersion
	//if c.writeWal {
	//	// wait for WAL write to complete
	//	err := <-c.walDone
	//	if err != nil {
	//		return storetypes.CommitID{}, err
	//	}
	//
	//	err = c.store.WriteWALCommit(stagedVersion)
	//	if err != nil {
	//		return storetypes.CommitID{}, err
	//	}
	//
	//	c.reinitWalProc()
	//}
	//
	//commitCtx := c.commitCtx
	//if commitCtx == nil {
	//	// make sure we have a non-nil commit context
	//	commitCtx = &commitContext{}
	//}
	//err := c.store.SaveRoot(c.root, commitCtx.leafNodeIdx, commitCtx.branchNodeIdx)
	//if err != nil {
	//	return storetypes.CommitID{}, err
	//}
	//
	//c.store.MarkOrphans(stagedVersion, c.pendingOrphans)
	//c.pendingOrphans = nil
	//
	//// start eviction if needed
	//c.startEvict(c.store.SavedVersion())
	//
	//// cache the committed tree as the latest version
	err = c.treeStore.SaveRoot(root)
	if err != nil {
		return storetypes.CommitID{}, err
	}

	commitId := storetypes.CommitID{
		Version: int64(stagedVersion),
		Hash:    hash,
	}
	c.lastCommitId = commitId
	////c.commitCtx = nil
	//

	if walDone != nil {
		err := <-walDone
		if err != nil {
			return storetypes.CommitID{}, err
		}
	}

	return commitId, nil
}

func (c *CommitTree) Latest() TreeReader {
	return &treeReader{root: c.treeStore.Latest()}
}

type TreeReader interface {
	Get(key []byte) ([]byte, error)
	Size() int64
}

type treeReader struct {
	root *internal.NodePointer
}

func (t treeReader) Get(key []byte) ([]byte, error) {
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return nil, err
	}
	value, _, err := root.Get(key)
	if err != nil {
		return nil, err
	}
	return value.SafeCopy(), nil
}

func (t treeReader) Size() int64 {
	if t.root == nil {
		return 0
	}
	root, pin, err := t.root.Resolve()
	defer pin.Unpin()
	if err != nil {
		return 0
	}
	return root.Size()
}

//
//func (c *CommitTree) LastCommitID() storetypes.CommitID {
//	return c.lastCommitId
//}
//
//func (c *CommitTree) SetPruning(pruningtypes.PruningOptions) {}
//
//func (c *CommitTree) GetPruning() pruningtypes.PruningOptions {
//	return pruningtypes.NewPruningOptions(pruningtypes.PruningDefault)
//}
//
//func (c *CommitTree) GetStoreType() storetypes.StoreType {
//	return storetypes.StoreTypeIAVL
//}
//
//func (c *CommitTree) CacheWrap() storetypes.CacheWrap {
//	return NewCacheTree(c)
//}
//
//func (c *CommitTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
//	// TODO support tracing
//	return c.CacheWrap()
//}
//
//func (c *CommitTree) Get(key []byte) []byte {
//	if c.root == nil {
//		return nil
//	}
//
//	root, err := c.root.Resolve()
//	if err != nil {
//		panic(err)
//	}
//
//	value, _, err := root.Get(key)
//	if err != nil {
//		panic(err)
//	}
//
//	return value
//}
//
//func (c *CommitTree) Has(key []byte) bool {
//	return c.Get(key) != nil
//}
//
//func (c *CommitTree) Set(key, value []byte) {
//	storetypes.AssertValidKey(key)
//	storetypes.AssertValidValue(value)
//
//	c.writeMutex.Lock()
//	defer c.writeMutex.Unlock()
//
//	stagedVersion := c.store.stagedVersion
//	leafNode := &MemNode{
//		height:  0,
//		size:    1,
//		version: stagedVersion,
//		key:     key,
//		value:   value,
//	}
//
//	if c.writeWal {
//		// start writing this to the WAL asynchronously before we even mutate the tree
//		c.walQueue.Send([]KVUpdate{{SetNode: leafNode}})
//	}
//
//	ctx := &MutationContext{Version: stagedVersion}
//	newRoot, _, err := setRecursive(c.root, leafNode, ctx)
//	if err != nil {
//		panic(err)
//	}
//
//	c.root = newRoot
//	c.pendingOrphans = append(c.pendingOrphans, ctx.Orphans)
//}
//
//func (c *CommitTree) Delete(key []byte) {
//	storetypes.AssertValidKey(key)
//
//	c.writeMutex.Lock()
//	defer c.writeMutex.Unlock()
//
//	if c.writeWal {
//		// start writing this to the WAL asynchronously before we even mutate the tree
//		c.walQueue.Send([]KVUpdate{{DeleteKey: key}})
//	}
//
//	ctx := &MutationContext{Version: c.store.stagedVersion}
//	_, newRoot, _, err := removeRecursive(c.root, key, ctx)
//	if err != nil {
//		panic(err)
//	}
//	c.root = newRoot
//	c.pendingOrphans = append(c.pendingOrphans, ctx.Orphans)
//}
//
//func (c *CommitTree) Iterator(start, end []byte) storetypes.Iterator {
//	return NewIterator(start, end, true, c.root, c.zeroCopy)
//}
//
//func (c *CommitTree) ReverseIterator(start, end []byte) storetypes.Iterator {
//	return NewIterator(start, end, false, c.root, c.zeroCopy)
//}
//
//func (c *CommitTree) reinitWalProc() {
//	if !c.writeWal {
//		return
//	}
//
//	walQueue := NewNonBlockingQueue[[]KVUpdate]()
//	walDone := make(chan error, 1)
//	c.walQueue = walQueue
//	c.walDone = walDone
//
//	go func() {
//		for {
//			batch := walQueue.Receive()
//			if batch == nil {
//				close(walDone)
//				return
//			}
//			for _, updates := range batch {
//				err := c.store.WriteWALUpdates(updates)
//				if err != nil {
//					walDone <- err
//					return
//				}
//			}
//		}
//	}()
//}
//
//func (c *CommitTree) startEvict(evictVersion uint32) {
//	if c.evictorRunning.Load() {
//		// eviction in progress
//		return
//	}
//
//	if evictVersion <= c.lastEvictVersion {
//		// no new version to evict
//		return
//	}
//
//	latest := c.latest.Load()
//	if latest == nil {
//		// nothing to evict
//		return
//	}
//
//	c.logger.Debug("start eviction", "version", evictVersion, "depth", c.evictionDepth)
//	c.evictorRunning.Store(true)
//	go func() {
//		evictedCount := evictTraverse(latest, 0, c.evictionDepth, evictVersion)
//		c.logger.Debug("eviction completed", "version", evictVersion, "lastEvict", c.lastEvictVersion, "evictedNodes", evictedCount)
//		c.lastEvictVersion = evictVersion
//		c.evictorRunning.Store(false)
//	}()
//}
//
//func (c *CommitTree) GetImmutable(version int64) (storetypes.KVStore, error) {
//	var rootPtr *NodePointer
//	if version == c.lastCommitId.Version {
//		rootPtr = c.root
//	} else {
//		var err error
//		rootPtr, err = c.store.ResolveRoot(uint32(version))
//		if err != nil {
//			return nil, err
//		}
//	}
//	return NewImmutableTree(rootPtr), nil
//}
//
//func (c *CommitTree) ResolveRoot(version uint32) (*NodePointer, error) {
//	if version == 0 {
//		version = c.store.stagedVersion - 1
//	}
//	return c.store.ResolveRoot(version)
//}
//
//func (c *CommitTree) Version() uint32 {
//	return c.store.stagedVersion - 1
//}
//
//func (c *CommitTree) Close() error {
//	if c.walQueue != nil {
//		c.walQueue.Close()
//		// TODO do we need to wait for WAL done??
//	}
//	return c.store.Close()
//}
//
//type commitContext struct {
//	version       uint32
//	savedVersion  uint32
//	branchNodeIdx uint32
//	leafNodeIdx   uint32
//}
//
//// commitTraverse performs a post-order traversal of the tree to compute hashes and assign node IDs.
//// if it is run multiple times and the tree has been mutated before being committed, node IDs will be reassigned.
//func commitTraverse(ctx *commitContext, np *NodePointer, depth uint8) (hash []byte, err error) {
//	memNode := np.mem.Load()
//	if memNode == nil {
//		node, err := np.Resolve()
//		if err != nil {
//			return nil, err
//		}
//		return node.Hash(), nil
//	}
//
//	if memNode.version != ctx.version {
//		return memNode.hash, nil
//	}
//
//	var leftHash, rightHash []byte
//	var id NodeID
//	if memNode.IsLeaf() {
//		ctx.leafNodeIdx++
//		id = NewNodeID(true, ctx.version, ctx.leafNodeIdx)
//	} else {
//		// post-order traversal
//		leftHash, err = commitTraverse(ctx, memNode.left, depth+1)
//		if err != nil {
//			return nil, err
//		}
//		rightHash, err = commitTraverse(ctx, memNode.right, depth+1)
//		if err != nil {
//			return nil, err
//		}
//
//		ctx.branchNodeIdx++
//		id = NewNodeID(false, ctx.version, ctx.branchNodeIdx)
//	}
//	np.id = id
//	memNode.nodeId = id
//
//	if memNode.hash != nil {
//		// hash previously computed node
//		return memNode.hash, nil
//	}
//
//	return computeAndSetHash(memNode, leftHash, rightHash)
//}
//
//func evictTraverse(np *NodePointer, depth, evictionDepth uint8, evictVersion uint32) (count int) {
//	// TODO check height, and don't traverse if tree is too short
//
//	memNode := np.mem.Load()
//	if memNode == nil {
//		return 0
//	}
//
//	// Evict nodes at or below the eviction depth
//	if memNode.version <= evictVersion && depth >= evictionDepth {
//		np.mem.Store(nil)
//		count = 1
//	}
//
//	if memNode.IsLeaf() {
//		return count
//	}
//
//	// Continue traversing to find nodes to evict
//	count += evictTraverse(memNode.left, depth+1, evictionDepth, evictVersion)
//	count += evictTraverse(memNode.right, depth+1, evictionDepth, evictVersion)
//	return count
//}
//
//var (
//	_ storetypes.CommitStore = &CommitTree{}
//)
