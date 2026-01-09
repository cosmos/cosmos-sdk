package iavlx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tidwall/btree"
)

type TreeStore struct {
	logger *slog.Logger
	dir    string

	currentWriter         *ChangesetWriter
	currentChangesetEntry *changesetEntry // Entry for the current batch being written
	changesets            *btree.Map[uint32, *changesetEntry]
	changesetsMapLock     sync.RWMutex
	savedVersion          atomic.Uint32 // Last version with a readable changeset
	stagedVersion         uint32        // Latest written version (may not be readable yet)

	opts Options

	syncQueue *NonBlockingQueue[*ChangesetWriter]
	syncDone  chan error

	cleanupProc *cleanupProc
}

type markOrphansReq struct {
	version uint32
	orphans [][]NodeID
}

type changesetEntry struct {
	changeset atomic.Pointer[Changeset]
}

func NewTreeStore(dir string, options Options, logger *slog.Logger) (*TreeStore, error) {
	ts := &TreeStore{
		dir:           dir,
		changesets:    &btree.Map[uint32, *changesetEntry]{},
		logger:        logger,
		opts:          options,
		stagedVersion: 1,
	}

	err := ts.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load existing changesets: %w", err)
	}

	err = ts.initNewWriter()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize first writer: %w", err)
	}

	ts.cleanupProc = newCleanupProc(ts)

	if options.FsyncEnabled() {
		ts.syncQueue = NewNonBlockingQueue[*ChangesetWriter]()
		ts.syncDone = make(chan error)
		go ts.syncProc()
	}

	return ts, nil
}

func (ts *TreeStore) initNewWriter() error {
	stagedVersion := ts.savedVersion.Load() + 1
	writer, err := NewChangesetWriter(ts.dir, stagedVersion, ts)
	if err != nil {
		return fmt.Errorf("failed to create changeset writer: %w", err)
	}
	ts.currentWriter = writer

	return nil
}

func (ts *TreeStore) getChangesetEntryForVersion(version uint32) *changesetEntry {
	ts.changesetsMapLock.RLock()
	defer ts.changesetsMapLock.RUnlock()

	var res *changesetEntry
	// Find the changeset with the highest start version <= the requested version
	ts.changesets.Descend(version, func(key uint32, cs *changesetEntry) bool {
		res = cs
		return false // Take the first (highest) entry <= version
	})
	return res
}

func (ts *TreeStore) getChangesetForVersion(version uint32) *Changeset {
	cs := ts.getChangesetEntryForVersion(version)
	if cs == nil {
		return nil
	} else {
		return cs.changeset.Load()
	}
}

func (ts *TreeStore) ReadK(nodeId NodeID) (key []byte, err error) {
	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	cs.Pin()
	defer cs.Unpin()

	if cs == nil {
		return nil, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}

	var offset uint32
	if nodeId.IsLeaf() {
		leaf, err := cs.ResolveLeaf(nodeId, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve leaf %s: %w", nodeId.String(), err)
		}
		offset = leaf.KeyOffset
	} else {
		branch, err := cs.ResolveBranch(nodeId, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve branch %s: %w", nodeId.String(), err)
		}
		offset = branch.KeyOffset
	}

	return cs.ReadK(nodeId, offset)
}

func (ts *TreeStore) ReadKV(nodeId NodeID) (key, value []byte, err error) {
	if !nodeId.IsLeaf() {
		return nil, nil, fmt.Errorf("node %s is not a leaf", nodeId.String())
	}

	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	if cs == nil {
		return nil, nil, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}

	cs.Pin()
	defer cs.Unpin()

	leaf, err := cs.ResolveLeaf(nodeId, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve leaf %s: %w", nodeId.String(), err)
	}

	return cs.ReadKV(nodeId, leaf.KeyOffset)
}

func (ts *TreeStore) ReadV(nodeId NodeID) ([]byte, error) {
	// TODO reduce code duplication with ReadKV

	if !nodeId.IsLeaf() {
		return nil, fmt.Errorf("node %s is not a leaf", nodeId.String())
	}

	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	if cs == nil {
		return nil, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}

	cs.Pin()
	defer cs.Unpin()
	leaf, err := cs.ResolveLeaf(nodeId, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve leaf %s: %w", nodeId.String(), err)
	}

	return cs.ReadV(nodeId, leaf.KeyOffset)
}

func (ts *TreeStore) ResolveLeaf(nodeId NodeID) (LeafLayout, error) {
	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	if cs == nil {
		return LeafLayout{}, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}
	return cs.ResolveLeaf(nodeId, 0)
}

func (ts *TreeStore) ResolveBranch(nodeId NodeID) (BranchLayout, error) {
	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	if cs == nil {
		return BranchLayout{}, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}
	return cs.ResolveBranch(nodeId, 0)
}

func (ts *TreeStore) Resolve(nodeId NodeID, _ uint32) (Node, error) {
	cs := ts.getChangesetForVersion(uint32(nodeId.Version()))
	if cs == nil {
		return nil, fmt.Errorf("no changeset found for version %d", nodeId.Version())
	}

	return cs.Resolve(nodeId, 0)
}

func (ts *TreeStore) ResolveRoot(version uint32) (*NodePointer, error) {
	cs := ts.getChangesetForVersion(version)
	if cs == nil {
		return nil, fmt.Errorf("no changeset found for version %d", version)
	}
	return cs.ResolveRoot(version)
}

func (ts *TreeStore) SavedVersion() uint32 {
	return ts.savedVersion.Load()
}

