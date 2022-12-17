package cachekv

import (
	"io"

	"github.com/cosmos/cosmos-sdk/store/cachekv/internal"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Store wraps an in-memory cache around an underlying types.KVStore.
// If a cached value is nil but deleted is defined for the corresponding key,
// it means the parent doesn't have the key. (No need to delete upon Write())
type Store struct {
	// always ascending sorted
	cache  internal.MemCache
	parent types.KVStore
}

var _ types.CacheKVStore = (*Store)(nil)

// NewStore creates a new Store object
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:  internal.NewMemCache(),
		parent: parent,
	}
}

// GetStoreType implements Store.
func (store *Store) GetStoreType() types.StoreType {
	return store.parent.GetStoreType()
}

// Get implements types.KVStore.
func (store *Store) Get(key []byte) []byte {
	types.AssertValidKey(key)

	if value, found := store.cache.Get(key); found {
		return value
	}

	value := store.parent.Get(key)
	store.setCacheValue(key, value, false)
	return value
}

// Set implements types.KVStore.
func (store *Store) Set(key []byte, value []byte) {
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
	types.AssertValidKey(key)
	store.setCacheValue(key, nil, true)
}

// Implements Cachetypes.KVStore.
func (store *Store) Write() {
	store.cache.ScanDirtyItems(func(key, value []byte) {
		if value == nil {
			store.parent.Delete(key)
		} else {
			store.parent.Set(key, value)
		}
	})

	store.cache = internal.NewMemCache()
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
	return internal.NewCacheMergeIterator(
		store.parent.Iterator(start, end),
		store.cache.Iterator(start, end),
		true,
	)
}

// ReverseIterator implements types.KVStore.
func (store *Store) ReverseIterator(start, end []byte) types.Iterator {
	return internal.NewCacheMergeIterator(
		store.parent.ReverseIterator(start, end),
		store.cache.ReverseIterator(start, end),
		false,
	)
}

// Only entrypoint to mutate store.cache.
func (store *Store) setCacheValue(key, value []byte, dirty bool) {
	store.cache.Set(key, value, dirty)
}
