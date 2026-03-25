package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"iter"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	ics23 "github.com/cosmos/ics23/go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	errorsmod "cosmossdk.io/errors"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/store/v2/cachekv"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/store/v2/types/kv"
)

type CommitTree struct {
	commitMutex sync.Mutex

	opts TreeOptions

	treeStore *TreeStore

	lastCommitId storetypes.CommitID
}

func NewCommitTree(dir string, opts TreeOptions, logger log.Logger) (*CommitTree, error) {
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("creating tree directory: %w", err)
	}

	treeStore, err := NewTreeStore(dir, opts, logger)
	if err != nil {
		return nil, fmt.Errorf("creating tree store: %w", err)
	}

	return &CommitTree{
		opts:      opts,
		treeStore: treeStore,
	}, nil
}

var emptyHash = sha256.New().Sum(nil)

func rootHash(ctx context.Context, rootPtr *NodePointer) ([]byte, error) {
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

	scheduler := NewAsyncHashScheduler(ctx, int32(runtime.NumCPU()))
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
	walDone            chan error
	walOnce            sync.Once
	walErr             error
}

func (c *CommitTree) startCommit(ctx context.Context, updates iter.Seq[cachekv.Update[[]byte]], updateCount int) *commitTreeFinalizer {
	cancelCtx, cancel := context.WithCancel(ctx)
	committer := &commitTreeFinalizer{
		CommitTree:         c,
		cancel:             cancel,
		finalizeOrRollback: make(chan struct{}),
		hashReady:          make(chan struct{}),
		done:               make(chan struct{}),
		walDone:            make(chan error, 1),
	}
	// We start the commit go routine here.
	// All the commit work happens in the background, and we must signal it to finalize or rollback.
	go func() {
		// Prevent context leak: WithCancel registers a child in the parent context's tree,
		// and that registration is only cleaned up when cancel() is called.
		// Safe here because all ctx.Err() checks are inside commit(), which has already
		// returned by the time this defer fires. On rollback, cancel() is called first;
		// calling it again here is a no-op.
		defer cancel()
		err := committer.commit(cancelCtx, updates, updateCount)
		if err != nil {
			committer.err.Store(err)
		}
		// At the end we close the committer.done channel to communicate that the commit go routine has completed.
		close(committer.done)
	}()
	return committer
}

var rolledbackErr = errors.New("commit rolled back")

// commit does the actual work of commiting a tree or rolling back if there is an error or if the caller cancels the context.
func (c *commitTreeFinalizer) commit(ctx context.Context, updates iter.Seq[cachekv.Update[[]byte]], updateCount int) error {
	// We lock the mutex because only one commit operation can happen at a time.
	c.commitMutex.Lock()
	defer c.commitMutex.Unlock()

	stagedVersion := c.treeStore.StagedVersion()
	ctx, span := tracer.Start(ctx, "CommitTree.Commit",
		trace.WithAttributes(
			attribute.String("treeName", c.opts.TreeName),
			attribute.Int64("version", int64(stagedVersion)),
			attribute.Int("updateCount", updateCount),
		),
	)
	defer span.End()

	// prepareCommit does almost all the work of committing including
	// - updating the tree,
	// - computing hashes,
	// - writing the WAL
	prepareRes, err := c.prepareCommit(ctx, updates, updateCount)
	if err != nil {
		// If there was an error (either signaled by the user or otherwise) we rollback.
		// Even if we already rolled back the WAL due to an error, it's idempotent to call rollback again.
		rbErr := c.treeStore.RollbackWAL()
		if rbErr != nil {
			return fmt.Errorf("commit failed: %w; rollback failed: %w", err, rbErr)
		}
		return fmt.Errorf("%w; root cause %w", rolledbackErr, err)
	}

	// At this point we are finalizing the commit, there is no rolling back from here forward.
	root := prepareRes.root
	// This updates the in-memory tree root and version and also starts the checkpoint go routine
	// if a checkpoint needs to be taken at this point.
	// Checkpointing DOES NOT need to complete before the commit is finalized.
	// Checkpointing is considered an optimization to make loading the tree from disk more efficient
	// and it DOES NOT need the same level of durability as writing the WAL which we have already done.
	err = c.treeStore.SaveRoot(
		ctx,
		root,
		prepareRes.mutationCtx,
	)
	if err != nil {
		// TODO: we should consider if a rollback is needed/possible here - in what conditions would we have an error here
		return err
	}

	c.lastCommitId = c.workingHash

	if root != nil {
		if mem := root.Mem.Load(); mem != nil {
			// Instrument tree size, height and orphan count
			span.SetAttributes(
				attribute.Int64("size", mem.Size()),
				attribute.Int64("height", int64(mem.Height())),
			)
		}
	}

	return nil
}

