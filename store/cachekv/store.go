package cachekv

import (
	"io"
	"sync"

	"github.com/cosmos/cosmos-sdk/store/cachekv/internal"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Store wraps an in-memory cache around an underlying types.KVStore.
type Store struct {
	mtx    sync.Mutex
	cache  internal.BTree // always ascending sorted
	parent types.KVStore
}

var _ types.CacheKVStore = (*Store)(nil)

// NewStore creates a new Store object
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:  internal.NewBTree(),
		parent: parent,
	}
}

// GetStoreType implements Store.
func (store *Store) GetStoreType() types.StoreType {
	return store.parent.GetStoreType()
}

// Clone creates a snapshot of the cache store.
// This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (store *Store) Clone() types.CacheKVStore {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	return &Store{
		cache:  store.cache.Copy(),
		parent: store.parent,
	}
}

// swapCache swap out the internal cache store and leave the current store in a unusable state.
func (store *Store) swapCache() internal.BTree {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	cache := store.cache
	store.cache = internal.BTree{}
	return cache
}

// Restore restores the store cache to a given snapshot.
func (store *Store) Restore(s types.CacheKVStore) {
	cache := s.(*Store).swapCache()

	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.cache = cache
}

// Get implements types.KVStore.
func (store *Store) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	types.AssertValidKey(key)

	if value, found := store.cache.Get(key); found {
		return value
	}
	value = store.parent.Get(key)
	store.setCacheValue(key, value, false)
	return value
}

// Set implements types.KVStore.
func (store *Store) Set(key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	types.AssertValidKey(key)
	types.AssertValidValue(value)

	store.setCacheValue(key, value, true)
}

// Has implements types.KVStore.
func (store *Store) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

// Delete implements types.KVStore.
func (store *Store) Delete(key []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	types.AssertValidKey(key)
	store.setCacheValue(key, nil, true)
}

// Implements Cachetypes.KVStore.
func (store *Store) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	store.cache.ScanDirtyItems(func(key, value []byte) {
		if value == nil {
			store.parent.Delete(key)
		} else {
			store.parent.Set(key, value)
		}
	})

	store.cache = internal.NewBTree()
}

// CacheWrap implements CacheWrapper.
func (store *Store) CacheWrap() types.CacheWrap {
	return NewStore(store)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (store *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return NewStore(tracekv.NewStore(store, w, tc))
}

//----------------------------------------
// Iteration

// Iterator implements types.KVStore.
func (store *Store) Iterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, true)
}

// ReverseIterator implements types.KVStore.
func (store *Store) ReverseIterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, false)
}

func (store *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	isoSortedCache := store.cache.Copy()

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

//----------------------------------------
// etc

// Only entrypoint to mutate store.cache.
// A `nil` value means a deletion.
func (store *Store) setCacheValue(key, value []byte, dirty bool) {
	store.cache.Set(key, value, dirty)
}
