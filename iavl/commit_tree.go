package iavl

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	io "io"
	"iter"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal/cachekv"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type CommitTree struct {
	commitMutex sync.Mutex

	opts Options

	treeStore *internal.TreeStore

	lastCommitId storetypes.CommitID

	lastNodeIDsAssigned chan struct{}
}

func NewCommitTree(dir string, opts Options) (*CommitTree, error) {
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("creating tree directory: %w", err)
	}

	changesetRolloverSize := opts.ChangesetRolloverSize
	if changesetRolloverSize == 0 {
		changesetRolloverSize = 2 * 1024 * 1024 * 1024 // 2GB default
	}
	evictDepth := opts.EvictDepth
	if evictDepth == 0 {
		evictDepth = 16 // default evict depth 2^16 = 65536 leaf nodes
	}
	treeStore, err := internal.NewTreeStore(dir, internal.TreeStoreOptions{
		ChangesetRolloverSize: changesetRolloverSize,
		EvictDepth:            evictDepth,
		CheckpointInterval:    opts.CheckpointInterval,
	})

	if err != nil {
		return nil, fmt.Errorf("creating tree store: %w", err)
	}
	return &CommitTree{
		opts:      opts,
		treeStore: treeStore,
	}, nil
}

var emptyHash = sha256.New().Sum(nil)

func rootHash(ctx context.Context, rootPtr *internal.NodePointer) ([]byte, error) {
	// IMPORTANT: this function assumes the write lock is held

	// if we have no root, return empty hash
	if rootPtr == nil {
		return emptyHash, nil
	}

	root := rootPtr.Mem.Load()
	if root == nil {
		rootNode, pin, err := rootPtr.Resolve()
		defer pin.Unpin()
		if err != nil {
			return nil, fmt.Errorf("resolving root node: %w", err)
		}
		return rootNode.Hash().SafeCopy(), nil
	}

	scheduler := internal.NewAsyncHashScheduler(ctx, int32(runtime.NumCPU()))
	hash, err := root.ComputeHash(scheduler)
	if err != nil {
		return nil, fmt.Errorf("computing root hash: %w", err)
	}

	return hash, nil
}

type committer struct {
	*CommitTree
	cancel             context.CancelFunc
	finalizeOrRollback chan struct{}
	hashReady          chan struct{}
	done               chan struct{}
	err                atomic.Value
	workingHash        storetypes.CommitID
}

func (c *CommitTree) StartCommit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) storetypes.CommitFinalizer {
	cancelCtx, cancel := context.WithCancel(ctx)
	committer := &committer{
		CommitTree:         c,
		cancel:             cancel,
		finalizeOrRollback: make(chan struct{}),
		hashReady:          make(chan struct{}),
		done:               make(chan struct{}),
	}
	go func() {
		err := committer.commit(cancelCtx, updates, updateCount)
		committer.err.Store(err)
		close(committer.done)
	}()
	return committer
}

func (c *committer) commit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) error {
	c.commitMutex.Lock()
	defer c.commitMutex.Unlock()

	stagedVersion := c.treeStore.StagedVersion()
	ctx, span := tracer.Start(ctx, "CommitTree.Commit",
		trace.WithAttributes(
			attribute.Int64("version", int64(stagedVersion)),
			attribute.Int("updateCount", updateCount),
		),
	)
	defer span.End()

	prepareRes, err := c.prepareCommit(ctx, updates, updateCount)
	if err != nil {
		rbErr := c.treeStore.RollbackWAL()
		if !errors.Is(rbErr, context.Canceled) {
			return fmt.Errorf("commit failed: %w; rollback failed: %v", err, rbErr)
		}
		return fmt.Errorf("successful rollback: %w; rollback cause %v", rbErr, err)
	}

	// save the new root after the WAL is fully written so that all offsets are populated correctly
	c.lastNodeIDsAssigned = prepareRes.nodeIdsAssigned
	err = c.treeStore.SaveRoot(
		prepareRes.root,
		prepareRes.mutationCtx,
		prepareRes.nodeIdsAssigned,
	)
	if err != nil {
		return err
	}

	c.lastCommitId = c.workingHash
	return nil
}

type prepareCommitResult struct {
	root            *internal.NodePointer
	mutationCtx     *internal.MutationContext
	nodeIdsAssigned chan struct{}
}