type prepareCommitResult struct {
	root        *NodePointer
	mutationCtx *MutationContext
}

// prepareCommit does most of the work for updating the tree and committing its new state,
// while still allowing it to be cleanly rolled back
func (c *commitTreeFinalizer) prepareCommit(ctx context.Context, updates iter.Seq[cachekv.Update[[]byte]], updateCount int) (*prepareCommitResult, error) {
	ctx, span := tracer.Start(ctx, "PrepareCommit")
	defer span.End()

	stagedVersion := c.treeStore.StagedVersion()
	// Create a mutation context where cowVersion == stagedVersion, this means nodes created in this version
	// will be mutated in place, but any nodes from previous versions will be safely copied.
	// This is okay because nodes from this version will not be shared until the commit is done.
	mutationCtx := NewMutationContext(stagedVersion, stagedVersion)

	// We start by creating a KVUpdate for every update in the commit.
	// The main difference between KVUpdate and Update, is that KVUpdate actually creates the leaf MemNode's
	// that will end up in the new version of the tree.
	// TODO pre-allocate a decent sized slice that we reuse and reset across commits
	nodeUpdates := make([]KVUpdate, 0, updateCount)
	if updates != nil { // updates can be nil for empty commits
		for update := range updates {
			if update.Delete {
				nodeUpdates = append(nodeUpdates, KVUpdate{DeleteKey: update.Key})
			} else {
				nodeUpdates = append(
					nodeUpdates,
					KVUpdate{
						// This actually creates a new leaf MemNode that will end up in the new tree once all updates have been applied.
						SetNode: mutationCtx.NewLeafNode(update.Key, update.Value),
					},
				)
			}
		}
	}

	// Our first optimization is to start writing the WAL right away in a background go routine.
	// The WAL is simply a log of all the sets and deletes we are going to apply to the tree.
	// We haven't even mutated the tree at all yet, but we're going to log these updates to
	// disk "ahead" of updating the actual tree thus "write-ahead log".
	// Because this WAL is append-only, it is really easy to rollback the commit
	// by just truncating the WAL file back to its previous size.
	go func(ctx context.Context) {
		_, walSpan := tracer.Start(ctx, "WALWrite")
		defer walSpan.End()
		defer close(c.walDone) // notify listeners that we are done writing the WAL
		c.walDone <- c.treeStore.WriteWALUpdates(ctx, nodeUpdates, !c.opts.DisableWALFsync)
	}(ctx) // pass ctx as a parameter to avoid a data race: the main goroutine reassigns ctx later (e.g. in tracer.Start) while this goroutine may still be reading it

	// We always wait for the WAL goroutine to finish before returning from prepareCommit.
	// Without this, commit() may call WALWriter.Rollback() while the WAL goroutine is still running,
	// which causes a data race between two go routines trying to mutate the WAL.
	// The WALWriter is actually listening to context cancellation, so if a rollback is initiated
	// externally before WAL writing has completed, the WALWriter will actually do the rollback
	// itself and return immediately - when commit() calls WALWriter.Rollback() later the second call is idempotent.
	defer func() { _ = c.WaitForWAL() }()

	// Our second optimization is to spin off a go routine which computes leaf hashes in parallel,
	// also before we have done any mutations to the root of the tree.
	// Because leaf hashes can be computed independently, we can hash all the leaf nodes in parallel
	// batches to speed things up a bit more.
	leafHashErr := make(chan error, 1)
	go func() {
		defer close(leafHashErr)
		_, hashLeavesSpan := tracer.Start(ctx, "LeafHashCompute")
		defer hashLeavesSpan.End()

		n := len(nodeUpdates)
		if n == 0 {
			return
		}

		// This concurrency algorithm is rather naive that assumes leaf nodes updates are evenly distributed.
		// If we see consistent latency here, we could tune the algorithm to accomodate extra large key-value
		// pairs that slow down hashing.
		// Minimum items per worker to avoid goroutine overhead dominating small workloads.
		const minBucketSize = 64
		numCPUs := runtime.NumCPU()
		// Cap workers at CPU count, but also ensure each worker gets at least minBucketSize items.
		// When n < minBucketSize, n/minBucketSize == 0, so max(1, ...) guarantees at least one worker.
		numWorkers := min(numCPUs, max(1, n/minBucketSize))
		// Ceiling division: distribute items as evenly as possible, with the last bucket potentially smaller.
		bucketSize := (n + numWorkers - 1) / numWorkers

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			start := i * bucketSize
			end := min(start+bucketSize, n)
			wg.Add(1)
			go func(updates []KVUpdate) {
				defer wg.Done()
				if ctx.Err() != nil {
					return
				}
				for _, nu := range updates {
					if setNode := nu.SetNode; setNode != nil {
						// SyncHashScheduler will compute the hash in the current go routine which is already running.
						if _, err := setNode.ComputeHash(SyncHashScheduler{}); err != nil {
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

	// Now we actually do the work of updating the tree while the other go routines
	// happily write the WAL and compute leaf node hashes.
	root, err := c.treeStore.GetRootForUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("tree store not ready for update: %w", err)
	}

	// Create a span to see how long it takes to update the root of the tree.
	_, nodeUpdatesSpan := tracer.Start(ctx, "ApplyNodeUpdates")
	// Here we do the actual work of updating the root of the tree by applying each set or delete, in order,
	// to the root of the tree.
	// Note that after each set or delete we actually have a new root because the tree is essentially immutable.
	// We only have the next version root after applying all updates.
	// We use the mutation context we created above which has the correct version and cowVersion set and
	// which is tracking any orphans that we create during this update process.
	// (Orphans are basically nodes that were in this tree but are no longer in the new tree and which we may want to
	// delete from disk sometime later.)
	// Regarding performance of updating the tree, in benchmarks this depends heavily on the size
	// of the tree and how much of it can be kept in RAM.
	// Even if the tree is mostly kept in memory, a larger, deeper tree will take longer to update because
	// we simply need to do more updating and balancing of tree nodes for every key-value update.
	// But if we need to retrieve nodes from disk, that is an order of magnitude slower than reading them
	// from memory however large the tree is.
	// Also, note that there are two layers here - even if nodes are technically "read from disk", we may
	// actually be reading them from the OS mmap cache and not the actual physical storage.
	// Reading from the OS mmap cache is actually only a little bit slower than reading them from the heap,
	// but it's harder to control and measure.
	// The key lesson here for large trees is the more memory the better.
	// If there is more memory available, node operators can increase the eviction depths in options
	// to retain more nodes in memory.
	for _, nu := range nodeUpdates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var err error
		if setNode := nu.SetNode; setNode != nil {
			// The setNode that we are passing in here is the leaf MemNode we created above
			// at the beginning of prepareCommit in the nodeUpdates array.
			// While we are inserting it into the tree here:
			// - the WAL writer go routine may be reading it, writing its key and value to the WAL, and saving
			//   the key and value offsets to the MemNode - this is okay because concurrent reads of key and value are okay,
			//   and we won't read those offsets until much later after commit is done (if we write a checkpoint)
			// - one of the leaf hash go routines may be reading its key, value and version and writing the hash value
			//   in the MemNode.hash field - this is also okay because key, value and version are already set,
			//   and we won't be reading the hash
			root, _, err = SetRecursive(root, setNode, mutationCtx)
			if err != nil {
				return nil, err
			}
		} else {
			_, root, _, err = RemoveRecursive(root, nu.DeleteKey, mutationCtx)
			if err != nil {
				return nil, err
			}
		}
	}
	// Track that we're done updating the tree so we can see how long this step takes.
	nodeUpdatesSpan.End()

	// This timestamp is simply for us to track a metric about whether it took longer for leaf hashes to compute
	// than it took to update the tree.
	// In the happy path our wait time should be 0, but if not, it alerts operators that we may need to optimize
	// leaf node hashing
	startWaitForLeafHashes := time.Now()

	// Wait for the leaf node hashing go routines to return.
	// These go routines must complete before we can start computing the root hash - otherwise there's a race condition.
	select {
	case err := <-leafHashErr: // This channel returns nil if the leaf hashing completed successfully
		if err != nil {
			return nil, err
		}
	case <-ctx.Done(): // Also listen for ctx.Done() in case we rollback in the meantime
		return nil, ctx.Err()
	}
	span.AddEvent("leaf hash compute returned")

	// Track any latency caused by waiting for leaf hashes to finish, which ideally should be negligible or zero.
	leafHashLatency.Record(ctx, time.Since(startWaitForLeafHashes).Milliseconds())

	// Now we compute the root hash and create a span to track how long that takes.
	ctx, rootHashSpan := tracer.Start(ctx, "ComputeRootHash")
	// Compute the root hash.
	// This rootHash function attempts to speed things up by computing the hash of different branches of the tree in parallel.
	hash, err := rootHash(ctx, root)
	if err != nil {
		return nil, err
	}
	// Save the root hash.
	c.workingHash = storetypes.CommitID{
		Hash:    hash,
		Version: int64(stagedVersion),
	}
	// Close the hashReady go channel to notify anyone who was waiting for the root hash that it is ready and that
	// they can read the hashReady field.
	// This allows for the root hash of a multi-tree to be computed before WAL writing finishes (if it actually takes longer)
	// and it also allows FinalizeBlock to return before the full commit is done.
	close(c.hashReady)
	// Close the span so we track how long it took to compute the root hash.
	rootHashSpan.End()

	// Like we did when waiting for leaf hash go routines to return,
	// we create a timestamp to measure how long it takes us to wait for the WAL writing go routine to return.
	startWaitForWAL := time.Now()
	// Wait for the WAL go routine to return.
	if err := c.WaitForWAL(); err != nil {
		return nil, err
	}
	// Track any latency we observed when waiting for WAL writing to complete.
	// Ideally, this value is zero, but if it isn't, it likely communicates to the
	// node operator that they need to get a faster storage device.
	// The WAL is an append-only file, so on a fast SSD, the WAL should be written long before everything else is done.
	// Our work should primarily be CPU-bound here - we should be able to speed things up with more CPUs
	// to split up hashing and with faster CPUs to compute the root faster, but if the WAL writing is slow,
	// that should be easily fixable with better storage.
	walWriteLatency.Record(ctx, time.Since(startWaitForWAL).Milliseconds())
	span.AddEvent("WAL write returned")

	// Here we wait for one of two signals, either:
	// - the caller has explicitly request to finalize the commit or rollback in which case the finalizeOrRollback channel will be closed first
	// - the context cancelled and ctx.Done() returns first
	// We are waiting because the caller needs to know that we are ready to either finalize or rollback once this method returns
	select {
	case <-c.finalizeOrRollback:
	case <-ctx.Done():
	}

	return &prepareCommitResult{
		root:        root,
		mutationCtx: mutationCtx,
	}, ctx.Err() // we return ctx.Err() here because if we are rolling back it will definitely be non-nil at this point
}

// WaitForHash waits for the hash to be ready and returns it or an error.
// We do not know whether the commit has been or will be finalized when WaitForHash returns.
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

// WaitForWAL waits for WAL writing to complete.
func (c *commitTreeFinalizer) WaitForWAL() error {
	// The sync.Once used here ensures that we drain the walDone channel exactly once
	// so that multiple callers can call WaitForWAL, and they will all get the same result without
	// blocking on the channel receive multiple times.
	c.walOnce.Do(func() {
		c.walErr = <-c.walDone
	})
	return c.walErr
}

// StartFinalize signals that the commit should proceed with finalization and
// blocks until the hash is ready.
func (c *commitTreeFinalizer) StartFinalize() (storetypes.CommitID, error) {
	if err := c.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}
	return c.WaitForHash()
}

// Rollback rolls back the commit.
func (c *commitTreeFinalizer) Rollback() error {
	// Rollback works via context cancellation.
	// We start by canceling the context which communicates to all worker go routines that they should
	// stop any in-progress work and return.
	c.cancel()
	// We also close the finalizeOrRollback channel in case anyone is waiting for that channel to close,
	// but only after context cancellation is used to trigger the rollback.
	c.finalizeOnce.Do(func() {
		close(c.finalizeOrRollback)
	})
	// We wait for
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

func (c *CommitTree) LatestVersion() uint32 {
	return c.treeStore.LatestVersion()
}

func (c *CommitTree) Latest() TreeReader {
	version, root := c.treeStore.Latest()
	return NewTreeReader(version, root)
}

func (c *CommitTree) GetVersion(version uint32) (TreeReader, error) {
	root, err := c.treeStore.RootAtVersion(version)
	if err != nil {
		return TreeReader{}, err
	}
	return NewTreeReader(version, root), nil
}

func (c *CommitTree) Close() error {
	return c.treeStore.Close()
}

var _ storetypes.Queryable = (*CommitTree)(nil)

// Query handles query paths for a single IAVL-backed store.
func (c *CommitTree) Query(req *storetypes.RequestQuery) (_ *storetypes.ResponseQuery, err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			queryLatency.Record(context.Background(), time.Since(start).Milliseconds())
		}
	}()

	height, err := c.queryHeight(req.Height)
	if err != nil {
		return &storetypes.ResponseQuery{}, err
	}

	res := &storetypes.ResponseQuery{Height: height}

	tree, err := c.GetVersion(uint32(height))
	if err != nil {
		// Keep this as response metadata to match existing query behavior.
		res.Log = err.Error()
		return res, nil
	}

	switch req.Path {
	case "/key":
		if len(req.Data) == 0 {
			return &storetypes.ResponseQuery{}, errorsmod.Wrap(storetypes.ErrTxDecode, "query cannot be zero length")
		}

		key := req.Data
		res.Key = key

		value, err := tree.GetErr(key)
		if err != nil {
			return &storetypes.ResponseQuery{}, err
		}
		res.Value = value

		if req.Prove {
			res.ProofOps, err = iavlProofOps(&tree, key, value != nil)
			if err != nil {
				return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "failed to create proof: %v", err)
			}
		}

		return res, nil

	case "/subspace":
		subspace := req.Data
		res.Key = subspace

		iterator := storetypes.KVStorePrefixIterator(tree, subspace)
		pairs := kv.Pairs{
			Pairs: make([]kv.Pair, 0),
		}
		for ; iterator.Valid(); iterator.Next() {
			pairs.Pairs = append(pairs.Pairs, kv.Pair{
				Key:   bytes.Clone(iterator.Key()),
				Value: bytes.Clone(iterator.Value()),
			})
		}
		if err := iterator.Close(); err != nil {
			return &storetypes.ResponseQuery{}, fmt.Errorf("failed to close iterator: %w", err)
		}

		bz, err := pairs.Marshal()
		if err != nil {
			panic(fmt.Errorf("failed to marshal KV pairs: %w", err))
		}

		res.Value = bz

		return res, nil

	default:
		return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrUnknownRequest, "unexpected query path: %v", req.Path)
	}
}

