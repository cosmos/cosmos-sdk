package internal

import "sync/atomic"

type ChangesetHandle struct {
	changeset atomic.Pointer[ChangesetReaderPinner]
	treeStore *TreeStore
}

func (h *ChangesetHandle) TryPinReader() (*ChangesetReader, Pin) {
	pinner := h.changeset.Load()
	if pinner == nil {
		return nil, nil
	}
	return pinner.TryPin()
}

func (h *ChangesetHandle) PinCompactedReader(version uint32) (*ChangesetReader, Pin) {
	return h.treeStore.GetChangesetForVersion(version)
}

func (h *ChangesetHandle) SwapActiveReader(newRdr *ChangesetReader) *ChangesetReaderPinner {
	var newPinner *ChangesetReaderPinner
	if newRdr != nil {
		newPinner = &ChangesetReaderPinner{rdr: newRdr}
	}
	return h.changeset.Swap(newPinner)
}
