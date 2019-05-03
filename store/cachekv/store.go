package cachekv

import (
	"bytes"
	"io"
	"sort"
	"sync"

	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/store/tracekv"
)

// If value is nil but deleted is false, it means the parent doesn't have the
// key.  (No need to delete upon Write())
type cValue struct {
	value   []byte
	deleted bool
	dirty   bool
}

// Store wraps an in-memory cache around an underlying types.KVStore.
type Store struct {
	mtx           sync.Mutex
	cache         map[string]cValue
	unsortedCache map[string]struct{}
	sortedCache   []cmn.KVPair
	parent        types.KVStore
}

var _ types.CacheKVStore = (*Store)(nil)

// nolint
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:         make(map[string]cValue),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   nil,
		parent:        parent,
	}
}

// Implements Store.
func (store *Store) GetStoreType() types.StoreType {
	return store.parent.GetStoreType()
}

// Implements types.KVStore.
func (store *Store) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

	cacheValue, ok := store.cache[string(key)]
	if !ok {
		value = store.parent.Get(key)
		store.setCacheValue(key, value, false, false)
	} else {
		value = cacheValue.value
	}

	return value
}

// Implements types.KVStore.
func (store *Store) Set(key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)
	types.AssertValidValue(value)

	store.setCacheValue(key, value, false, true)
}

// Implements types.KVStore.
func (store *Store) Has(key []byte) bool {
	value := store.Get(key)
	return value != nil
}

// Implements types.KVStore.
func (store *Store) Delete(key []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	types.AssertValidKey(key)

	store.setCacheValue(key, nil, true, true)
}

// Implements Cachetypes.KVStore.
func (store *Store) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	// We need a copy of all of the keys.
	// Not the best, but probably not a bottleneck depending.
	keys := make([]string, 0, len(store.cache))
	for key, dbValue := range store.cache {
		if dbValue.dirty {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	// TODO: Consider allowing usage of Batch, which would allow the write to
	// at least happen atomically.
	for _, key := range keys {
		cacheValue := store.cache[key]
		if cacheValue.deleted {
			store.parent.Delete([]byte(key))
		} else if cacheValue.value == nil {
			// Skip, it already doesn't exist in parent.
		} else {
			store.parent.Set([]byte(key), cacheValue.value)
		}
	}

	// Clear the cache
	store.cache = make(map[string]cValue)
	store.unsortedCache = make(map[string]struct{})
	store.sortedCache = nil
}

//----------------------------------------
// To cache-wrap this Store further.

// Implements CacheWrapper.
func (store *Store) CacheWrap() types.CacheWrap {
	return NewStore(store)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (store *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return NewStore(tracekv.NewStore(store, w, tc))
}

//----------------------------------------
// Iteration

// Implements types.KVStore.
func (store *Store) Iterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, true)
}

// Implements types.KVStore.
func (store *Store) ReverseIterator(start, end []byte) types.Iterator {
	return store.iterator(start, end, false)
}

func (store *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent, cache types.Iterator

	if ascending {
		parent = store.parent.Iterator(start, end)
	} else {
		parent = store.parent.ReverseIterator(start, end)
	}

	items := store.dirtyItems(start, end, ascending)
	cache = newMemIterator(start, end, items)

	return newCacheMergeIterator(parent, cache, ascending)
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (store *Store) dirtyItems(start, end []byte, ascending bool) []cmn.KVPair {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	unsorted := make([]cmn.KVPair, 0)

	for key := range store.unsortedCache {
		cacheValue := store.cache[key]
		if dbm.IsKeyInDomain([]byte(key), start, end) {
			unsorted = append(unsorted, cmn.KVPair{Key: []byte(key), Value: cacheValue.value})
			delete(store.unsortedCache, key)
		}
	}

	sort.Slice(unsorted, func(i, j int) bool {
		if ascending {
			return bytes.Compare(unsorted[i].Key, unsorted[j].Key) < 0
		}
		return bytes.Compare(unsorted[i].Key, unsorted[j].Key) > 0
	})

	// merge with already sortedCache
	sorted := store.sortedCache
	items := make([]cmn.KVPair, 0, len(unsorted)+len(sorted))

	for {
		if len(unsorted) == 0 {
			items = append(items, sorted...)
			break
		}
		if len(sorted) == 0 {
			items = append(items, unsorted...)
			break
		}
		uitem := unsorted[0]
		sitem := sorted[0]
		comp := bytes.Compare(uitem.Key, sitem.Key)
		if !ascending {
			comp = comp * -1
		}
		switch comp {
		case -1:
			unsorted = unsorted[1:]
			items = append(items, uitem)
		case 1:
			sorted = sorted[1:]
			items = append(items, sitem)
		case 0:
			unsorted = unsorted[1:]
			sorted = sorted[1:]
			items = append(items, uitem)
		}
	}

	store.sortedCache = items

	return items
}

//----------------------------------------
// etc

// Only entrypoint to mutate store.cache.
func (store *Store) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	store.cache[string(key)] = cValue{
		value:   value,
		deleted: deleted,
		dirty:   dirty,
	}
	if dirty {
		store.unsortedCache[string(key)] = struct{}{}
	}
}
