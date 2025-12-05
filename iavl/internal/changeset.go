package internal

import "sync/atomic"

type Changeset struct {
	changeset atomic.Pointer[ChangesetReaderRef]
	treeStore *TreeStore
}

func (h *Changeset) TryPinReader() (*ChangesetReader, Pin) {
	pinner := h.changeset.Load()
	if pinner == nil {
		return nil, nil
	}
	return pinner.TryPin()
}

func (h *Changeset) PinCompactedReader(version uint32) (*ChangesetReader, Pin) {
	return h.treeStore.GetChangesetForVersion(version)
}

func (h *Changeset) SwapActiveReader(newRdr *ChangesetReader) *ChangesetReaderRef {
	var newPinner *ChangesetReaderRef
	if newRdr != nil {
		newPinner = &ChangesetReaderRef{rdr: newRdr}
	}
	return h.changeset.Swap(newPinner)
}
