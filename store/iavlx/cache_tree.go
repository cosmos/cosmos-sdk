package iavlx

import (
	io "io"
	"sync"

	"github.com/tidwall/btree"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/internal/conv"
)

type CacheTree struct {
	mtx    sync.Mutex // TODO do we really need a mutex or could this be part of the caller contract?
	parent storetypes.KVStore
	dirty  bool
	cache  btree.Map[string, cacheEntry]
}

func NewCacheTree(parent storetypes.KVStore) *CacheTree {
	return &CacheTree{parent: parent}
}

type cacheEntry struct {
	value []byte
	dirty bool
}

func (store *CacheTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (store *CacheTree) CacheWrap() storetypes.CacheWrap {
	return NewCacheTree(store)
}

func (store *CacheTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement tracing
	return NewCacheTree(store)
}

func (store *CacheTree) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	storetypes.AssertValidKey(key)

	keyStr := conv.UnsafeBytesToStr(key)
	cacheValue, ok := store.cache.Get(keyStr)
	if !ok {
		value = store.parent.Get(key)
		store.setCacheValue(keyStr, value, false)
	} else {
		value = cacheValue.value
	}
	return value
}

func (store *CacheTree) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

func (store *CacheTree) Set(key, value []byte) {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)

	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.setCacheValue(conv.UnsafeBytesToStr(key), value, true)
}

func (store *CacheTree) Delete(key []byte) {
	storetypes.AssertValidKey(key)

	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.setCacheValue(conv.UnsafeBytesToStr(key), nil, true)
}

func (store *CacheTree) Iterator(start, end []byte) storetypes.Iterator {
	//TODO implement me
	panic("implement me")
}

func (store *CacheTree) ReverseIterator(start, end []byte) storetypes.Iterator {
	//TODO implement me
	panic("implement me")
}

func (store *CacheTree) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	if !store.dirty {
		return
	}

	// TODO if we are concerned about retaining the whole tree in memory, we could maybe drain the cache using Map.PopMin
	store.cache.Scan(func(key string, value cacheEntry) bool {
		if !value.dirty {
			// TODO we could save these cached reads in the tree but for now we just clear the whole cache
			return true
		}

		// We use []byte(key) instead of conv.UnsafeStrToBytes because we cannot
		// be sure if the underlying store might do a save with the byteslice or
		// not. Once we get confirmation that .Delete is guaranteed not to
		// save the byteslice, then we can assume only a read-only copy is sufficient.
		if value.value == nil {
			store.parent.Delete([]byte(key))
		} else {
			store.parent.Set([]byte(key), value.value)
		}
		return true
	})

	store.cache.Clear()
	store.dirty = false
}

func (store *CacheTree) setCacheValue(key string, value []byte, dirty bool) {
	if dirty {
		store.dirty = true
	}
	store.cache.Set(key, cacheEntry{
		value: value,
		dirty: dirty,
	})
}

func (store *CacheTree) iterator(start, end []byte, ascending bool) types.Iterator {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.dirtyItems(start, end)
	isoSortedCache := store.sortedCache.Copy()

	var (
		err           error
		parent, cache types.Iterator
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

	return internal.NewCacheMergeIterator(parent, cache, ascending)
}

var _ storetypes.CacheKVStore = (*CacheTree)(nil)
