package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/cosmos/btree"
	"github.com/jellydator/ttlcache/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"cosmossdk.io/log/v2"
)

// TreeStore is the storage engine for a single IAVL tree. It manages:
//   - The current in-memory root (atomic pointer, swapped on each commit via SaveRoot)
//   - Changesets: the on-disk data directories containing WAL, checkpoint, and node data files.
//     Changesets are organized by version range and may be compacted together over time.
//   - The ChangesetWriter: writes WAL entries, checkpoints, and node data for new commits
//   - The Checkpointer: background goroutine that writes periodic tree snapshots to disk
//   - The cleanup proc: background goroutine that disposes old readers and deletes compacted changesets
//   - A root-by-version cache: caches historical tree roots for repeated historical queries
//
// On startup, load() (in tree_store_load.go) reconstructs the tree by loading changesets,
// finding the latest checkpoint, and replaying WAL entries forward to the committed version.
//
// TreeStore is the lower-level type — CommitTree wraps it to provide the commit/rollback
// protocol used by CommitBranch and CommitFinalizer.
type TreeStore struct {
	dir    string
	opts   TreeOptions
	logger log.Logger

	root           atomic.Pointer[versionedRoot]
	lastCheckpoint atomic.Uint32
	rootMemUsage   atomic.Int64

	currentWriter         *ChangesetWriter
	checkpointer          *Checkpointer
	cleanupProc           *cleanupProc
	lastCheckpointVersion uint32
	shouldCheckpoint      bool
	shouldRollover        bool

	changesetsByVersion btree.Map[uint32, *Changeset]
	changesetsLock      sync.RWMutex
	lastNodeIDsAssigned chan struct{}

	rootByVersionCache *ttlcache.Cache[uint32, *NodePointer]
}

type versionedRoot struct {
	version uint32
	root    *NodePointer
}

// NewTreeStore creates a TreeStore by loading existing data from dir (if any) and
// initializing background goroutines for checkpointing, cleanup, and cache management.
func NewTreeStore(dir string, opts TreeOptions, logger log.Logger) (*TreeStore, error) {
	ts := &TreeStore{
		dir:          dir,
		opts:         opts,
		logger:       logger,
		checkpointer: NewCheckpointer(NewBasicEvictor(opts.LeafEvictDepth, opts.BranchEvictDepth)),
		rootByVersionCache: ttlcache.New[uint32, *NodePointer](
			ttlcache.WithCapacity[uint32, *NodePointer](opts.RootCacheSize),
			ttlcache.WithTTL[uint32, *NodePointer](opts.RootCacheExpiry),
		),
		cleanupProc: newCleanupProc(logger),
	}
	// start automatic cache cleanup
	go ts.rootByVersionCache.Start()
	// start cleanup proc
	ts.cleanupProc.Start(context.Background())

	err := ts.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load tree store: %w", err)
	}

	err = ts.initNewWriter()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// StagedVersion returns the next version that will be committed (current version + 1).
func (ts *TreeStore) StagedVersion() uint32 {
	return ts.root.Load().version + 1
}

func (ts *TreeStore) initNewWriter() error {
	var err error
	ts.currentWriter, err = NewChangesetWriter(ts.dir, ts.StagedVersion(), ts)
	if err != nil {
		return fmt.Errorf("failed to create new changeset writer: %w", err)
	}

	ts.changesetsLock.Lock()
	ts.changesetsByVersion.Set(ts.currentWriter.changeset.files.StartVersion(), ts.currentWriter.Changeset())
	ts.changesetsLock.Unlock()
	return nil
}

// GetRootForUpdate returns the root of the current tree waiting for any node ID assignment happening in the
// background to complete before returning the root.
// NodeID's get assigned in a background thread AFTER commit when we are writing a checkpoint.
// Checkpoint writing itself can proceed WHILE we work with the root returned from this function,
// but if we return BEFORE NodeID's are assigned we will not track orphans properly, and generally
// there will be unsynchronized access to NodeID fields.
// The context parameter passed to this function allows us to return early in the case of a rollback.
func (ts *TreeStore) GetRootForUpdate(ctx context.Context) (*NodePointer, error) {
	if nodeIDsAssigned := ts.lastNodeIDsAssigned; nodeIDsAssigned != nil {
		// If we have a channel that will signal that nodes have been assigned, that means
		// the NodeID assignment go routine is currently running and we must wait.
		select {
		// If the ctx finishes before NodeID's are assigned, we're rolling back so just return and let the caller handle the error.
		case <-ctx.Done():
			return nil, ctx.Err()
		// Otherwise wait for the NodeIDs to be assigned and then clear the lastNodeIDsAssigned field because we know the go routine has returned.
		case <-nodeIDsAssigned:
			ts.lastNodeIDsAssigned = nil
		}
	}
	return ts.root.Load().root, nil
}

