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
	EvictDepth            uint8
	CheckpointInterval    int
	MemoryBudget          int64
	RootCacheSize         uint64
	RootCacheExpiry       time.Duration
}

type TreeStore struct {
	dir  string
	opts TreeStoreOptions

	version      atomic.Uint32
	checkpoint   atomic.Uint32
	root         atomic.Pointer[NodePointer]
	rootMemUsage atomic.Int64

	currentWriter         *ChangesetWriter
	checkpointer          *Checkpointer
	cleanupProc           *cleanupProc
	lastCheckpointVersion uint32
	shouldCheckpoint      bool
	shouldRollover        bool

	changesetsByVersion *btree.Map[uint32, *Changeset]
	changesetsLock      sync.RWMutex
	lastNodeIDsAssigned chan struct{}

	rootByVersionCache *ttlcache.Cache[uint32, *NodePointer]
}

func NewTreeStore(dir string, opts TreeStoreOptions) (*TreeStore, error) {
	ts := &TreeStore{
		dir:                 dir,
		opts:                opts,
		checkpointer:        NewCheckpointer(BasicEvictor{EvictDepth: opts.EvictDepth}),
		changesetsByVersion: &btree.Map[uint32, *Changeset]{},
		rootByVersionCache: ttlcache.New[uint32, *NodePointer](
			// cache up to 10 recent roots by version
			ttlcache.WithCapacity[uint32, *NodePointer](opts.RootCacheSize),
			// default ttl of 5 seconds
			ttlcache.WithTTL[uint32, *NodePointer](opts.RootCacheExpiry),
		),
	}
	// start automatic cache cleanup
	go ts.rootByVersionCache.Start()

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
	return ts.version.Load() + 1
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
	return ts.root.Load(), nil
}

func (ts *TreeStore) SaveRoot(ctx context.Context, newRoot *NodePointer, mutationCtx *MutationContext) error {
	// sanity check
	if mutationCtx.version != ts.StagedVersion() {
		return fmt.Errorf("mutation context version %d does not match staged version %d", mutationCtx.version, ts.StagedVersion())
	}
	// save last root to cache
	ts.rootByVersionCache.Set(ts.version.Load(), ts.root.Load(), ttlcache.DefaultTTL)
	// store new root and increment version
	ts.root.Store(newRoot)
	version := ts.version.Add(1)

	writer := ts.currentWriter
	if ts.shouldCheckpoint {
		checkpoint := ts.checkpoint.Add(1)
		nodeIDsAssigned := make(chan struct{})
		go func() {
			defer close(nodeIDsAssigned)
			_, span := tracer.Start(ctx, "AssignNodeIDs")
			defer span.End()

			AssignNodeIDs(newRoot, checkpoint)
		}()
		ts.lastNodeIDsAssigned = nodeIDsAssigned

		err := ts.checkpointer.Checkpoint(writer, newRoot, checkpoint, version, nodeIDsAssigned, ts.shouldRollover)
		if err != nil {
			return fmt.Errorf("failed to checkpoint changeset: %w", err)
		}
		ts.lastCheckpointVersion = version
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
	}

	checkpointInterval := ts.opts.CheckpointInterval
	nextStagedVersion := version + 1
	versionsSinceLastCheckpoint := nextStagedVersion - ts.lastCheckpointVersion
	ts.shouldCheckpoint = ts.shouldRollover ||
		(checkpointInterval > 0 &&
			versionsSinceLastCheckpoint >= uint32(ts.opts.CheckpointInterval))

	// TODO cleanup orphans
	//ts.cleanupProc.MarkOrphans(mutationCtx.version, mutationCtx.orphans)

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

func (ts *TreeStore) Latest() *NodePointer {
	return ts.root.Load()
}

func (ts *TreeStore) LatestVersion() uint32 {
	return ts.version.Load()
}

func (ts *TreeStore) RootAtVersion(targetVersion uint32) (*NodePointer, error) {
	ctx, span := tracer.Start(context.Background(),
		"TreeStore.RootAtVersion",
		trace.WithAttributes(attribute.Int64("targetVersion", int64(targetVersion))),
	)
	defer span.End()

	// check cache first
	if item := ts.rootByVersionCache.Get(targetVersion); item != nil {
		return item.Value(), nil
	}

	root, err := ts.loadRootAtVersion(ctx, targetVersion)
	if err != nil {
		return nil, err
	}

	// save to cache
	ts.rootByVersionCache.Set(targetVersion, root, ttlcache.DefaultTTL)

	return root, nil
}

func (ts *TreeStore) loadRootAtVersion(ctx context.Context, targetVersion uint32) (*NodePointer, error) {
	// find the latest checkpoint root that is <= targetVersion
	root, curVersion, err := ts.checkpointForVersion(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find checkpoint for version %d: %w", targetVersion, err)
	}

	if curVersion == targetVersion {
		return root, nil
	}

	// then ascend through each changeset and replay the WAL until we reach the desired version
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

func (ts *TreeStore) checkpointForVersion(version uint32) (cpRoot *NodePointer, cpVersion uint32, err error) {
	for {
		changeset := ts.changesetForVersion(version)
		if changeset == nil {
			return nil, 0, fmt.Errorf("no changeset found for version %d", version)
		}
		rdr, pin := changeset.TryPinReader()
		if rdr == nil {
			pin.Unpin()
			return nil, 0, fmt.Errorf("changeset reader is not available for version %d", version)
		}

		cpRoot, cpVersion = rdr.CheckpointForVersion(version)
		pin.Unpin()

		if cpVersion != 0 {
			return cpRoot, cpVersion, nil
		}

		startVersion := changeset.Files().StartVersion()
		if startVersion <= 1 {
			// we're at the beginning of history, return empty tree
			return nil, 0, nil
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

func (ts *TreeStore) Close() error {
	// TODO save a checkpoint before closing if needed
	errs := []error{
		ts.checkpointer.Close(),
		ts.currentWriter.Seal(),
	}
	ts.changesetsByVersion.Ascend(0, func(version uint32, cs *Changeset) bool {
		errs = append(errs, cs.files.Close())
		return true
	})
	ts.rootByVersionCache.Stop() // stop automatic cache cleanup
	return errors.Join(errs...)
}

func (ts *TreeStore) addToDisposalQueue(existing *ChangesetReaderRef) {
	// TODO
}

const memNodeOverhead = int64(unsafe.Sizeof(MemNode{})) + int64(unsafe.Sizeof(NodePointer{}))*2 + 32 /* hash size */
