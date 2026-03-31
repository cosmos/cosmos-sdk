package internal

import (
	"fmt"
	"sync"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// MultiTree is an in-memory cached view of multiple stores for a given version.
// It implements storetypes.CacheMultiStore with lazy initialization: individual store
// caches are created on first access via initCacheWrapFromParent.
type MultiTree struct {
	// trees is a sync.Map for block STM compatibility at the root store layer.
	trees                   sync.Map
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

func (t *MultiTree) LatestVersion() int64 {
	return t.latestVersion
}

var _ storetypes.CacheMultiStore = &MultiTree{}
