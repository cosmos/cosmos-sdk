package internal

import "sync/atomic"

type Changeset struct {
	readerRef atomic.Pointer[ChangesetReaderRef]
	treeStore *TreeStore
}

func NewChangeset(treeStore *TreeStore) *Changeset {
	return &Changeset{treeStore: treeStore}
}

func (h *Changeset) TryPinReader() (*ChangesetReader, Pin) {
	pinner := h.readerRef.Load()
	if pinner == nil {
		return nil, nil
	}
	return pinner.TryPin()
}

func (h *Changeset) PinCompactedReader(layer uint32) (*ChangesetReader, Pin) {
	return h.treeStore.GetChangesetForLayer(layer)
}

func (h *Changeset) SwapActiveReader(newRdr *ChangesetReader) *ChangesetReaderRef {
	var newPinner *ChangesetReaderRef
	if newRdr != nil {
		newPinner = &ChangesetReaderRef{rdr: newRdr}
	}
	return h.readerRef.Swap(newPinner)
}
