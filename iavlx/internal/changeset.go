package internal

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Changeset represents the WAL log and saved checkpoints for a given range of versions in a tree.
// It manages the lifecycle of the changeset files and readers, and tracks when it has been compacted into a new changeset.
type Changeset struct {
	files             *ChangesetFiles
	treeStore         *TreeStore
	readerRef         atomic.Pointer[ChangesetReaderRef]
	activeReaderCount atomic.Int32
	sealed            atomic.Bool
	compacted         atomic.Pointer[Changeset]
	orphanWriter      *StructWriter[OrphanEntry]
}

// NewChangeset creates a new Changeset with the given TreeStore and ChangesetFiles.
func NewChangeset(treeStore *TreeStore, files *ChangesetFiles) (*Changeset, error) {
	const orphanWriterBufSize = SizeOrphanEntry * 342 // roughly 4kb, but aligned to orphan size, so the file size is correct in a crash scenario
	cs := &Changeset{
		treeStore:    treeStore,
		files:        files,
		orphanWriter: NewStructWriterSize[OrphanEntry](files.OrphansFile(), orphanWriterBufSize),
	}
	err := cs.openNewReader()
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset reader: %w", err)
	}
	return cs, nil
}

// OpenChangeset opens existing changeset files in the given directory.
func OpenChangeset(treeStore *TreeStore, dir string, autoRepair bool) (*Changeset, error) {
	files, err := OpenChangesetFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}
	cs, err := NewChangeset(treeStore, files)
	if err != nil {
		return nil, err
	}
	// mark as sealed since this is an existing changeset that won't be written to anymore
	// otherwise the compactor will skip it
	cs.sealed.Store(true)
	// TODO if this verification check ends up being expensive we can instead verify only the last changeset when loading the tree store, but it shouldn't be a problem to do this for every changeset
	err = cs.VerifyAndFix(autoRepair) // attempt to fix any issues with the changeset before we start using it
	if err != nil {
		return nil, fmt.Errorf("failed to verify and fix changeset: %w", err)
	}
	return cs, nil
}

// TryPinReader attempts to pin the active ChangesetReader for this changeset,
// or for the changeset that this changeset was compacted into.
// This method will always return a valid pin which should be unpinned,
// but a nil reader may be returned in the case where the changeset has been closed.
func (ch *Changeset) TryPinReader() (*ChangesetReader, Pin) {
	rdr, pin := ch.TryPinUncompactedReader()
	if rdr != nil {
		return rdr, pin
	}
	if compacted := ch.compacted.Load(); compacted != nil {
		// changeset was compacted, try the new one
		return compacted.TryPinReader()
	}
	return nil, Pin{}
}

// TryPinUncompactedReader attempts to pin a reader for only this changeset and
// will return a nil reader if this changeset has been compacted or closed.
// This is used when we ONLY want to read THIS changeset and not its compaction.
func (ch *Changeset) TryPinUncompactedReader() (*ChangesetReader, Pin) {
	for {
		pinner := ch.readerRef.Load()
		if pinner == nil {
			return nil, Pin{}
		}
		rdr, pin := pinner.TryPin()
		if rdr == nil {
			// evicted, we probably have a new reader now, try again
			continue
		}
		return rdr, pin
	}
}

// TreeStore returns the TreeStore to which this changeset belongs.
func (ch *Changeset) TreeStore() *TreeStore {
	return ch.treeStore
}

// Compacted returns the Changeset this has been compacted into if this
// Changeset was compacted, otherwise it returns nil.
func (ch *Changeset) Compacted() *Changeset {
	return ch.compacted.Load()
}

// openNewReader opens a new reader for this changeset.
// This should only be used internally by methods managing changeset writing and loading.
// Other callers shouldn't need to call this method.
// Generally this would be called when we have written more data to the changeset and
// want to swap out the current mmap's for mmap's that reference the newly written data.
func (ch *Changeset) openNewReader() error {
	// Open a new reader with new mmap's.
	newRdr, err := NewChangesetReader(ch)
	if err != nil {
		return fmt.Errorf("failed to create new changeset reader: %w", err)
	}

	var newPinner *ChangesetReaderRef
	if newRdr != nil { // should never be nil here, not sure why we need this check??
		// Create a pinner reference to this changeset reader to do reference counting
		newPinner = &ChangesetReaderRef{rdr: newRdr, changeset: ch}
	}
	// Swap the reader reference to point to the new reader.
	existing := ch.readerRef.Swap(newPinner)
	if existing != nil {
		// If there is an existing reader, evict it.
		existing.Evict()
	}
	// Increment the active reader count. This will be used to know when we can safely delete this changeset if it's compacted.
	ch.activeReaderCount.Add(1)
	return nil
}

