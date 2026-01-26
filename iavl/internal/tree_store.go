package internal

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type TreeStoreOptions struct {
	ChangesetRolloverSize int
	EvictDepth            uint8
	CheckpointInterval    int
	MemoryBudget          int64
}

type TreeStore struct {
	dir  string
	opts TreeStoreOptions

	version      atomic.Uint32
	root         atomic.Pointer[NodePointer]
	rootMemUsage atomic.Int64

	currentWriter         *ChangesetWriter
	checkpointer          *Checkpointer
	cleanupProc           *cleanupProc
	lastCheckpointVersion uint32
	shouldCheckpoint      bool
	shouldRollover        bool

	//latestRoots     *btree.Map[uint32, *NodePointer]
	//latestRootsLock sync.RWMutex
}

func NewTreeStore(dir string, opts TreeStoreOptions) (*TreeStore, error) {
	ts := &TreeStore{
		dir:          dir,
		opts:         opts,
		checkpointer: NewCheckpointer(BasicEvictor{EvictDepth: opts.EvictDepth}),
	}

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
	return nil
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer, mutationCtx *MutationContext, nodeIdsAssigned chan struct{}) error {
	ts.root.Store(newRoot)
	version := ts.version.Add(1)

	writer := ts.currentWriter
	if ts.shouldCheckpoint {
		err := ts.checkpointer.Checkpoint(writer, newRoot, version, nodeIdsAssigned, ts.shouldRollover)
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

	ts.cleanupProc.MarkOrphans(mutationCtx.version, mutationCtx.orphans)

	return nil
}

func (ts *TreeStore) WriteWALUpdates(updates []KVUpdate, fsync bool) error {
	version := ts.StagedVersion()
	walWriter := ts.currentWriter.WALWriter()

	err := walWriter.StartVersion(uint64(version))
	if err != nil {
		return err
	}

	err = walWriter.WriteWALUpdates(updates)
	if err != nil {
		return err
	}

	err = walWriter.WriteWALCommit(uint64(version))
	if err != nil {
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

func (ts *TreeStore) ShouldCheckpoint() bool {
	return ts.shouldCheckpoint
}

func (ts *TreeStore) ChangesetForCheckpoint(checkpoint uint32) *Changeset {
	return ts.checkpointer.ChangesetByCheckpoint(checkpoint)
}

func (ts *TreeStore) ForceToDisk() error {
	return fmt.Errorf("ForceToDisk disabled")
}

func (ts *TreeStore) Latest() *NodePointer {
	return ts.root.Load()
}

func (ts *TreeStore) Close() error {
	return ts.checkpointer.Close()
}

func (ts *TreeStore) addToDisposalQueue(existing *ChangesetReaderRef) {
	// TODO
}

const memNodeOverhead = int64(unsafe.Sizeof(MemNode{})) + int64(unsafe.Sizeof(NodePointer{}))*2 + 32 /* hash size */
