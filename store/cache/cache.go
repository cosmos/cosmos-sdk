package cache

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/types"
)

var (
	_ types.MultiStorePersistentCache = (*KVStoreCacheManager)(nil)

	// DefaultCommitKVStoreCacheSize defines the persistent ARC cache size for a
	// KVStoreCache.
	DefaultCommitKVStoreCacheSize uint = 1000
)

type (
	// KVStoreCache implements an inter-block (persistent) cache that wraps a
	// KVStore. Reads first hit the internal ARC (Adaptive Replacement Cache).
	// During a cache miss, the read is delegated to the underlying KVStore
	// and cached. Deletes and writes always happen to both the cache and the
	// KVStore in a write-through manner. Caching performed in the
	// KVStore and below is completely irrelevant to this layer.
	KVStoreCache struct {
		types.KVStore
		cache *lru.ARCCache
	}

	// KVStoreCacheManager maintains a mapping from a StoreKey to a
	// KVStoreCache. Each KVStore, per StoreKey, is meant to be used
	// in an inter-block (persistent) manner and typically provided by a
	// CommitMultiStore.
	KVStoreCacheManager struct {
		cacheSize uint
		caches    map[string]types.KVStore
	}
)

func NewKVStoreCache(store types.KVStore, size uint) *KVStoreCache {
	cache, err := lru.NewARC(int(size))
	if err != nil {
		panic(fmt.Errorf("failed to create KVStore cache: %w", err))
	}

	return &KVStoreCache{
		KVStore: store,
		cache:   cache,
	}
}

func NewKVStoreCacheManager(size uint) *KVStoreCacheManager {
	return &KVStoreCacheManager{
		cacheSize: size,
		caches:    make(map[string]types.KVStore),
	}
}

// GetStoreCache returns a Cache from the KVStoreCacheManager for a given
// StoreKey. If no Cache exists for the StoreKey, then one is created and set.
// The returned Cache is meant to be used in a persistent manner.
func (cmgr *KVStoreCacheManager) GetStoreCache(key types.StoreKey, store types.KVStore) types.KVStore {
	if cmgr.caches[key.Name()] == nil {
		cmgr.caches[key.Name()] = NewKVStoreCache(store, cmgr.cacheSize)
	}

	return cmgr.caches[key.Name()]
}

// Reset resets in the internal caches.
func (cmgr *KVStoreCacheManager) Reset() {
	// Clear the map.
	// Please note that we are purposefully using the map clearing idiom.
	// See https://github.com/cosmos/cosmos-sdk/issues/6681.
	for key := range cmgr.caches {
		delete(cmgr.caches, key)
	}
}

// CacheWrap implements the CacheWrapper interface
func (ckv *KVStoreCache) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(ckv)
}

// Get retrieves a value by key. It will first look in the write-through cache.
// If the value doesn't exist in the write-through cache, the query is delegated
// to the underlying KVStore.
func (ckv *KVStoreCache) Get(key []byte) []byte {
	types.AssertValidKey(key)

	keyStr := string(key)
	valueI, ok := ckv.cache.Get(keyStr)
	if ok {
		// cache hit
		return valueI.([]byte)
	}

	// cache miss; write to cache
	value := ckv.KVStore.Get(key)
	ckv.cache.Add(keyStr, value)

	return value
}

// Set inserts a key/value pair into both the write-through cache and the
// underlying KVStore.
func (ckv *KVStoreCache) Set(key, value []byte) {
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	ckv.cache.Add(string(key), value)
	ckv.KVStore.Set(key, value)
}

// Delete removes a key/value pair from both the write-through cache and the
// underlying KVStore.
func (ckv *KVStoreCache) Delete(key []byte) {
	ckv.cache.Remove(string(key))
	ckv.KVStore.Delete(key)
}
