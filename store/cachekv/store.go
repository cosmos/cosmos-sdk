package cachekv

import (
	"bytes"
	"io"
	"sort"
	"sync"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// If value is nil but deleted is false, it means the parent doesn't have the
// key.  (No need to delete upon Write())
type cValue struct {
	value []byte
	dirty bool
}

// Store wraps an in-memory cache around an underlying types.KVStore.
type Store struct {
	mtx           sync.Mutex
	cache         map[string]*cValue
	deleted       map[string]struct{}
	unsortedCache map[string]struct{}
	sortedCache   *dbm.MemDB // always ascending sorted
	parent        types.KVStore

	activeIterators int
}

var _ types.CacheKVStore = (*Store)(nil)

// NewStore creates a new Store object
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:         make(map[string]*cValue),
		deleted:       make(map[string]struct{}),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   dbm.NewMemDB(),
		parent:        parent,
	}
}

// GetStoreType implements Store.
func (store *Store) GetStoreType() types.StoreType {
	return store.parent.GetStoreType()
}

// Get implements types.KVStore.
func (store *Store) Get(key []byte) (value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	types.AssertValidKey(key)

	cacheValue, ok := store.cache[conv.UnsafeBytesToStr(key)]
	if !ok {
		value = store.parent.Get(key)
		store.setCacheValue(key, value, false, false)
	} else {
		value = cacheValue.value
	}

	return value
}

// Set implements types.KVStore.
func (store *Store) Set(key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	types.AssertValidKey(key)
	types.AssertValidValue(value)

	store.setCacheValue(key, value, false, true)
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
		if store.isDeleted(key) {
			// We use []byte(key) instead of conv.UnsafeStrToBytes because we cannot
			// be sure if the underlying store might do a save with the byteslice or
			// not. Once we get confirmation that .Delete is guaranteed not to
			// save the byteslice, then we can assume only a read-only copy is sufficient.
			store.parent.Delete([]byte(key))
			continue
		}

		cacheValue := store.cache[key]
		if cacheValue.value != nil {
			// It already exists in the parent, hence delete it.
			store.parent.Set([]byte(key), cacheValue.value)
		}
	}

	// Clear the cache using the map clearing idiom
	// and not allocating fresh objects.
	// Please see https://bencher.orijtech.com/perfclinic/mapclearing/
	for key := range store.cache {
		delete(store.cache, key)
	}
	for key := range store.deleted {
		delete(store.deleted, key)
	}
	for key := range store.unsortedCache {
		delete(store.unsortedCache, key)
	}
	store.sortedCache = dbm.NewMemDB()
}

// CacheWrap implements CacheWrapper.
func (store *Store) CacheWrap() types.CacheWrap {
	return NewStore(store)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (store *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return NewStore(tracekv.NewStore(store, w, tc))
}

// CacheWrapWithListeners implements the CacheWrapper interface.
func (store *Store) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return NewStore(listenkv.NewStore(store, storeKey, listeners))
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
	store.activeIterators += 1

	var parent, cache types.Iterator

	if ascending {
		parent = store.parent.Iterator(start, end)
	} else {
		parent = store.parent.ReverseIterator(start, end)
	}

	store.dirtyItems(start, end)
	cache = newMemIterator(start, end, store.sortedCache, store.deleted, ascending)

	return newCacheMergeIterator(store, parent, cache, ascending)
}

func (store *Store) decrementActiveIteratorCount() {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	store.activeIterators -= 1
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (store *Store) dirtyItems(start, end []byte) {
	startStr, endStr := conv.UnsafeBytesToStr(start), conv.UnsafeBytesToStr(end)
	if startStr > endStr {
		// Nothing to do here.
		return
	}

	// dirty items needs to clear the unsorted items from the unsortedCache, then put them into the sortedCache.
	// Subsequently an iterator is constructed using the sorted cache and parent store.
	//
	// However:
	// if we have multiple iterators open, we have to avoid writes to the store.sortedCache MemDB.
	// This is because a write to mem-db will definitely cause deadlocks in tm-db's memdb abstractions.
	// (Though it could potentially be thread-safe per golang btree's docs)
	// Thus we can only safely clear the unsorted cache to construct an iterator if we are:
	// - only doing one iterator at a time
	// - doing multiple iterators, over the same data, and you didn't write to the subset between
	//   first iterator creation and current iterator creation
	// - doing multiple iterators over disjoint pieces of data,
	//   and have done no writes over any iterated component since first iterator creation
	//
	// To reword, in order to better accomodate multiple iterators,
	// every time we construct an iterator when no others are active, we clear the entire unsorted cache.
	// otherwise we don't do any clearing of the unsorted cache,
	// and panic if we would be constructing an invalid iterator.
	// TODO: Come back and vastly simplify the cacheKV store / iterator architecture,
	// so we don't have these messy issues.

	numUnsortedEntries := len(store.unsortedCache)
	if store.activeIterators > 1 {
		numUnsortedEntries = 0
	}
	unsorted := make([]*kv.Pair, 0, numUnsortedEntries)
	for key := range store.unsortedCache {
		cacheValue := store.cache[key]
		keyBz := conv.UnsafeStrToBytes(key)
		if store.activeIterators == 1 {
			// if we have one iterator, build a list of every KV pair thats not sorted.
			unsorted = append(unsorted, &kv.Pair{Key: keyBz, Value: cacheValue.value})
		} else if store.activeIterators > 1 {
			// if we have more than one iterator we can't clear unsorted.
			// therefore we check here if any of the unsorted values are in our iterator range, and if so panic.
			if dbm.IsKeyInDomain(conv.UnsafeStrToBytes(key), start, end) {
				panic("Invalid concurrent iterator construction!" +
					" If you are using multiple iterators concurrently, there must be no writes after the first" + "concurrent iterator's creations over data ranges you will iterate over." +
					" Writes over these data ranges can resume once all concurrent iterators have been created," + " or the iterators have been closed.")
			}
		}
	}

	if store.activeIterators == 1 {
		store.clearEntireUnsortedCache()
		store.updateSortedCache(unsorted)
	}
}

func (store *Store) clearEntireUnsortedCache() {
	// This pattern allows the Go compiler to emit the map clearing idiom for the entire map.
	for key := range store.unsortedCache {
		delete(store.unsortedCache, key)
	}
}

func (store *Store) updateSortedCache(unsorted []*kv.Pair) {
	sort.Slice(unsorted, func(i, j int) bool {
		return bytes.Compare(unsorted[i].Key, unsorted[j].Key) < 0
	})

	for _, item := range unsorted {
		if item.Value == nil {
			// deleted element, tracked by store.deleted
			// setting arbitrary value
			store.sortedCache.Set(item.Key, []byte{})
			continue
		}
		err := store.sortedCache.Set(item.Key, item.Value)
		if err != nil {
			panic(err)
		}
	}
}

//----------------------------------------
// etc

// Only entrypoint to mutate store.cache.
func (store *Store) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	keyStr := conv.UnsafeBytesToStr(key)
	store.cache[keyStr] = &cValue{
		value: value,
		dirty: dirty,
	}
	if deleted {
		store.deleted[keyStr] = struct{}{}
	} else if dirty {
		delete(store.deleted, keyStr)
	}
	if dirty {
		store.unsortedCache[keyStr] = struct{}{}
	}
}

func (store *Store) isDeleted(key string) bool {
	_, ok := store.deleted[key]
	return ok
}
