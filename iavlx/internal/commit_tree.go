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
		close(committer.done)
	}()
	return committer
}

var rolledbackErr = errors.New("commit rolled back")

func (c *commitTreeFinalizer) commit(ctx context.Context, updates iter.Seq[cachekv.Update[[]byte]], updateCount int) error {
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

	prepareRes, err := c.prepareCommit(ctx, updates, updateCount)
	if err != nil {
		rbErr := c.treeStore.RollbackWAL()
		if rbErr != nil {
			return fmt.Errorf("commit failed: %w; rollback failed: %w", err, rbErr)
		}
		return fmt.Errorf("%w; root cause %w", rolledbackErr, err)
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

	if root != nil {
		if mem := root.Mem.Load(); mem != nil {
			// instrument tree size, height and orphan count
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
		// Here choose a minimum bucket size to avoid too many goroutines being created for small buckets
		const minBucketSize = 64
		numCPUs := runtime.NumCPU()
		// Choose the number of workers based on the number of CPUs and minimum bucket size.
		numWorkers := min(numCPUs, max(1, n/minBucketSize))
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
	if err := c.WaitForWAL(); err != nil {
		return nil, err
	}
	walWriteLatency.Record(ctx, time.Since(startWaitForWAL).Milliseconds())
	span.AddEvent("WAL write returned")

	select {
	case <-c.finalizeOrRollback:
	case <-ctx.Done():
	}

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

func (c *commitTreeFinalizer) WaitForWAL() error {
	c.walOnce.Do(func() {
		c.walErr = <-c.walDone
	})
	return c.walErr
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