func (c *CommitTree) queryHeight(reqHeight int64) (int64, error) {
	if reqHeight == 0 {
		return int64(c.LatestVersion()), nil
	}
	if reqHeight < 0 {
		return 0, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "invalid query height: %d", reqHeight)
	}
	if reqHeight > int64(^uint32(0)) {
		return 0, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "query height %d exceeds max supported height", reqHeight)
	}

	return reqHeight, nil
}

func iavlProofOps(tree *TreeReader, key []byte, exists bool) (*cmtprotocrypto.ProofOps, error) {
	var (
		commitmentProof *ics23.CommitmentProof
		err             error
	)

	if exists {
		commitmentProof, err = tree.GetMembershipProof(key)
	} else {
		commitmentProof, err = tree.GetNonMembershipProof(key)
	}
	if err != nil {
		return nil, err
	}

	op := storetypes.NewIavlCommitmentOp(key, commitmentProof)
	return &cmtprotocrypto.ProofOps{Ops: []cmtprotocrypto.ProofOp{op.ProofOp()}}, nil
}

func (c *CommitTree) compact(ctx context.Context, retainVersion uint32) error {
	return RunCompactor(ctx, c.treeStore, CompactionOptions{
		RetainVersion:          retainVersion,
		CompactionRolloverSize: c.opts.CompactionRolloverSize,
	})
}

func (c *CommitTree) RollbackToVersion(ctx context.Context, version uint32) error {
	c.commitMutex.Lock()
	defer c.commitMutex.Unlock()

	err := c.treeStore.rollbackToVersion(version)
	if err != nil {
		return fmt.Errorf("rolling back tree store to version %d: %w", version, err)
	}

	//// save metadata before closing since we'll need it to reopen
	//dir, opts, logger := c.treeStore.dir, c.treeStore.opts, c.treeStore.logger
	//
	//// just close and reload the tree store
	//err = c.treeStore.Close()
	//if err != nil {
	//	return fmt.Errorf("closing tree store: %w", err)
	//}
	//newStore, err := NewTreeStore(dir, opts, logger)
	//if err != nil {
	//	return fmt.Errorf("creating tree store: %w", err)
	//}
	//c.treeStore = newStore

	return nil
}