// SaveRoot atomically swaps the in-memory root to newRoot, caches the previous root,
// and triggers checkpointing or orphan tracking as needed based on the checkpoint interval.
func (ts *TreeStore) SaveRoot(ctx context.Context, newRoot *NodePointer, mutationCtx *MutationContext) error {
	// sanity check
	lastRoot := ts.root.Load()
	stagedVersion := lastRoot.version + 1
	newVersion := mutationCtx.version
	if newVersion != stagedVersion {
		return fmt.Errorf("mutation context version %d does not match staged version %d", newVersion, stagedVersion)
	}
	// save last root to cache
	if ts.opts.RootCacheSize > 0 {
		ts.rootByVersionCache.Set(lastRoot.version, lastRoot.root, ttlcache.DefaultTTL)
	}
	// store new root and increment version
	swapped := ts.root.CompareAndSwap(lastRoot, &versionedRoot{
		version: newVersion,
		root:    newRoot,
	})
	if !swapped {
		return fmt.Errorf("concurrent root update detected, fatal concurrency error! expected version %d", stagedVersion)
	}

	writer := ts.currentWriter
	if ts.shouldCheckpoint {
		checkpoint := ts.lastCheckpoint.Add(1)
		nodeIDsAssigned := make(chan struct{})
		go func() {
			defer close(nodeIDsAssigned)
			_, span := tracer.Start(ctx, "AssignNodeIDs")
			defer span.End()

			AssignNodeIDs(newRoot, checkpoint)
		}()
		ts.lastNodeIDsAssigned = nodeIDsAssigned

		err := ts.checkpointer.Checkpoint(mutationCtx, writer, newRoot, checkpoint, nodeIDsAssigned, ts.shouldRollover)
		if err != nil {
			return fmt.Errorf("failed to checkpoint changeset: %w", err)
		}
		ts.lastCheckpointVersion = newVersion
		if ts.shouldRollover {
			err = ts.initNewWriter()
			if err != nil {
				return fmt.Errorf("failed to create new changeset writer after rollover: %w", err)
			}
			ts.shouldRollover = false
		}
	} else if ts.shouldRollover {
		return fmt.Errorf("cannot rollover without checkpointing")
	} else {
		ts.shouldRollover = writer.WALWriter().Size() >= int(ts.opts.ChangesetRolloverSize)
		// just mark orphans
		err := ts.checkpointer.QueueOrphans(mutationCtx)
		if err != nil {
			return fmt.Errorf("failed to mark orphans for changeset: %w", err)
		}
	}

	checkpointInterval := ts.opts.CheckpointInterval
	nextStagedVersion := newVersion + 1
	versionsSinceLastCheckpoint := nextStagedVersion - ts.lastCheckpointVersion
	ts.shouldCheckpoint = ts.shouldRollover ||
		(checkpointInterval > 0 &&
			versionsSinceLastCheckpoint >= uint32(ts.opts.CheckpointInterval))

	return nil
}

