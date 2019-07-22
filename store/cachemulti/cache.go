package cachemulti

import (
	"github.com/cosmos/cosmos-sdk/store/types"

	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

var (
	_ types.KVStore   = (*StoreCache)(nil)
	_ types.CacheWrap = (*StoreCache)(nil)

	// The interBlockCache contains an inter-block write-through cache for each
	// KVStore. Each underlying KVStore itself may be cache-wrapped, but the
	// cache here is persistent through block production.
	interBlockCache = NewStoreCacheManager()

	// PersistentStoreCacheSize defines the persistent ARC cache size for each
	// KVStore.
	PersistentStoreCacheSize = 1000
)

type (
	// StoreCache defines a cache that is meant to be used in a persistent
	// (inter-block) fashion and which wraps an underlying KVStore. Reads first
	// hit the ARC (Adaptive Replacement Cache). During a cache miss, the read
	// is delegated to the underlying KVStore. Deletes and writes always happen
	// to both the cache and the KVStore in a write-through manner. Caching
	// performed in the KVStore is completely irrelevant to this layer.
	StoreCache struct {
		types.KVStore
		cache *lru.ARCCache
	}

	// StoreCacheManager defines a manager that handles a mapping of StoreKeys
	// to StoreCache references. Each StoreCache reference is meant to be
	// persistent between blocks.
	StoreCacheManager struct {
		caches map[types.StoreKey]*StoreCache
	}
)

func NewStoreCache(store types.KVStore) *StoreCache {
	cache, err := lru.NewARC(PersistentStoreCacheSize)
	if err != nil {
		panic(fmt.Errorf("failed to create cache: %s", err))
	}

	return &StoreCache{
		KVStore: store,
		cache:   cache,
	}
}

func NewStoreCacheManager() *StoreCacheManager {
	return &StoreCacheManager{
		caches: make(map[types.StoreKey]*StoreCache),
	}
}

// GetOrSetStoreCache attempts to get a CacheWrap store from the StoreCacheManager.
// If the CacheWrap does not exist in the mapping, it is added. Each CacheWrap
// store contains a persistent cache through the StoreCache.
func (cmgr *StoreCacheManager) GetOrSetStoreCache(key types.StoreKey, store types.CacheWrap) types.CacheWrap {
	if cmgr.caches[key] == nil {
		cmgr.caches[key] = NewStoreCache(store.(types.KVStore))
	}

	return cmgr.caches[key]
}

// Write implements the CacheWrap interface and simply delegates the Write call
// to the underlying KVStore.
func (sc *StoreCache) Write() {
	sc.KVStore.(types.CacheWrap).Write()
}

// Get retrieves a value by key. It will first look in the write-through cache.
// If the value doesn't exist in the write-through cache, the Get call is
// delegated to the underlying KVStore.
func (sc *StoreCache) Get(key []byte) []byte {
	val, ok := sc.cache.Get(string(key))
	if ok {
		// cache hit
		return val.([]byte)
	}

	// cache miss; add to cache
	bz := sc.KVStore.Get(key)
	sc.cache.Add(string(key), bz)

	return bz
}

// Set inserts a key/value pair into both the write-through cache and the
// underlying KVStore.
func (sc *StoreCache) Set(key, value []byte) {
	sc.cache.Add(string(key), value)
	sc.KVStore.Set(key, value)
}

// Delete removes a key/value pair from both the write-through cache and the
// underlying KVStore.
func (sc *StoreCache) Delete(key []byte) {
	sc.cache.Remove(string(key))
	sc.KVStore.Delete(key)
}
