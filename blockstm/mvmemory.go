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
	storage              MultiStore
	scheduler            *Scheduler
	stores               map[storetypes.StoreKey]int
	data                 []MVStore
	lastWrittenLocations []atomic.Pointer[MultiLocations]
	lastReadSet          []atomic.Pointer[MultiReadSet]
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
		data[i] = NewMVStore(key)
	}

	mv := &MVMemory{
		storage:              storage,
		scheduler:            scheduler,
		stores:               stores,
		data:                 data,
		lastWrittenLocations: make([]atomic.Pointer[MultiLocations], block_size),
		lastReadSet:          make([]atomic.Pointer[MultiReadSet], block_size),
	}

	// init with pre-estimates
	for txn, est := range estimates {
		mv.rcuUpdateWrittenLocations(TxnIndex(txn), est)
		mv.ConvertWritesToEstimates(TxnIndex(txn))
	}

	return mv
}

func (mv *MVMemory) Record(version TxnVersion, view *MultiMVMemoryView) bool {
	newLocations := view.ApplyWriteSet(version)
	wroteNewLocation := mv.rcuUpdateWrittenLocations(version.Index, newLocations)
	mv.lastReadSet[version.Index].Store(view.ReadSet())
	return wroteNewLocation
}

// newLocations are sorted
func (mv *MVMemory) rcuUpdateWrittenLocations(txn TxnIndex, newLocations MultiLocations) bool {
	var wroteNewLocation bool

	prevLocations := mv.readLastWrittenLocations(txn)
	for i, newLoc := range newLocations {
		prevLoc, ok := prevLocations[i]
		if !ok {
			if len(newLocations[i]) > 0 {
				wroteNewLocation = true
			}
			continue
		}

		DiffOrderedList(prevLoc, newLoc, func(key Key, is_new bool) bool {
			if is_new {
				wroteNewLocation = true
			} else {
				mv.data[i].Delete(key, txn)
			}
			return true
		})
	}

	// delete all the keys in un-touched stores
	for i, prevLoc := range prevLocations {
		if _, ok := newLocations[i]; ok {
			continue
		}

		for _, key := range prevLoc {
			mv.data[i].Delete(key, txn)
		}
	}

	mv.lastWrittenLocations[txn].Store(&newLocations)
	return wroteNewLocation
}

func (mv *MVMemory) ConvertWritesToEstimates(txn TxnIndex) {
	for i, locations := range mv.readLastWrittenLocations(txn) {
		for _, key := range locations {
			mv.data[i].WriteEstimate(key, txn)
		}
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

func (mv *MVMemory) readLastWrittenLocations(txn TxnIndex) MultiLocations {
	p := mv.lastWrittenLocations[txn].Load()
	if p != nil {
		return *p
	}
	return nil
}

func (mv *MVMemory) WriteSnapshot(storage MultiStore) {
	for name, i := range mv.stores {
		mv.data[i].SnapshotToStore(storage.GetStore(name))
	}
}

// View creates a view for a particular transaction.
func (mv *MVMemory) View(txn TxnIndex) *MultiMVMemoryView {
	return NewMultiMVMemoryView(mv.stores, mv.newMVView, txn)
}

func (mv *MVMemory) newMVView(name storetypes.StoreKey, txn TxnIndex) MVView {
	i := mv.stores[name]
	return NewMVView(i, mv.storage.GetStore(name), mv.GetMVStore(i), mv.scheduler, txn)
}

func (mv *MVMemory) GetMVStore(i int) MVStore {
	return mv.data[i]
}
