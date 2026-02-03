package internal

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type Changeset struct {
	treeStore    *TreeStore
	readerRef    atomic.Pointer[ChangesetReaderRef]
	sealed       atomic.Bool
	files        *ChangesetFiles
	orphanWriter *OrphanWriter
}

func NewChangeset(treeStore *TreeStore, files *ChangesetFiles) *Changeset {
	return &Changeset{
		treeStore:    treeStore,
		files:        files,
		orphanWriter: NewOrphanWriter(files.OrphansFile()),
	}
}

func OpenChangeset(treeStore *TreeStore, dir string) (*Changeset, error) {
	files, err := OpenChangesetFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open changeset files: %w", err)
	}
	cs := NewChangeset(treeStore, files)
	cr, err := NewChangesetReader(cs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize changeset reader: %w", err)
	}
	cs.readerRef.Store(&ChangesetReaderRef{rdr: cr})
	return cs, nil
}

// TryPinReader attempts to pin the active ChangesetReader.
// If this Changeset was just compacted, it may return (nil, NoopPin{}).
func (ch *Changeset) TryPinReader() (*ChangesetReader, Pin) {
	for {
		pinner := ch.readerRef.Load()
		if pinner == nil {
			// changeset was compacted, no active reader
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

//func (ch *Changeset) LoadRoot(version uint32) (*NodePointer, error) {
//	rdr, pin := ch.TryPinReader()
//	defer pin.Unpin()
//	if rdr == nil {
//		return nil, fmt.Errorf("changeset reader is not available")
//	}
//	cpVersion, cpRoot, err := rdr.FindNearestCheckpoint(version)
//	if err != nil {
//		return nil, fmt.Errorf("failed to find nearest checkpoint for version %d: %w", version, err)
//	}
//	if cpVersion == version {
//		return cpRoot, nil
//	}
//
//	return ReplayWAL(context.Background(), cpRoot, ch.files.WALFile(), cpVersion, version)
//}

func (ch *Changeset) swapActiveReader(newRdr *ChangesetReader) {
	var newPinner *ChangesetReaderRef
	if newRdr != nil {
		newPinner = &ChangesetReaderRef{rdr: newRdr}
	}
	existing := ch.readerRef.Swap(newPinner)
	if existing != nil {
		existing.Evict()
	}
}

func (ch *Changeset) MarkOrphan(version uint32, nodeId NodeID) error {
	err := ch.orphanWriter.WriteOrphan(version, nodeId)
	if err != nil {
		return fmt.Errorf("failed to write orphan node: %w", err)
	}

	// TODO track orphan stats
	//info := ch.files.info
	//if nodeId.IsLeaf() {
	//	info.LeafOrphans++
	//	info.LeafOrphanVersionTotal += uint64(version)
	//} else {
	//	info.BranchOrphans++
	//	info.BranchOrphanVersionTotal += uint64(version)
	//}

	return nil
}

func (ch *Changeset) Close() error {
	readerRef := ch.readerRef.Load()
	if readerRef != nil {
		readerRef.Evict()
	}
	return errors.Join(
		ch.orphanWriter.Flush(), // TODO sync
		ch.files.Close(),
	)
}

func (ch *Changeset) Files() *ChangesetFiles {
	return ch.files
}
