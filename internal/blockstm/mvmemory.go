package blockstm

import (
	"sync/atomic"

	storetypes "cosmossdk.io/store/types"
)

type (
	Locations      []Key // keys are sorted
	MultiLocations map[int]Locations
)

// MVMemory implements `Algorithm 2 The MVMemory module`
type MVMemory struct {
	storage     MultiStore
	scheduler   *Scheduler
	stores      map[storetypes.StoreKey]int
	data        []MVStore
	lastReadSet []atomic.Pointer[MultiReadSet]
}

func NewMVMemory(
	block_size int, stores map[storetypes.StoreKey]int,
	storage MultiStore, scheduler *Scheduler,
) *MVMemory {
	return NewMVMemoryWithEstimates(block_size, stores, storage, scheduler, nil)
}

func NewMVMemoryWithEstimates(
	block_size int, stores map[storetypes.StoreKey]int,
	storage MultiStore, scheduler *Scheduler, estimates []MultiLocations,
) *MVMemory {
	data := make([]MVStore, len(stores))
	for key, i := range stores {
		data[i] = NewMVStore(key, block_size)
	}

	mv := &MVMemory{
		storage:     storage,
		scheduler:   scheduler,
		stores:      stores,
		data:        data,
		lastReadSet: make([]atomic.Pointer[MultiReadSet], block_size),
	}

	// init with pre-estimates
	for txn, est := range estimates {
		for store, locs := range est {
			mv.data[store].InitWithEstimates(TxnIndex(txn), locs)
		}
	}

	return mv
}

func (mv *MVMemory) Record(version TxnVersion, view *MultiMVMemoryView) bool {
	wroteNewLocation := view.ApplyWriteSet(version)
	mv.lastReadSet[version.Index].Store(view.ReadSet())
	return wroteNewLocation
}

func (mv *MVMemory) ConvertWritesToEstimates(txn TxnIndex) {
	for _, data := range mv.data {
		data.ConvertWritesToEstimates(txn)
	}
}

// ClearEstimates removes estimate marks for canceled transactions.
func (mv *MVMemory) ClearEstimates(txn TxnIndex) {
	for _, data := range mv.data {
		data.ClearEstimates(txn)
	}
}

func (mv *MVMemory) ValidateReadSet(txn TxnIndex) bool {
	// Invariant: at least one `Record` call has been made for `txn`
	rs := *mv.lastReadSet[txn].Load()
	for store, readSet := range rs {
		if !mv.data[store].ValidateReadSet(txn, readSet) {
			return false
		}
	}
	return true
}

func (mv *MVMemory) WriteSnapshot(storage MultiStore) {
	for name, i := range mv.stores {
		mv.data[i].SnapshotToStore(storage.GetStore(name))
	}
}

// View creates a view for a particular transaction.
func (mv *MVMemory) View(txn TxnIndex) *MultiMVMemoryView {
	return NewMultiMVMemoryView(mv, txn)
}

func (mv *MVMemory) newMVView(name storetypes.StoreKey, txn TxnIndex) MVView {
	i := mv.stores[name]
	return NewMVView(i, mv.storage.GetStore(name), mv.GetMVStore(i), mv.scheduler, txn)
}

func (mv *MVMemory) GetMVStore(i int) MVStore {
	return mv.data[i]
}
