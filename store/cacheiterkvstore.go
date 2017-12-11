package store

// TODO: Consider merge w/ tendermint/tmlibs/db/cachedb.go.

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"
)

// If value is nil but deleted is false, it means the parent doesn't have the
// key.  (No need to delete upon Write())
type cValue struct {
	value   []byte
	deleted bool
	dirty   bool
}

// cacheIterKVStore wraps an in-memory cache around an underlying IterKVStore.
type cacheIterKVStore struct {
	mtx         sync.Mutex
	cache       map[string]cValue
	parent      IterKVStore
	lockVersion interface{}

	cwwMutex
}

var _ CacheIterKVStore = (*cacheIterKVStore)(nil)

// Users should typically not be required to call NewCacheIterKVStore directly, as the
// IterKVStore implementations here provide a .CacheIterKVStore() function already.
// `lockVersion` is typically provided by parent.GetWriteLockVersion().
func NewCacheIterKVStore(parent IterKVStore, lockVersion interface{}) *cacheIterKVStore {
	ci := &cacheIterKVStore{
		cache:       make(map[string]cValue),
		parent:      parent,
		lockVersion: lockVersion,
		cwwMutex:    NewCWWMutex(),
	}
	return ci
}

func (ci *cacheIterKVStore) Get(key []byte) (value []byte) {
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

func (ci *cacheIterKVStore) Set(key []byte, value []byte) {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()
	ci.assertValidKey(key)

	ci.setCacheValue(key, value, false, true)
}

func (ci *cacheIterKVStore) Has(key []byte) bool {
	value := ci.Get(key)
	return value != nil
}

func (ci *cacheIterKVStore) Remove(key []byte) {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()
	ci.assertValidKey(key)

	ci.setCacheValue(key, nil, true, true)
}

// Write writes pending updates to the parent database and clears the cache.
// NOTE: Not atomic.
func (ci *cacheIterKVStore) Write() {
	ci.mtx.Lock()
	defer ci.mtx.Unlock()

	// Optional sanity check to ensure that cacheIterKVStore is valid
	if parent, ok := ci.parent.(WriteLocker); ok {
		if parent.TryWriteLock(ci.lockVersion) {
			// All good!
		} else {
			panic("parent.Write() failed. Did this CacheIterKVStore expire?")
		}
	}

	// We need a copy of all of the keys.
	// Not the best, but probably not a bottleneck depending.
	keys := make([]string, 0, len(ci.cache))
	for key, dbValue := range ci.cache {
		if dbValue.dirty {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	// TODO in tmlibs/db we use Batch to write atomically.
	// Consider locking the underlying IterKVStore during write.
	for _, key := range keys {
		cacheValue := ci.cache[key]
		if cacheValue.deleted {
			ci.parent.Remove([]byte(key))
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
// To cache-wrap this cacheIterKVStore further.

func (ci *cacheIterKVStore) CacheWrap() CacheWrap {
	return ci.CacheIterKVStore()
}

func (ci *cacheIterKVStore) CacheKVStore() CacheKVStore {
	return ci.CacheIterKVStore()
}

func (ci *cacheIterKVStore) CacheIterKVStore() CacheIterKVStore {
	return NewCacheIterKVStore(ci, ci.GetWriteLockVersion())
}

// If the parent parent DB implements this, (e.g. such as a cacheIterKVStore
// parent to a cacheIterKVStore child), cacheIterKVStore will call
// `parent.TryWriteLock()` before attempting to write.
// This prevents multiple siblings from Write'ing to the parent.
type WriteLocker interface {
	GetWriteLockVersion() (lockVersion interface{})
	TryWriteLock(lockVersion interface{}) bool
}

// Implements TryWriteLocker.  Embed this in DB structs if desired.
type cwwMutex struct {
	mtx sync.Mutex
	// CONTRACT: reading/writing to `*written` should use `atomic.*`.
	// CONTRACT: replacing `written` with another *int32 should use `.mtx`.
	written *int32
}

func NewCWWMutex() cwwMutex {
	return cwwMutex{
		written: new(int32),
	}
}

func (cww *cwwMutex) GetWriteLockVersion() interface{} {
	cww.mtx.Lock()
	defer cww.mtx.Unlock()

	// `written` works as a "version" object because it gets replaced upon
	// successful TryWriteLock.
	return cww.written
}

func (cww *cwwMutex) TryWriteLock(version interface{}) bool {
	cww.mtx.Lock()
	defer cww.mtx.Unlock()

	if version != cww.written {
		return false // wrong "WriteLockVersion"
	}
	if !atomic.CompareAndSwapInt32(cww.written, 0, 1) {
		return false // already written
	}

	// New "WriteLockVersion"
	cww.written = new(int32)
	return true
}

//----------------------------------------
// Iteration

func (ci *cacheIterKVStore) Iterator(start, end []byte) Iterator {
	return ci.iterator(start, end, true)
}

func (ci *cacheIterKVStore) ReverseIterator(start, end []byte) Iterator {
	return ci.iterator(start, end, false)
}

func (ci *cacheIterKVStore) iterator(start, end []byte, ascending bool) Iterator {
	var parent, cache Iterator
	if ascending {
		parent = ci.parent.Iterator(start, end)
	} else {
		parent = ci.parent.ReverseIterator(start, end)
	}
	items := ci.dirtyItems(ascending)
	cache = newMemIterator(start, end, items)
	return newCacheMergeIterator(parent, cache, ascending)
}

func (ci *cacheIterKVStore) First(start, end []byte) (kv KVPair, ok bool) {
	return iteratorFirst(ci, start, end)
}

func (ci *cacheIterKVStore) Last(start, end []byte) (kv KVPair, ok bool) {
	return iteratorLast(ci, start, end)
}

// Constructs a slice of dirty items, to use w/ memIterator.
func (ci *cacheIterKVStore) dirtyItems(ascending bool) []KVPair {
	items := make([]KVPair, 0, len(ci.cache))
	for key, cacheValue := range ci.cache {
		if !cacheValue.dirty {
			continue
		}
		items = append(items,
			KVPair{[]byte(key), cacheValue.value})
	}
	sort.Slice(items, func(i, j int) bool {
		if ascending {
			return bytes.Compare(items[i].Key, items[j].Key) < 0
		} else {
			return bytes.Compare(items[i].Key, items[j].Key) > 0
		}
	})
	return items
}

//----------------------------------------
// etc

func (ci *cacheIterKVStore) assertValidKey(key []byte) {
	if key == nil {
		panic("key is nil")
	}
}

// Only entrypoint to mutate ci.cache.
func (ci *cacheIterKVStore) setCacheValue(key, value []byte, deleted bool, dirty bool) {
	cacheValue := cValue{
		value:   value,
		deleted: deleted,
		dirty:   dirty,
	}
	ci.cache[string(key)] = cacheValue
}
