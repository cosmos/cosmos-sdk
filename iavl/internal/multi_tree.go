package internal

import (
	"fmt"
	io "io"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

type MultiTree struct {
	latestVersion           int64
	trees                   map[storetypes.StoreKey]storetypes.CacheWrap // index of the trees by name
	initCacheWrapFromParent func(storetypes.StoreKey) storetypes.CacheWrap
}

func NewMultiTree(version int64, initCacheWrapFromParent func(key storetypes.StoreKey) storetypes.CacheWrap) *MultiTree {
	return &MultiTree{
		latestVersion:           version,
		trees:                   map[storetypes.StoreKey]storetypes.CacheWrap{},
		initCacheWrapFromParent: initCacheWrapFromParent,
	}
}

func (t *MultiTree) GetObjKVStore(key storetypes.StoreKey) storetypes.ObjKVStore {
	store, ok := t.getCacheWrap(key).(storetypes.ObjKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not ObjKVStore", key))
	}
	return store
}

func (t *MultiTree) GetCacheWrapIfExists(key storetypes.StoreKey) storetypes.CacheWrap {
	store, ok := t.trees[key]
	if ok {
		return store
	}
	return nil
}

func (t *MultiTree) getCacheWrap(key storetypes.StoreKey) storetypes.CacheWrap {
	store, ok := t.trees[key]
	if ok {
		return store
	}
	store = t.initCacheWrapFromParent(key)
	t.trees[key] = store
	return store
}

func (t *MultiTree) Write() {
	var wg sync.WaitGroup
	for _, tree := range t.trees {
		// TODO check if trees are dirty before spinning off a goroutine
		wg.Add(1)
		go func(t storetypes.CacheWrap) {
			defer wg.Done()
			t.Write()
		}(tree)
	}
	wg.Wait()
}

func (t *MultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (t *MultiTree) CacheWrap() storetypes.CacheWrap {
	return t.CacheMultiStore()
}

func (t *MultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	logger.Warn("CacheWrapWithTrace called on MultiTree: tracing not implemented")
	return t.CacheWrap()
}

func (t *MultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	// create a nested MultiTree, which in turn creates CacheWraps for each store
	return NewMultiTree(t.latestVersion, func(key storetypes.StoreKey) storetypes.CacheWrap {
		return t.getCacheWrap(key).CacheWrap()
	})
}

func (t *MultiTree) CacheMultiStoreWithVersion(int64) (storetypes.CacheMultiStore, error) {
	return nil, fmt.Errorf("CacheMultiStoreWithVersion can only be called on CommitMultiStore")
}

func (t *MultiTree) GetStore(key storetypes.StoreKey) storetypes.Store {
	store, ok := t.getCacheWrap(key).(storetypes.Store)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not Store", key))
	}
	return store
}

func (t *MultiTree) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	store := t.getCacheWrap(key)

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
