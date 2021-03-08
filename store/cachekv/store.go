package cachekv

import (
	"bytes"
	"io"
	"reflect"
	"sort"
	"sync"
	"time"
	"unsafe"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types/kv"
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
	cache         map[string]*cValue
	unsortedCache map[string]struct{}
	sortedCache   *kv.List // always ascending sorted
	parent        types.KVStore
}

var _ types.CacheKVStore = (*Store)(nil)

// NewStore creates a new Store object
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:         make(map[string]*cValue),
		unsortedCache: make(map[string]struct{}),
		sortedCache:   kv.NewList(),
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
	defer telemetry.MeasureSince(time.Now(), "store", "cachekv", "get")

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

// Set implements types.KVStore.
func (store *Store) Set(key []byte, value []byte) {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	defer telemetry.MeasureSince(time.Now(), "store", "cachekv", "set")

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
	defer telemetry.MeasureSince(time.Now(), "store", "cachekv", "delete")

	types.AssertValidKey(key)
	store.setCacheValue(key, nil, true, true)
}

// Implements Cachetypes.KVStore.
func (store *Store) Write() {
	store.mtx.Lock()
	defer store.mtx.Unlock()
	defer telemetry.MeasureSince(time.Now(), "store", "cachekv", "write")

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

		switch {
		case cacheValue.deleted:
			store.parent.Delete([]byte(key))
		case cacheValue.value == nil:
			// Skip, it already doesn't exist in parent.
		default:
			store.parent.Set([]byte(key), cacheValue.value)
		}
	}

	// Clear the cache
	store.cache = make(map[string]*cValue)
	store.unsortedCache = make(map[string]struct{})
	store.sortedCache = kv.NewList()
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

	var parent, cache types.Iterator

	if ascending {
		parent = store.parent.Iterator(start, end)
	} else {
		parent = store.parent.ReverseIterator(start, end)
	}

	store.dirtyItems(start, end)
	cache = newMemIterator(start, end, store.sortedCache, ascending)

	return newCacheMergeIterator(parent, cache, ascending)
}

// strToByte is meant to make a zero allocation conversion
// from string -> []byte to speed up operations, it is not meant
// to be used generally, but for a specific pattern to check for available
// keys within a domain.
func strToByte(s string) []byte {
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Cap = len(s)
	hdr.Len = len(s)
	hdr.Data = (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	return b
}

// byteSliceToStr is meant to make a zero allocation conversion
// from []byte -> string to speed up operations, it is not meant
// to be used generally, but for a specific pattern to delete keys
// from a map.
func byteSliceToStr(b []byte) string {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&b))
	return *(*string)(unsafe.Pointer(hdr))
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (store *Store) dirtyItems(start, end []byte) {
	unsorted := make([]*kv.Pair, 0)

	n := len(store.unsortedCache)
	for key := range store.unsortedCache {
		if dbm.IsKeyInDomain(strToByte(key), start, end) {
			cacheValue := store.cache[key]
			unsorted = append(unsorted, &kv.Pair{Key: []byte(key), Value: cacheValue.value})
		}
	}

	if len(unsorted) == n { // This pattern allows the Go compiler to emit the map clearing idiom for the entire map.
		for key := range store.unsortedCache {
			delete(store.unsortedCache, key)
		}
	} else { // Otherwise, normally delete the unsorted keys from the map.
		for _, kv := range unsorted {
			delete(store.unsortedCache, byteSliceToStr(kv.Key))
		}
	}

	sort.Slice(unsorted, func(i, j int) bool {
		return bytes.Compare(unsorted[i].Key, unsorted[j].Key) < 0
	})

	for e := store.sortedCache.Front(); e != nil && len(unsorted) != 0; {
		uitem := unsorted[0]
		sitem := e.Value
		comp := bytes.Compare(uitem.Key, sitem.Key)

		switch comp {
		case -1:
			unsorted = unsorted[1:]

			store.sortedCache.InsertBefore(uitem, e)
		case 1:
			e = e.Next()
		case 0:
			unsorted = unsorted[1:]
			e.Value = uitem
			e = e.Next()
		}
	}

	for _, kvp := range unsorted {
		store.sortedCache.PushBack(kvp)
	}
}

//----------------------------------------
// etc

// Only entrypoint to mutate store.cache.
func (store *Store) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	store.cache[string(key)] = &cValue{
		value:   value,
		deleted: deleted,
		dirty:   dirty,
	}
	if dirty {
		store.unsortedCache[string(key)] = struct{}{}
	}
}
