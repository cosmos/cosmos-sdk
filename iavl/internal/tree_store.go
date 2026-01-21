package internal

import (
	"fmt"
	"sync/atomic"
)

type TreeStoreOptions struct {
	ChangesetRolloverSize int
	EvictDepth            uint8
	//CheckpointInterval    int
}

type TreeStore struct {
	dir  string
	opts TreeStoreOptions

	version atomic.Uint32
	root    atomic.Pointer[NodePointer]

	currentWriter *ChangesetWriter
	checkpointer  *Checkpointer

	//latestRoots     *btree.Map[uint32, *NodePointer]
	//latestRootsLock sync.RWMutex
}

func NewTreeStore(dir string, opts TreeStoreOptions) (*TreeStore, error) {
	ts := &TreeStore{
		dir:          dir,
		opts:         opts,
		checkpointer: NewCheckpointer(BasicEvictor{EvictDepth: opts.EvictDepth}),
	}
	writer, err := NewChangesetWriter(dir, 1, ts)
	if err != nil {
		return nil, err
	}
	ts.currentWriter = writer

	return ts, nil
}

func (ts *TreeStore) StagedVersion() uint32 {
	return ts.version.Load() + 1
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer) error {
	ts.root.Store(newRoot)
	version := ts.version.Add(1)

	// if we are at rollover size, create new changeset writer
	writer := ts.currentWriter
	if writer.WALWriter().Size() >= ts.opts.ChangesetRolloverSize {
		ts.checkpointer.Checkpoint(writer, newRoot, version, true)
	}
	var err error
	ts.currentWriter, err = NewChangesetWriter(ts.dir, ts.StagedVersion(), ts)
	if err != nil {
		return fmt.Errorf("failed to create new changeset writer: %w", err)
	}

	return nil
}

func (ts *TreeStore) WriteWALUpdates(updates []KVUpdate, fsync bool) error {
	walWriter := ts.currentWriter.WALWriter()
	err := walWriter.WriteWALUpdates(updates)
	if err != nil {
		return err
	}

	version := ts.StagedVersion()

	err = walWriter.WriteWALCommit(uint64(version))
	if err != nil {
		return err
	}

	if fsync {
		err = walWriter.Sync()
		if err != nil {
			return err
		}
	}

	return nil
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
