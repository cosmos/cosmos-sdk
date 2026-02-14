package blockstm

import storetypes "cosmossdk.io/store/types"

const ViewsPreAllocate = 4

// MultiMVMemoryView don't need to be thread-safe, there's a dedicated instance for each tx execution.
type MultiMVMemoryView struct {
	mv     *MVMemory
	stores map[storetypes.StoreKey]int
	views  map[storetypes.StoreKey]MVView
	txn    TxnIndex
}

var _ MultiStore = (*MultiMVMemoryView)(nil)

func NewMultiMVMemoryView(
	mv *MVMemory,
	txn TxnIndex,
) *MultiMVMemoryView {
	return &MultiMVMemoryView{
		stores: mv.stores,
		views:  make(map[storetypes.StoreKey]MVView, ViewsPreAllocate),
		txn:    txn,
		mv:     mv,
	}
}

func (mv *MultiMVMemoryView) getViewOrInit(name storetypes.StoreKey) MVView {
	view, ok := mv.views[name]
	if !ok {
		view = mv.mv.newMVView(name, mv.txn)
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

func (mv *MultiMVMemoryView) ApplyWriteSet(version TxnVersion) bool {
	var wroteNewLocation bool
	for _, view := range mv.views {
		if view.ApplyWriteSet(version) {
			wroteNewLocation = true
		}
	}

	// handle un-touched stores
	for name, i := range mv.stores {
		if _, ok := mv.views[name]; !ok {
			mv.mv.GetMVStore(i).ConsolidateEmpty(version.Index)
		}
	}
	return wroteNewLocation
}
