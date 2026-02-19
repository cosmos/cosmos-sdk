package internal

import (
	"fmt"
	io "io"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

type MultiTree struct {
	// for now we use sync.Map but this is only needed by block STM for the root store layer - we should find a fix that doesn't force sync.Map on every layer
	trees                   sync.Map // index of the trees by name
	initCacheWrapFromParent func(storetypes.StoreKey) storetypes.CacheWrap
	latestVersion           int64
}

// NewMultiTree creates a new MultiTree from the given CommitInfo, which contains the version and store info for each store.
// CommitInfo will be nil if this is the first version of the MultiTree, in which case the version will be set to 0 and the store info will be empty.
func NewMultiTree(version int64, initCacheWrapFromParent func(key storetypes.StoreKey) storetypes.CacheWrap) *MultiTree {
	return &MultiTree{
		latestVersion:           version,
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
	store, ok := t.trees.Load(key)
	if ok {
		return store.(storetypes.CacheWrap)
	}
	return nil
}

func (t *MultiTree) getCacheWrap(key storetypes.StoreKey) storetypes.CacheWrap {
	store, ok := t.trees.Load(key)
	if ok {
		return store.(storetypes.CacheWrap)
	}
	newStore := t.initCacheWrapFromParent(key)
	t.trees.Store(key, newStore)
	return newStore
}

func (t *MultiTree) Write() {
	var wg sync.WaitGroup
	t.trees.Range(func(key, value any) bool {
		wg.Add(1)
		go func(t storetypes.CacheWrap) {
			defer wg.Done()
			t.Write()
		}(value.(storetypes.CacheWrap))
		return true
	})
	wg.Wait()
}

func (t *MultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (t *MultiTree) CacheWrap() storetypes.CacheWrap {
	return t.CacheMultiStore()
}

func (t *MultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement me
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
