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

	parentTree func(storetypes.StoreKey) storetypes.CacheWrapper
}

func (t *MultiTree) GetObjKVStore(key storetypes.StoreKey) storetypes.ObjKVStore {
	store, ok := t.getCacheWrapper(key).(storetypes.ObjKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not ObjKVStore", key))
	}
	return store
}

func (t *MultiTree) getCacheWrapper(key storetypes.StoreKey) storetypes.CacheWrapper {
	var store storetypes.CacheWrapper
	treeIdx, ok := t.treesByKey[key]
	if !ok && t.parentTree != nil {
		if t.parentTree != nil {
			store = t.initStore(key, t.parentTree(key))
		} else {
			panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
		}
	} else {
		store = t.trees[treeIdx]
	}

	return store
}

func (t *MultiTree) initStore(key storetypes.StoreKey, store storetypes.CacheWrapper) storetypes.CacheWrap {
	cache := store.CacheWrap()
	t.trees = append(t.trees, cache)
	t.treesByKey[key] = len(t.trees) - 1
	return cache
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
	// TODO implement tracing
	return t.CacheWrap()
}

func (t *MultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	wrapped := &MultiTree{
		treesByKey: t.treesByKey,
		trees:      make([]storetypes.CacheWrap, len(t.trees)),
		parentTree: t.getCacheWrapper,
	}
	for i, tree := range t.trees {
		wrapped.trees[i] = tree.CacheWrap()
	}
	return wrapped
}

func (t *MultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	return nil, fmt.Errorf("CacheMultiStoreWithVersion can only be called on CommitMultiStore")
}

func (t *MultiTree) GetStore(key storetypes.StoreKey) storetypes.Store {
	store, ok := t.getCacheWrapper(key).(storetypes.Store)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not Store", key))
	}
	return store
}

func (t *MultiTree) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	store := t.getCacheWrapper(key)

	kvStore, ok := store.(storetypes.KVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not KVStore: store type=%T", key, store))
	}
	return kvStore
}

func (t *MultiTree) TracingEnabled() bool {
	return false
}

func (t *MultiTree) SetTracer(w io.Writer) storetypes.MultiStore {
	// TODO implement me
	panic("implement me")
}

func (t *MultiTree) SetTracingContext(context storetypes.TraceContext) storetypes.MultiStore {
	// TODO implement me
	panic("implement me")
}

func (t *MultiTree) LatestVersion() int64 {
	return t.latestVersion
}

var _ storetypes.MultiStore = &MultiTree{}
