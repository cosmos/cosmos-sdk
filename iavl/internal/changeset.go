package internal

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type Changeset struct {
	files        *ChangesetFiles
	treeStore    *TreeStore
	readerRef    atomic.Pointer[ChangesetReaderRef]
	sealed       atomic.Bool
	compacted    atomic.Pointer[Changeset]
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
	err = cs.OpenNewReader()
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// TryPinReader attempts to pin the active ChangesetReader.
func (ch *Changeset) TryPinReader() (*ChangesetReader, Pin) {
	for {
		pinner := ch.readerRef.Load()
		if pinner == nil {
			if compacted := ch.compacted.Load(); compacted != nil {
				// changeset was compacted, try the new one
				return compacted.TryPinReader()
			}
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

func (ch *Changeset) OpenNewReader() error {
	newRdr, err := NewChangesetReader(ch)
	if err != nil {
		return fmt.Errorf("failed to create new changeset reader: %w", err)
	}

	var newPinner *ChangesetReaderRef
	if newRdr != nil {
		newPinner = &ChangesetReaderRef{rdr: newRdr}
	}
	existing := ch.readerRef.Swap(newPinner)
	if existing != nil {
		existing.Evict()
	}
	return nil
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

func (ch *Changeset) MarkCompacted(compacted *Changeset) {
	ch.compacted.Store(compacted)
	// evict our reader since we won't be needed anymore
	existing := ch.readerRef.Swap(nil)
	if existing != nil {
		existing.Evict()
	}
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
