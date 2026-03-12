package blockstm

import storetypes "cosmossdk.io/store/types"

const ViewsPreAllocate = 4

// MultiMVMemoryView don't need to be thread-safe, there's a dedicated instance for each tx execution.
type MultiMVMemoryView struct {
	stores    map[storetypes.StoreKey]int
	views     map[storetypes.StoreKey]MVView
	newMVView func(storetypes.StoreKey, TxnIndex) MVView
	txn       TxnIndex
}

var _ MultiStore = (*MultiMVMemoryView)(nil)

func NewMultiMVMemoryView(
	stores map[storetypes.StoreKey]int,
	newMVView func(storetypes.StoreKey, TxnIndex) MVView,
	txn TxnIndex,
) *MultiMVMemoryView {
	return &MultiMVMemoryView{
		stores:    stores,
		views:     make(map[storetypes.StoreKey]MVView, ViewsPreAllocate),
		newMVView: newMVView,
		txn:       txn,
	}
}

func (mv *MultiMVMemoryView) getViewOrInit(name storetypes.StoreKey) MVView {
	view, ok := mv.views[name]
	if !ok {
		view = mv.newMVView(name, mv.txn)
		mv.views[name] = view
	}
	return view
}

func (mv *MultiMVMemoryView) GetStore(name storetypes.StoreKey) storetypes.Store {
	return mv.getViewOrInit(name)
}

func (mv *MultiMVMemoryView) GetKVStore(name storetypes.StoreKey) storetypes.KVStore {
	return mv.GetStore(name).(storetypes.KVStore)
}

func (mv *MultiMVMemoryView) GetObjKVStore(name storetypes.StoreKey) storetypes.ObjKVStore {
	return mv.GetStore(name).(storetypes.ObjKVStore)
}

func (mv *MultiMVMemoryView) ReadSet() *MultiReadSet {
	rs := make(MultiReadSet, len(mv.views))
	for key, view := range mv.views {
		rs[mv.stores[key]] = view.ReadSet()
	}
	return &rs
}

func (mv *MultiMVMemoryView) ApplyWriteSet(version TxnVersion) MultiLocations {
	newLocations := make(MultiLocations, len(mv.views))
	for key, view := range mv.views {
		newLocations[mv.stores[key]] = view.ApplyWriteSet(version)
	}
	return newLocations
}

// CountReads returns the total number of reads across all stores
func (mv *MultiMVMemoryView) CountReads() int {
	count := 0
	for _, view := range mv.views {
		rs := view.ReadSet()
		count += len(rs.Reads)
		for _, iter := range rs.Iterators {
			count += len(iter.Reads)
		}
	}
	return count
}

// CountWrites returns the total number of writes across all stores
func (mv *MultiMVMemoryView) CountWrites() int {
	count := 0
	for _, view := range mv.views {
		count += view.WriteCount()
	}
	return count
}
