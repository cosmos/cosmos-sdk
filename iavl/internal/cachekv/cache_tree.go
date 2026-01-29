package cachekv

import (
	io "io"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

type CacheKVStore struct {
	mtx    sync.Mutex // TODO do we really need a mutex or could this be part of the caller contract?
	parent storetypes.KVStore
	dirty  bool
	cache  BTree
}

func NewCacheKVStore(parent storetypes.KVStore) *CacheKVStore {
	return &CacheKVStore{parent: parent, cache: NewBTree()}
}

func (store *CacheKVStore) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (store *CacheKVStore) CacheWrap() storetypes.CacheWrap {
	return NewCacheKVStore(store)
}

func (store *CacheKVStore) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement tracing
	return NewCacheKVStore(store)
}

func (store *CacheKVStore) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	storetypes.AssertValidKey(key)

	var ok bool
	value, ok = store.cache.Get(key)
	if !ok {
		value = store.parent.Get(key)
		store.cache.SetCached(key, value)
	}
	return value
}

func (store *CacheKVStore) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

func (store *CacheKVStore) Set(key, value []byte) {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)

	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.cache.Set(key, value)
	store.dirty = true
}

func (store *CacheKVStore) Delete(key []byte) {
	storetypes.AssertValidKey(key)

	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.cache.Delete(key)
	store.dirty = true
}

func (store *CacheKVStore) Iterator(start, end []byte) storetypes.Iterator {
	return store.iterator(start, end, true)
}

func (store *CacheKVStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	return store.iterator(start, end, false)
}

func (store *CacheKVStore) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	if !store.dirty {
		return
	}

	// TODO if we are concerned about retaining the whole tree in memory, we could maybe drain the cache using Map.PopMin
	store.cache.Scan(func(key, value []byte, dirty bool) bool {
		if !dirty {
			// TODO we could save these cached reads in the tree but for now we just clear the whole cache
			return true
		}

		// We use []byte(key) instead of conv.UnsafeStrToBytes because we cannot
		// be sure if the underlying store might do a save with the byteslice or
		// not. Once we get confirmation that .Delete is guaranteed not to
		// save the byteslice, then we can assume only a read-only copy is sufficient.
		if value == nil {
			store.parent.Delete(key)
		} else {
			store.parent.Set(key, value)
		}
		return true
	})

	store.cache.Clear()
	store.dirty = false
}

func (store *CacheKVStore) iterator(start, end []byte, ascending bool) storetypes.Iterator {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	isoSortedCache := store.cache.Copy()

	var (
		err           error
		parent, cache storetypes.Iterator
	)

	if ascending {
		parent = store.parent.Iterator(start, end)
		cache, err = isoSortedCache.Iterator(start, end)
	} else {
		parent = store.parent.ReverseIterator(start, end)
		cache, err = isoSortedCache.ReverseIterator(start, end)
	}
	if err != nil {
		panic(err)
	}

	return NewCacheMergeIterator(parent, cache, ascending)
}