func (ts *TreeStore) WriteWALUpdates(updates []KVUpdate) error {
	return ts.currentWriter.WriteWALUpdates(updates)
}

func (ts *TreeStore) WriteWALCommit(version uint32) error {
	return ts.currentWriter.WriteWALCommit(version)
}

func (ts *TreeStore) SaveRoot(ctx context.Context, root *NodePointer, totalLeaves, totalBranches uint32) error {
	ctx, span := tracer.Start(ctx, "TreeStore.SaveRoot")
	defer span.End()
	version := ts.stagedVersion
	ts.logger.Debug("saving root", "version", version)
	err := ts.currentWriter.SaveRoot(root, version, totalLeaves, totalBranches)
	if err != nil {
		return err
	}
	ts.stagedVersion++

	currentSize := ts.currentWriter.TotalBytes()
	maxSize := ts.opts.GetChangesetMaxTarget()
	readerInterval := ts.opts.GetReaderUpdateInterval()

	ts.logger.Debug("saved root", "version", version, "changeset_size", currentSize, "max_size", maxSize, "start_version", ts.currentWriter.StartVersion())

	// Queue changeset for async WAL sync if enabled
	if ts.syncQueue != nil {
		select {
		case err := <-ts.syncDone:
			if err != nil {
				return err
			}
		default:
		}
	}

	// Determine if we should create a reader
	shouldCreateReader := false
	shouldSeal := uint64(currentSize) >= maxSize

	startVersion := ts.currentWriter.StartVersion()
	if shouldSeal {
		shouldCreateReader = true
	} else if readerInterval > 0 {
		// Create reader periodically based on interval
		versions := version - startVersion + 1
		if versions%readerInterval == 0 {
			shouldCreateReader = true
		}
	}

	if !shouldCreateReader {
		// Just continue batching without creating reader
		return nil
	}

	// Create reader (either shared or sealed)
	var reader *Changeset
	if shouldSeal {
		// Size limit reached - seal the current batch
		curWriter := ts.currentWriter
		reader, err = curWriter.Seal()
		if err != nil {
			return fmt.Errorf("failed to seal changeset for version %d: %w", version, err)
		}
		if ts.syncQueue != nil {
			// if sync queue is enabled, queue the files for fsync
			ts.syncQueue.Send(curWriter)
		}
	} else {
		// Create shared reader for periodic update
		reader, err = ts.currentWriter.CreatedSharedReader()
		if err != nil {
			return fmt.Errorf("failed to create updated changeset reader: %w", err)
		}
	}

	ts.setActiveReader(startVersion, reader)
	ts.savedVersion.Store(version)

	if shouldSeal {
		ts.currentChangesetEntry = nil // Reset for next batch

		// Create new writer for next batch
		err = ts.initNewWriter()
		if err != nil {
			return fmt.Errorf("failed to initialize new writer after sealing version %d: %w", version, err)
		}
	}

	return nil
}

func (ts *TreeStore) setActiveReader(version uint32, reader *Changeset) {
	if ts.currentChangesetEntry == nil {
		// First time we're creating an entry for this batch
		ts.currentChangesetEntry = &changesetEntry{}
		ts.currentChangesetEntry.changeset.Store(reader)

		// Register at the start version only
		ts.changesetsMapLock.Lock()
		ts.changesets.Set(version, ts.currentChangesetEntry)
		ts.changesetsMapLock.Unlock()
	} else {
		// Update existing entry with new reader
		oldReader := ts.currentChangesetEntry.changeset.Swap(reader)
		if oldReader != nil {
			oldReader.Evict()
			if !oldReader.TryDispose() {
				ts.cleanupProc.addPendingDisposal(oldReader)
			}
		}
	}
}

func (ts *TreeStore) MarkOrphans(version uint32, nodeIds [][]NodeID) {
	ts.cleanupProc.markOrphans(version, nodeIds)
}

func (ts *TreeStore) syncProc() {
	tick := time.NewTicker(ts.opts.GetFsyncInterval())
	defer close(ts.syncDone)
	for {
		<-tick.C
		curWriter := ts.currentWriter
		err := curWriter.SyncWAL()
		if err != nil {
			ts.syncDone <- fmt.Errorf("failed to sync WAL file: %w", err)
			return
		}
		needsSync, closed := ts.syncQueue.MaybeReceive()
		for _, f := range needsSync {
			if err := f.SyncWAL(); err != nil {
				ts.syncDone <- fmt.Errorf("failed to sync WAL file: %w", err)
				return
			}
		}
		if closed {
			return
		}
	}
}

func (ts *TreeStore) Close() error {
	// save the current writer if it has uncommitted data
	startVersion := ts.currentWriter.files.info.StartVersion
	if startVersion != 0 {
		cs, err := ts.currentWriter.Seal()
		if err != nil {
			return fmt.Errorf("failed to seal current changeset on close: %w", err)
		}
		ts.setActiveReader(startVersion, cs)
		ts.savedVersion.Store(ts.currentWriter.files.info.EndVersion)
	}

	ts.cleanupProc.shutdown()

	if ts.syncQueue != nil {
		ts.syncQueue.Close()
		err := <-ts.syncDone
		if err != nil {
			return err
		}
	}

	ts.changesetsMapLock.Lock()

	var errs []error
	ts.changesets.Scan(func(version uint32, entry *changesetEntry) bool {
		errs = append(errs, entry.changeset.Load().Close())
		return true
	})
	return errors.Join(errs...)
}
