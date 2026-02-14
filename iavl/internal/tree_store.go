package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/tidwall/btree"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/jellydator/ttlcache/v3"
)

type TreeStoreOptions struct {
	ChangesetRolloverSize int
	LeafEvictDepth        uint8
	BranchEvictDepth      uint8
	CheckpointInterval    int
	MemoryBudget          int64
	RootCacheSize         uint64
	RootCacheExpiry       time.Duration
}

type TreeStore struct {
	dir  string
	opts TreeStoreOptions

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

func NewTreeStore(dir string, opts TreeStoreOptions) (*TreeStore, error) {
	ts := &TreeStore{
		dir:          dir,
		opts:         opts,
		checkpointer: NewCheckpointer(NewBasicEvictor(opts.LeafEvictDepth, opts.BranchEvictDepth)),
		rootByVersionCache: ttlcache.New[uint32, *NodePointer](
			// cache up to 10 recent roots by version
			ttlcache.WithCapacity[uint32, *NodePointer](opts.RootCacheSize),
			// default ttl of 5 seconds
			ttlcache.WithTTL[uint32, *NodePointer](opts.RootCacheExpiry),
		),
		cleanupProc: newCleanupProc(),
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

func (ts *TreeStore) GetRootForUpdate(ctx context.Context) (*NodePointer, error) {
	if nodeIDsAssigned := ts.lastNodeIDsAssigned; nodeIDsAssigned != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-nodeIDsAssigned:
			ts.lastNodeIDsAssigned = nil
		}
	}
	return ts.root.Load().root, nil
}

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
		ts.shouldRollover = writer.WALWriter().Size() >= ts.opts.ChangesetRolloverSize
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

func (ts *TreeStore) WriteWALUpdates(ctx context.Context, updates []KVUpdate, fsync bool) error {
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
		root, curVersion, err = ReplayWAL(ctx, root, changeset.files.WALFile(), curVersion, targetVersion)
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

func (ts *TreeStore) Describe() TreeDescription {
	ts.changesetsLock.RLock()
	defer ts.changesetsLock.RUnlock()

	changesetDescs := make([]ChangesetDescription, 0, ts.changesetsByVersion.Len())
	ts.changesetsByVersion.Ascend(0, func(version uint32, cs *Changeset) bool {
		rdr, pin := cs.TryPinReader()
		defer pin.Unpin()
		if rdr != nil {
			changesetDescs = append(changesetDescs, rdr.Describe())
		} else {
			changesetDescs = append(changesetDescs, ChangesetDescription{
				StartVersion: version,
				Incomplete:   true,
			})
		}
		return true
	})

	totalBytes := 0
	for _, cs := range changesetDescs {
		totalBytes += cs.TotalBytes
	}
	version, root := ts.Latest()
	desc := TreeDescription{
		Version:                 version,
		LatestCheckpointVersion: ts.lastCheckpointVersion,
		LatestCheckpoint:        ts.lastCheckpoint.Load(),
		LatestSavedCheckpoint:   ts.checkpointer.LatestSavedCheckpoint(),
		TotalBytes:              totalBytes,
		Changesets:              changesetDescs,
	}
	if root != nil {
		desc.RootID = root.id
	}
	return desc
}

const memNodeOverhead = int64(unsafe.Sizeof(MemNode{})) + int64(unsafe.Sizeof(NodePointer{}))*2 + 32 /* hash size */