func (c *committer) prepareCommit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) (*prepareCommitResult, error) {
	ctx, span := tracer.Start(ctx, "PrepareCommit")
	defer span.End()

	stagedVersion := c.treeStore.StagedVersion()
	mutationCtx := internal.NewMutationContext(stagedVersion, stagedVersion)

	// TODO pre-allocate a decent sized slice that we reuse and reset across commits
	nodeUpdates := make([]internal.KVUpdate, 0, updateCount)
	for update := range updates {
		if update.Delete {
			nodeUpdates = append(nodeUpdates, internal.KVUpdate{DeleteKey: update.Key})
		} else {
			nodeUpdates = append(
				nodeUpdates,
				internal.KVUpdate{SetNode: mutationCtx.NewLeafNode(update.Key, update.Value)},
			)
		}
	}

	// background start writing nodeUpdates to WAL immediately
	var walDone chan error
	walDone = make(chan error, 1)
	go func() {
		_, walSpan := tracer.Start(ctx, "WALWrite")
		defer walSpan.End()
		defer close(walDone)
		walDone <- c.treeStore.WriteWALUpdates(ctx, nodeUpdates, !c.opts.DisableWALFsync)
	}()

	// start computing hashes for leaf nodes in parallel
	// this is a rather naive algorithm that assumes leaf nodes updates are evenly distributed
	// if we see consistent latency here we can tune the algorithm
	leafHashErr := make(chan error, 1)
	go func() {
		defer close(leafHashErr)
		_, hashLeavesSpan := tracer.Start(ctx, "LeafHashCompute")
		defer hashLeavesSpan.End()

		n := len(nodeUpdates)
		if n == 0 {
			return
		}

		// we choose a minimum bucket size to avoid too many goroutines being created for small buckets
		const minBucketSize = 64
		numCPUs := runtime.NumCPU()
		numWorkers := min(numCPUs, max(1, n/minBucketSize))
		bucketSize := (n + numWorkers - 1) / numWorkers

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			start := i * bucketSize
			end := min(start+bucketSize, n)
			wg.Add(1)
			go func(updates []internal.KVUpdate) {
				defer wg.Done()
				if ctx.Err() != nil {
					return
				}
				for _, nu := range updates {
					if setNode := nu.SetNode; setNode != nil {
						if _, err := setNode.ComputeHash(internal.SyncHashScheduler{}); err != nil {
							select {
							case leafHashErr <- err:
							default:
							}
							return
						}
					}
				}
			}(nodeUpdates[start:end])
		}
		wg.Wait()
	}()

	// wait for any node IDs assignment from any previous commit to finish
	if nodeIDsAssigned := c.lastNodeIDsAssigned; nodeIDsAssigned != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-nodeIDsAssigned:
		}
		c.lastNodeIDsAssigned = nil
	}

	// process all the nodeUpdates against the current root to produce a new root
	root := c.treeStore.Latest()
	_, nodeUpdatesSpan := tracer.Start(ctx, "ApplyNodeUpdates")
	for _, nu := range nodeUpdates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var err error
		if setNode := nu.SetNode; setNode != nil {
			root, _, err = internal.SetRecursive(root, setNode, mutationCtx)
			if err != nil {
				return nil, err
			}
		} else {
			_, root, _, err = internal.RemoveRecursive(root, nu.DeleteKey, mutationCtx)
			if err != nil {
				return nil, err
			}
		}
	}
	nodeUpdatesSpan.End()

	var nodeIDsAssigned chan struct{}
	if c.treeStore.ShouldCheckpoint() {
		// if we need to checkpoint, we must start a goroutine to assign node IDs in the background
		nodeIDsAssigned = make(chan struct{})
		go func() {
			defer close(nodeIDsAssigned)
			_, span := tracer.Start(ctx, "AssignNodeIDs")
			defer span.End()

			internal.AssignNodeIDs(root, c.treeStore.StagedCheckpoint())
		}()
	}

	startWaitForLeafHashes := time.Now()

	// wait for the leaf node hash queue to finish
	select {
	case err := <-leafHashErr:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	span.AddEvent("leaf hash compute returned")

	leafHashLatency.Record(ctx, time.Since(startWaitForLeafHashes).Milliseconds())

	ctx, rootHashSpan := tracer.Start(ctx, "ComputeRootHash")
	// compute the root hash
	hash, err := rootHash(ctx, root)
	if err != nil {
		return nil, err
	}
	// notify that the root hash is ready
	c.workingHash = storetypes.CommitID{
		Hash:    hash,
		Version: int64(stagedVersion),
	}
	close(c.hashReady)
	rootHashSpan.End()

	// wait for the WAL write to finish
	startWaitForWAL := time.Now()

	err = <-walDone
	if err != nil {
		return nil, err
	}

	walWriteLatency.Record(ctx, time.Since(startWaitForWAL).Milliseconds())
	span.AddEvent("WAL write returned")

	<-c.finalizeOrRollback

	return &prepareCommitResult{
		root:            root,
		mutationCtx:     mutationCtx,
		nodeIdsAssigned: nodeIDsAssigned,
	}, ctx.Err()
}

func (c *committer) WorkingHash() (storetypes.CommitID, error) {
	select {
	case <-c.hashReady:
	case <-c.done:
	}
	err := c.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	return c.workingHash, nil
}

func (c *committer) Rollback() error {
	c.cancel()
	close(c.finalizeOrRollback)
	<-c.done
	err := c.err.Load()
	if err != nil {
		// we expect an error if we rolled back successfully
		return err.(error)
	}
	return nil
}

func (c *committer) SignalFinalize() error {
	close(c.finalizeOrRollback)
	return nil
}

func (c *committer) WaitFinalize() (storetypes.CommitID, error) {
	<-c.done
	err := c.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	// only return the lastCommitId after successful commit
	return c.lastCommitId, nil
}

func (c *CommitTree) Latest() internal.TreeReader {
	return internal.NewTreeReader(c.treeStore.Latest())
}

func (c *CommitTree) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(internal.KVStoreWrapper{TreeReader: c.Latest()})
}

func (c *CommitTree) CacheWrapWithTrace(io.Writer, storetypes.TraceContext) storetypes.CacheWrap {
	return c.CacheWrap()
}

func (c *CommitTree) Close() error {
	return c.treeStore.Close()
}

var _ storetypes.CacheWrapper = (*CommitTree)(nil)
