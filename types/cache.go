package types

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

var (
	_ KVStore = (*KVStoreCache)(nil)

	// DefaultPersistentKVStoreCacheSize defines the persistent ARC cache size for
	// each KVStore.
	DefaultPersistentKVStoreCacheSize uint = 1000
)

type (
	// KVStoreCache defines a cache that is meant to be used in a persistent
	// (inter-block) fashion and in which it wraps an underlying KVStore. Reads
	// first hit the ARC (Adaptive Replacement Cache). During a cache miss, the
	// read is delegated to the underlying KVStore and cached. Deletes and writes
	// always happen to both the cache and the KVStore in a write-through manner.
	// Caching performed in the KVStore and below is completely irrelevant to this
	// layer.
	KVStoreCache struct {
		KVStore
		cache *lru.ARCCache
	}

	// KVStoreCacheManager defines a manager that stores a mapping of StoreKeys
	// to KVStoreCache references. Each KVStoreCache reference is meant to be
	// persistent between blocks.
	KVStoreCacheManager struct {
		cacheSize uint
		caches    map[string]*KVStoreCache
	}
)

func NewKVStoreCache(store KVStore, size uint) *KVStoreCache {
	cache, err := lru.NewARC(int(size))
	if err != nil {
		panic(fmt.Errorf("failed to create KVStore cache: %s", err))
	}

	return &KVStoreCache{
		KVStore: store,
		cache:   cache,
	}
}

func NewKVStoreCacheManager(size uint) *KVStoreCacheManager {
	return &KVStoreCacheManager{
		cacheSize: size,
		caches:    make(map[string]*KVStoreCache),
	}
}

// GetKVStoreCache returns a KVStore from the KVStoreCacheManager which is
// deceratored with a persistent inter-block cache particular for that store's
// StoreKey. If the KVStore does not exist in the KVStoreCacheManager, it is
// added. Otherwise, the KVStore is updated to the provided parameter.
func (cmgr *KVStoreCacheManager) GetKVStoreCache(key StoreKey, store KVStore) KVStore {
	if cmgr.caches[key.Name()] == nil {
		cmgr.caches[key.Name()] = NewKVStoreCache(store, cmgr.cacheSize)
	} else {
		cmgr.caches[key.Name()].KVStore = store
	}

	return cmgr.caches[key.Name()]
}

// Reset resets in the internal caches.
func (cmgr *KVStoreCacheManager) Reset() {
	cmgr.caches = make(map[string]*KVStoreCache)
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
