package iavlx

import (
	"fmt"
	io "io"

	storetypes "cosmossdk.io/store/types"
)

type MultiTree struct {
	latestVersion int64
	trees         []storetypes.CacheWrap      // always ordered by tree name
	treesByKey    map[storetypes.StoreKey]int // index of the trees by name
}

func (t *MultiTree) Write() {
	for _, tree := range t.trees {
		tree.Write()
	}
}

func (t *MultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (t *MultiTree) CacheWrap() storetypes.CacheWrap {
	return t.CacheMultiStore()
}

func (t *MultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	//TODO implement tracing
	return t.CacheWrap()
}

func (t *MultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	wrapped := &MultiTree{
		trees:      make([]storetypes.CacheKVStore, len(t.trees)),
		treesByKey: t.treesByKey,
	}
	for i, tree := range t.trees {
		wrapped.trees[i] = tree.CacheWrap().(storetypes.CacheKVStore)
	}
	return wrapped
}

func (t *MultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	return nil, fmt.Errorf("CacheMultiStoreWithVersion can only be called on CommitMultiStore")
}

func (t *MultiTree) GetStore(key storetypes.StoreKey) storetypes.Store {
	return t.trees[t.treesByKey[key]]
}

func (t *MultiTree) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return t.trees[t.treesByKey[key]]
}

func (t *MultiTree) TracingEnabled() bool {
	return false
}

func (t *MultiTree) SetTracer(w io.Writer) storetypes.MultiStore {
	//TODO implement me
	panic("implement me")
}

func (t *MultiTree) SetTracingContext(context storetypes.TraceContext) storetypes.MultiStore {
	//TODO implement me
	panic("implement me")
}

func (t *MultiTree) LatestVersion() int64 {
	return t.latestVersion
}

var _ storetypes.CacheMultiStore = &MultiTree{}
