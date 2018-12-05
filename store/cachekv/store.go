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
	mtx    sync.Mutex
	cache  map[string]cValue
	parent types.KVStore
}

var _ types.CacheKVStore = (*Store)(nil)

// nolint
func NewStore(parent types.KVStore) *Store {
	return &Store{
		cache:  make(map[string]cValue),
		parent: parent,
	}
}

// Implements Store.
func (ci *Store) GetStoreType() types.StoreType {
	return ci.parent.GetStoreType()
}

// Implements types.KVStore.
func (ci *Store) Get(key []byte) (value []byte) {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()
	ci.assertValidKey(key)

	cacheValue, ok := ci.cache[string(key)]
	if !ok {
		value = ci.parent.Get(key)
		ci.setCacheValue(key, value, false, false)
	} else {
		value = cacheValue.value
	}

	return value
}

// Implements types.KVStore.
func (ci *Store) Set(key []byte, value []byte) {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()
	ci.assertValidKey(key)
	ci.assertValidValue(value)

	ci.setCacheValue(key, value, false, true)
}

// Implements types.KVStore.
func (ci *Store) Has(key []byte) bool {
	value := ci.Get(key)
	return value != nil
}

// Implements types.KVStore.
func (ci *Store) Delete(key []byte) {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()
	ci.assertValidKey(key)

	ci.setCacheValue(key, nil, true, true)
}

// Implements Cachetypes.KVStore.
func (ci *Store) Write() {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()

	// We need a copy of all of the keys.
	// Not the best, but probably not a bottleneck depending.
	keys := make([]string, 0, len(ci.cache))
	for key, dbValue := range ci.cache {
		if dbValue.dirty {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	// TODO: Consider allowing usage of Batch, which would allow the write to
	// at least happen atomically.
	for _, key := range keys {
		cacheValue := ci.cache[key]
		if cacheValue.deleted {
			ci.parent.Delete([]byte(key))
		} else if cacheValue.value == nil {
			// Skip, it already doesn't exist in parent.
		} else {
			ci.parent.Set([]byte(key), cacheValue.value)
		}
	}

	// Clear the cache
	ci.cache = make(map[string]cValue)
}

//----------------------------------------
// To cache-wrap this Store further.

// Implements CacheWrapper.
func (ci *Store) CacheWrap() types.CacheWrap {
	return NewStore(ci)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (ci *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return NewStore(tracekv.NewStore(ci, w, tc))
}

//----------------------------------------
// Iteration

// Implements types.KVStore.
func (ci *Store) Iterator(start, end []byte) types.Iterator {
	return ci.iterator(start, end, true)
}

// Implements types.KVStore.
func (ci *Store) ReverseIterator(start, end []byte) types.Iterator {
	return ci.iterator(start, end, false)
}

func (ci *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent, cache types.Iterator

	if ascending {
		parent = ci.parent.Iterator(start, end)
	} else {
		parent = ci.parent.ReverseIterator(start, end)
	}

	items := ci.dirtyItems(start, end, ascending)
	cache = newMemIterator(start, end, items)

	return newCacheMergeIterator(parent, cache, ascending)
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (ci *Store) dirtyItems(start, end []byte, ascending bool) []cmn.KVPair {
	items := make([]cmn.KVPair, 0)

	for key, cacheValue := range ci.cache {
		if !cacheValue.dirty {
			continue
		}
		if dbm.IsKeyInDomain([]byte(key), start, end) {
			items = append(items, cmn.KVPair{Key: []byte(key), Value: cacheValue.value})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if ascending {
			return bytes.Compare(items[i].Key, items[j].Key) < 0
		}
		return bytes.Compare(items[i].Key, items[j].Key) > 0
	})

	return items
}

//----------------------------------------
// etc

func (ci *Store) assertValidKey(key []byte) {
	if key == nil {
		panic("key is nil")
	}
}

func (ci *Store) assertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
}

// Only entrypoint to mutate ci.cache.
func (ci *Store) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	ci.cache[string(key)] = cValue{
		value:   value,
		deleted: deleted,
		dirty:   dirty,
	}
}
