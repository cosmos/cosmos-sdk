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
	orphanWriter      *OrphanWriter
}

// NewChangeset creates a new Changeset with the given TreeStore and ChangesetFiles.
func NewChangeset(treeStore *TreeStore, files *ChangesetFiles) (*Changeset, error) {
	cs := &Changeset{
		treeStore:    treeStore,
		files:        files,
		orphanWriter: NewOrphanWriter(files.OrphansFile()),
	}
	err := cs.OpenNewReader()
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset reader: %w", err)
	}
	return cs, nil
}

// OpenChangeset opens existing changeset files in the given directory.
func OpenChangeset(treeStore *TreeStore, dir string) (*Changeset, error) {
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
	return nil, NoopPin{}
}

func (ch *Changeset) TryPinUncompactedReader() (*ChangesetReader, Pin) {
	for {
		pinner := ch.readerRef.Load()
		if pinner == nil {
			return nil, NoopPin{}
		}
		rdr, pin := pinner.TryPin()
		if rdr == nil {
			// evicted, we probably have a new reader now, try again
			continue
		}
		return rdr, pin
	}
}

func (ch *Changeset) TreeStore() *TreeStore {
	return ch.treeStore
}

func (ch *Changeset) Compacted() *Changeset {
	return ch.compacted.Load()
}

func (ch *Changeset) OpenNewReader() error {
	newRdr, err := NewChangesetReader(ch)
	if err != nil {
		return fmt.Errorf("failed to create new changeset reader: %w", err)
	}

	var newPinner *ChangesetReaderRef
	if newRdr != nil {
		newPinner = &ChangesetReaderRef{rdr: newRdr, changeset: ch}
	}
	existing := ch.readerRef.Swap(newPinner)
	if existing != nil {
		existing.Evict()
	}
	ch.activeReaderCount.Add(1)
	return nil
}

func (ch *Changeset) MarkCompacted(compacted *Changeset) {
	ch.compacted.Store(compacted)
	// evict our reader since we won't be needed anymore
	existing := ch.readerRef.Swap(nil)
	if existing != nil {
		existing.Evict()
	}
	ch.orphanWriter = nil
	ch.treeStore.addToDeletionQueue(ch)
}

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

func (ch *Changeset) OrphanWriter() *OrphanWriter {
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
	logger.InfoContext(ctx, "deleting changeset", "dir", ch.files.Dir())
	return true, ch.files.DeleteFiles()
}