// markCompacted marks this changeset as compacted and sets its compacted pointer to the compacted changeset.
// Only the compactor should call this method.
func (ch *Changeset) markCompacted(compacted *Changeset) {
	ch.compacted.Store(compacted)
	// evict our reader since we won't be needed anymore
	existing := ch.readerRef.Swap(nil)
	if existing != nil {
		existing.Evict()
	}
	ch.orphanWriter = nil
	ch.treeStore.addToDeletionQueue(ch)
}

// Close closes this changeset.
func (ch *Changeset) Close() error {
	readerRef := ch.readerRef.Swap(nil)
	if readerRef != nil {
		readerRef.Evict()
	}
	if ch.orphanWriter != nil {
		if err := ch.orphanWriter.Flush(); err != nil {
			return fmt.Errorf("failed to flush orphan writer: %w", err)
		}
	}
	return ch.files.Close()
}

func (ch *Changeset) Files() *ChangesetFiles {
	return ch.files
}

func (ch *Changeset) OrphanWriter() *StructWriter[OrphanEntry] {
	if ch.orphanWriter == nil {
		return ch.compacted.Load().OrphanWriter()
	}
	return ch.orphanWriter
}

func (ch *Changeset) TryDelete(ctx context.Context) (bool, error) {
	if ch.compacted.Load() == nil {
		// not compacted yet
		return false, nil
	}
	if ch.activeReaderCount.Load() > 0 {
		// readers still active, can't delete yet
		return false, nil
	}
	ch.treeStore.logger.InfoContext(ctx, "deleting changeset", "dir", ch.files.Dir())
	return true, ch.files.DeleteFiles()
}

// VerifyAndFix performs integrity checks on the changeset data and attempts to fix any issues that it can.
func (ch *Changeset) VerifyAndFix(autoRepair bool) error {
	cr, pin := ch.TryPinUncompactedReader()
	defer pin.Unpin()
	if cr == nil {
		return fmt.Errorf("changeset reader is not available for verification")
	}

	err := cr.Verify()
	if err != nil {
		if !autoRepair {
			return fmt.Errorf("changeset verification failed and autoRepair is disabled, cannot fix: %w", err)
		}

		ch.treeStore.logger.Warn("changeset verification failed, attempting to fix if possible", "dir", ch.files.Dir(), "error", err)

		if err := ch.RollbackLastCheckpoint(cr); err != nil {
			return fmt.Errorf("failed to rollback last checkpoint during changeset verification: %w", err)
		}
	}
	return nil
}

// RollbackLastCheckpoint rolls back the most recent checkpoint in this changeset,
// truncating checkpoints.dat, branches.dat, leaves.dat, and kv.dat to the previous checkpoint's offsets.
// The caller must provide a pinned ChangesetReader for the current state.
func (ch *Changeset) RollbackLastCheckpoint(cr *ChangesetReader) error {
	cpCount := cr.checkpointsInfo.Count()
	if cpCount == 0 {
		return fmt.Errorf("no checkpoints to roll back in changeset: %s", ch.files.Dir())
	}
	newCpCount := cpCount - 1
	newLastCheckpointOffset := newCpCount * CheckpointInfoSize
	err := RollbackFileToOffset(ch.files.CheckpointsFile(), int64(newLastCheckpointOffset))
	if err != nil {
		return fmt.Errorf("failed to truncate checkpoint info file: %w", err)
	}

	var newBranchesOffset, newLeavesOffset, newKVDataOffset int64
	if newCpCount > 0 {
		// if we have another checkpoint, use its offsets
		// otherwise everything goes to zero
		lastGoodInfo := cr.checkpointsInfo.UnsafeItem(newCpCount - 1)
		if !lastGoodInfo.VerifyCRC32() {
			return fmt.Errorf("previous checkpoint also has invalid CRC32, cannot fix changeset: %s",
				ch.files.Dir())
		}

		newBranchesOffset = int64((lastGoodInfo.Branches.StartOffset + lastGoodInfo.Branches.Count) * SizeBranch)
		newLeavesOffset = int64((lastGoodInfo.Leaves.StartOffset + lastGoodInfo.Leaves.Count) * SizeLeaf)
		newKVDataOffset = int64(lastGoodInfo.KVEndOffset)
	}

	err = RollbackFileToOffset(ch.files.BranchesFile(), newBranchesOffset)
	if err != nil {
		return fmt.Errorf("failed to truncate branches file: %w", err)
	}

	err = RollbackFileToOffset(ch.files.LeavesFile(), newLeavesOffset)
	if err != nil {
		return fmt.Errorf("failed to truncate leaves file: %w", err)
	}

	err = RollbackFileToOffset(ch.files.KVDataFile(), newKVDataOffset)
	if err != nil {
		return fmt.Errorf("failed to truncate kv data file: %w", err)
	}

	// open a new reader to update our in-memory state to reflect the rolled back files
	err = ch.openNewReader()
	if err != nil {
		return fmt.Errorf("failed to open new reader after rollback: %w", err)
	}

	return nil
}
