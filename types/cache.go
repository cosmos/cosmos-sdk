package types

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

var (
	_ KVStore = (*KVStoreCache)(nil)

	// PersistentStoreCacheSize defines the persistent ARC cache size for each
	// KVStore.
	PersistentStoreCacheSize = 1000
)

type (
	// KVStoreCache defines a cache that is meant to be used in a persistent
	// (inter-block) fashion and in which it wraps an underlying KVStore. Reads
	// first hit the ARC (Adaptive Replacement Cache). During a cache miss, the
	// read is delegated to the underlying KVStore. Deletes and writes always
	// happen to both the cache and the KVStore in a write-through manner. Caching
	// performed in the KVStore and below is completely irrelevant to this layer.
	KVStoreCache struct {
		KVStore
		cache *lru.ARCCache
	}

	// StoreCacheManager defines a manager that stores a mapping of StoreKeys
	// to KVStoreCache references. Each KVStoreCache reference is meant to be
	// persistent between blocks.
	StoreCacheManager struct {
		caches map[StoreKey]*KVStoreCache
	}
)

func NewStoreCache(store KVStore) *KVStoreCache {
	cache, err := lru.NewARC(PersistentStoreCacheSize)
	if err != nil {
		panic(fmt.Errorf("failed to create cache: %s", err))
	}

	return &KVStoreCache{
		KVStore: store,
		cache:   cache,
	}
}

func NewStoreCacheManager() *StoreCacheManager {
	return &StoreCacheManager{
		caches: make(map[StoreKey]*KVStoreCache),
	}
}

// GetOrSetKVStoreCache attempts to get a CacheWrap store from the StoreCacheManager.
// If the CacheWrap does not exist in the mapping, it is added. Each CacheWrap
// store contains a persistent cache through the StoreCache.
func (cmgr *StoreCacheManager) GetOrSetKVStoreCache(key StoreKey, store KVStore) KVStore {
	if cmgr.caches[key] == nil {
		cmgr.caches[key] = NewStoreCache(store)
	}

	return cmgr.caches[key]
}

// Get retrieves a value by key. It will first look in the write-through cache.
// If the value doesn't exist in the write-through cache, the Get call is
// delegated to the underlying KVStore.
func (kvsc *KVStoreCache) Get(key []byte) []byte {
	val, ok := kvsc.cache.Get(string(key))
	if ok {
		// cache hit
		return val.([]byte)
	}

	// cache miss; add to cache
	bz := kvsc.KVStore.Get(key)
	kvsc.cache.Add(string(key), bz)

	return bz
}

// Set inserts a key/value pair into both the write-through cache and the
// underlying KVStore.
func (kvsc *KVStoreCache) Set(key, value []byte) {
	kvsc.cache.Add(string(key), value)
	kvsc.KVStore.Set(key, value)
}

// Delete removes a key/value pair from both the write-through cache and the
// underlying KVStore.
func (kvsc *KVStoreCache) Delete(key []byte) {
	kvsc.cache.Remove(string(key))
	kvsc.KVStore.Delete(key)
}
