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
	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type CommitTree struct {
	commitMutex sync.Mutex

	opts Options

	treeStore *internal.TreeStore

	lastCommitId storetypes.CommitID
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
	var rootCacheSize uint64 = 5 // default to caching 5 roots
	if opts.RootCacheSize > 0 {
		rootCacheSize = uint64(opts.RootCacheSize)
	} else if opts.RootCacheSize < 0 {
		rootCacheSize = 0
	}
	var rootCacheExpiry = 5 * time.Second // default to 5 seconds
	if opts.RootCacheExpiry > 0 {
		rootCacheExpiry = time.Duration(opts.RootCacheExpiry) * time.Millisecond
	}
	treeStore, err := internal.NewTreeStore(dir, internal.TreeStoreOptions{
		ChangesetRolloverSize: changesetRolloverSize,
		EvictDepth:            evictDepth,
		CheckpointInterval:    opts.CheckpointInterval,
		RootCacheSize:         rootCacheSize,
		RootCacheExpiry:       rootCacheExpiry,
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

type commitTreeFinalizer struct {
	*CommitTree
	cancel             context.CancelFunc
	finalizeOnce       sync.Once
	finalizeOrRollback chan struct{}
	hashReady          chan struct{}
	done               chan struct{}
	err                atomic.Value
	workingHash        storetypes.CommitID
}

func (c *CommitTree) StartCommit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) storetypes.CommitFinalizer {
	cancelCtx, cancel := context.WithCancel(ctx)
	committer := &commitTreeFinalizer{
		CommitTree:         c,
		cancel:             cancel,
		finalizeOrRollback: make(chan struct{}),
		hashReady:          make(chan struct{}),
		done:               make(chan struct{}),
	}
	go func() {
		err := committer.commit(cancelCtx, updates, updateCount)
		if err != nil {
			committer.err.Store(err)
		}
		close(committer.done)
	}()
	return committer
}

var rolledbackErr = errors.New("commit rolled back")

func (c *commitTreeFinalizer) commit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) error {
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
		if rbErr != nil {
			return fmt.Errorf("commit failed: %w; rollback failed: %v", err, rbErr)
		}
		return fmt.Errorf("%w; root cause %v", rolledbackErr, err)
	}

	// assign node IDs if needed
	root := prepareRes.root
	// save the new root after the WAL is fully written so that all offsets are populated correctly
	err = c.treeStore.SaveRoot(
		ctx,
		root,
		prepareRes.mutationCtx,
	)
	if err != nil {
		return err
	}

	c.lastCommitId = c.workingHash
	return nil
}

type prepareCommitResult struct {
	root        *internal.NodePointer
	mutationCtx *internal.MutationContext
}

func (c *commitTreeFinalizer) prepareCommit(ctx context.Context, updates iter.Seq[KVUpdate], updateCount int) (*prepareCommitResult, error) {
	ctx, span := tracer.Start(ctx, "PrepareCommit")
	defer span.End()

	stagedVersion := c.treeStore.StagedVersion()
	mutationCtx := internal.NewMutationContext(stagedVersion, stagedVersion)

	// TODO pre-allocate a decent sized slice that we reuse and reset across commits
	nodeUpdates := make([]internal.KVUpdate, 0, updateCount)
	if updates != nil { // updates can be nil for empty commits
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

	// process all the nodeUpdates against the current root to produce a new root
	root, err := c.treeStore.GetRootForUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("tree store not ready for update: %w", err)
	}

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
		root:        root,
		mutationCtx: mutationCtx,
	}, ctx.Err()
}

func (c *commitTreeFinalizer) WaitForHash() (storetypes.CommitID, error) {
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

func (c *commitTreeFinalizer) PrepareFinalize() (storetypes.CommitID, error) {
	if err := c.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}
	return c.WaitForHash()
}

func (c *commitTreeFinalizer) Rollback() error {
	c.cancel()
	c.finalizeOnce.Do(func() {
		close(c.finalizeOrRollback)
	})
	<-c.done
	err := c.err.Load()
	if err == nil {
		// we expect an error if we rolled back successfully
		return fmt.Errorf("rollback failed, commit succeeded")
	}
	if !errors.Is(err.(error), rolledbackErr) {
		return fmt.Errorf("rollback failed: %w", err.(error))
	}
	return nil
}

func (c *commitTreeFinalizer) SignalFinalize() error {
	c.finalizeOnce.Do(func() {
		close(c.finalizeOrRollback)
	})
	return nil
}

func (c *commitTreeFinalizer) Finalize() (storetypes.CommitID, error) {
	if err := c.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}

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

func (c *CommitTree) LatestVersion() uint32 {
	return c.treeStore.LatestVersion()
}

func (c *CommitTree) GetVersion(version int64) (internal.TreeReader, error) {
	root, err := c.treeStore.RootAtVersion(uint32(version))
	if err != nil {
		return internal.TreeReader{}, err
	}
	return internal.NewTreeReader(root), nil
}

func (c *CommitTree) CacheWrap() storetypes.CacheWrap {
	return c.Latest().CacheWrap()
}

func (c *CommitTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	return c.Latest().CacheWrapWithTrace(w, tc)
}

func (c *CommitTree) Close() error {
	return c.treeStore.Close()
}

var _ storetypes.CacheWrapper = (*CommitTree)(nil)