// WriteWALUpdates writes the given node updates to the current WAL file for the staged version.
// If fsync is true, the WAL is synced to disk before returning.
func (ts *TreeStore) WriteWALUpdates(ctx context.Context, updates []NodeUpdate, fsync bool) error {
	version := ts.StagedVersion()
	walWriter := ts.currentWriter.WALWriter()

	err := walWriter.WriteWALVersion(ctx, uint64(version), updates, ts.ShouldCheckpoint())
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if fsync {
		err = walWriter.Sync()
		if err != nil {
			return err
		}
	} else {
		err = walWriter.writer.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// RollbackWAL truncates the current WAL back to the last committed version.
func (ts *TreeStore) RollbackWAL() error {
	return ts.currentWriter.WALWriter().Rollback()
}

func (ts *TreeStore) ShouldCheckpoint() bool {
	return ts.shouldCheckpoint
}

func (ts *TreeStore) ChangesetForCheckpoint(checkpoint uint32) *Changeset {
	return ts.checkpointer.ChangesetByCheckpoint(checkpoint)
}

func (ts *TreeStore) Latest() (uint32, *NodePointer) {
	latest := ts.root.Load()
	return latest.version, latest.root
}

func (ts *TreeStore) LatestVersion() uint32 {
	return ts.root.Load().version
}

// RootAtVersion returns the root NodePointer for a historical version of the tree.
// This is used for historical queries (CacheMultiStoreWithVersion) and during rollback.
//
// The algorithm: find the latest checkpoint at or before targetVersion, then replay WAL
// entries forward from that checkpoint to reach the exact target version.
// Results are cached (rootByVersionCache) to avoid repeated WAL replay for popular versions.
func (ts *TreeStore) RootAtVersion(targetVersion uint32) (*NodePointer, error) {
	latest := ts.root.Load()
	if latest == nil && targetVersion == 0 {
		// empty tree at version 0
		return nil, nil
	}
	if latest != nil && targetVersion == latest.version {
		// fast path for latest version
		return latest.root, nil
	}

	ctx, span := tracer.Start(context.Background(),
		"TreeStore.RootAtVersion",
		trace.WithAttributes(attribute.Int64("targetVersion", int64(targetVersion))),
	)
	defer span.End()

	// check cache first
	if ts.opts.RootCacheSize > 0 {
		if item := ts.rootByVersionCache.Get(targetVersion); item != nil {
			return item.Value(), nil
		}
	}

	root, err := ts.loadRootAtVersion(ctx, targetVersion)
	if err != nil {
		return nil, err
	}

	// save to cache
	if ts.opts.RootCacheSize > 0 {
		ts.rootByVersionCache.Set(targetVersion, root, ttlcache.DefaultTTL)
	}

	return root, nil
}

// loadRootAtVersion reconstructs the tree root at a specific historical version by:
//  1. Finding the latest checkpoint at or before targetVersion (via checkpointForVersion).
//  2. If the checkpoint IS the target version, return its root directly.
//  3. Otherwise, replay WAL entries forward from the checkpoint version through successive
//     changesets until we reach the target version.
//
// This may cross changeset boundaries — if a changeset's WAL doesn't reach the target version,
// we move to the next changeset and continue replaying.
func (ts *TreeStore) loadRootAtVersion(ctx context.Context, targetVersion uint32) (*NodePointer, error) {
	// find the latest checkpoint root that is <= targetVersion
	res, err := ts.checkpointForVersion(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find checkpoint for version %d: %w", targetVersion, err)
	}

	if res.Version == targetVersion {
		return res.Root, nil
	}

	// then ascend through each changeset and replay the WAL until we reach the desired version
	curVersion := res.Version
	root := res.Root
	for {
		changeset := ts.changesetForVersion(curVersion + 1)
		if changeset == nil {
			return nil, fmt.Errorf("no changeset found for version %d", curVersion+1)
		}

		prevVersion := curVersion
		root, curVersion, err = ReplayWALForQuery(ctx, root, changeset.files.WALFile(), curVersion, targetVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to replay WAL for version %d: %w", targetVersion, err)
		}
		if curVersion == targetVersion {
			return root, nil
		}
		// sanity check to make sure we made progress
		if curVersion <= prevVersion {
			return nil, fmt.Errorf("replay did not advance version from %d", curVersion)
		}
	}
}

// checkpointForVersion finds the latest checkpoint root at or before the given version.
// It searches backwards through changesets: for each changeset, it asks the ChangesetReader
// for the best checkpoint at or before the version. If no checkpoint is found in that changeset,
// it tries the previous one (by looking at startVersion - 1).
// If we reach the beginning of history (startVersion <= 1) with no checkpoint, we return a
// zero-value (empty tree at version 0) so the caller can replay the WAL from scratch.
func (ts *TreeStore) checkpointForVersion(version uint32) (CheckpointRootInfo, error) {
	for {
		changeset := ts.changesetForVersion(version)
		if changeset == nil {
			return CheckpointRootInfo{}, fmt.Errorf("no changeset found for version %d", version)
		}
		rdr, pin := changeset.TryPinReader()
		if rdr == nil {
			pin.Unpin()
			return CheckpointRootInfo{}, fmt.Errorf("changeset reader is not available for version %d", version)
		}

		res := rdr.CheckpointForVersion(version)
		pin.Unpin()

		if res.Version != 0 {
			// for a valid checkpoint
			return res, nil
		}

		startVersion := changeset.Files().StartVersion()
		if startVersion <= 1 {
			// beginning of history, return zero value (empty tree at version 0)
			// so the caller can replay the WAL from the start
			return CheckpointRootInfo{}, nil
		}
		// try an earlier changeset
		version = startVersion - 1
	}
}

func (ts *TreeStore) changesetForVersion(version uint32) *Changeset {
	ts.changesetsLock.RLock()
	defer ts.changesetsLock.RUnlock()

	var res *Changeset
	ts.changesetsByVersion.Descend(version, func(_ uint32, cs *Changeset) bool {
		res = cs
		return false // Take the first (highest) entry <= version
	})
	return res
}

func (ts *TreeStore) LockOrphanProc() {
	ts.checkpointer.orphanProc.Lock()
}

func (ts *TreeStore) UnlockOrphanProc() {
	ts.checkpointer.orphanProc.Unlock()
}

// Close shuts down the checkpointer and cleanup goroutines, seals the current changeset
// writer, closes all changeset readers, and stops the root cache.
func (ts *TreeStore) Close() error {
	// TODO save a checkpoint before closing if needed
	errs := []error{
		ts.checkpointer.Close(),
		ts.currentWriter.Seal(),
		ts.cleanupProc.Close(),
	}
	ts.changesetsByVersion.Ascend(0, func(version uint32, cs *Changeset) bool {
		errs = append(errs, cs.Close())
		return true
	})
	ts.rootByVersionCache.Stop() // stop automatic cache cleanup
	return errors.Join(errs...)
}

func (ts *TreeStore) addToDisposalQueue(existing *ChangesetReaderRef) {
	ts.cleanupProc.AddDisposal(existing)
}

func (ts *TreeStore) addToDeletionQueue(ch *Changeset) {
	ts.cleanupProc.AddDeletion(ch)
}

const memNodeOverhead = int64(unsafe.Sizeof(MemNode{})) + int64(unsafe.Sizeof(NodePointer{}))*2 + 32 /* hash size */
