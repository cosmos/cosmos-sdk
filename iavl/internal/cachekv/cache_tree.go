package cachekv

import (
	io "io"
	"iter"
	"sync"

	storetypes "cosmossdk.io/store/types"
)

type Store struct {
	mtx    sync.Mutex // TODO do we really need a mutex or could this be part of the caller contract?
	parent storetypes.KVStore
	dirty  bool
	cache  BTree
}

func NewStore(parent storetypes.KVStore) *Store {
	return &Store{parent: parent, cache: NewBTree()}
}

func (store *Store) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (store *Store) CacheWrap() storetypes.CacheWrap {
	return NewStore(store)
}

func (store *Store) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement tracing
	return NewStore(store)
}

func (store *Store) Get(key []byte) (value []byte) {
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

func (store *Store) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

func (store *Store) Set(key, value []byte) {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)

	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.cache.Set(key, value)
	store.dirty = true
}

func (store *Store) Delete(key []byte) {
	storetypes.AssertValidKey(key)

	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.cache.Delete(key)
	store.dirty = true
}

func (store *Store) Iterator(start, end []byte) storetypes.Iterator {
	return store.iterator(start, end, true)
}

func (store *Store) ReverseIterator(start, end []byte) storetypes.Iterator {
	return store.iterator(start, end, false)
}

type KVUpdate = struct {
	Key, Value []byte
	Delete     bool
}

func (store *Store) Updates() iter.Seq[KVUpdate] {
	return func(yield func(KVUpdate) bool) {
		store.mtx.Lock()
		defer store.mtx.Unlock()

		store.cache.Scan(func(key, value []byte, dirty bool) bool {
			if !dirty {
				return true
			}

			update := KVUpdate{
				Key:    key,
				Value:  value,
				Delete: value == nil,
			}
			return yield(update)
		})
	}
}

func (store *Store) Write() {
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

func (store *Store) iterator(start, end []byte, ascending bool) storetypes.Iterator {
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
