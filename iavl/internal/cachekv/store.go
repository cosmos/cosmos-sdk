package cachekv

import (
	"io"
	"iter"
	"unsafe"

	"github.com/tidwall/btree"

	storetypes "cosmossdk.io/store/types"
)

// Store is an in-memory KV store that buffers writes and deletions until Write() is called.
// Unlike earlier versions of the cachekv.Store, we do not do any caching of reads,
// as that responsibility should be shifted to the base layer.
// Also, there is no mutex and the caller contract is the following:
// - concurrent writes (calls to Set/Delete) are not allowed and may cause undefined behavior
// - concurrent reads (calls to Get/Has/Iterator) are allowed, as long as there are no concurrent writes
type Store struct {
	parent   storetypes.KVStore
	writeMap btree.Map[string, []byte]
	dirty    bool
}

func NewStore(parent storetypes.KVStore) *Store {
	return &Store{parent: parent}
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
	storetypes.AssertValidKey(key)

	if !store.dirty {
		return store.parent.Get(key)
	}

	value, ok := store.getDirty(key)
	if ok {
		return value
	}

	return store.parent.Get(key)
}

func (store *Store) Has(key []byte) bool {
	storetypes.AssertValidKey(key)

	if !store.dirty {
		return store.parent.Has(key)
	}

	value, found := store.getDirty(key)
	if found {
		return value != nil
	}
	return store.parent.Has(key)
}

func (store *Store) getDirty(key []byte) (value []byte, found bool) {
	// we use unsafeBytesToString here because we are doing a lookup, we are not modifying the key, and we want to avoid unnecessary allocations
	return store.writeMap.Get(unsafeBytesToString(key))
}

func (store *Store) Set(key, value []byte) {
	storetypes.AssertValidKey(key)
	storetypes.AssertValidValue(value)

	store.dirty = true

	store.writeMap.Set(string(key), value)
}

func (store *Store) Delete(key []byte) {
	storetypes.AssertValidKey(key)

	store.dirty = true

	store.writeMap.Set(string(key), nil)
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
	if !store.dirty {
		return func(yield func(KVUpdate) bool) {}
	}

	return func(yield func(KVUpdate) bool) {
		store.writeMap.Scan(func(key string, value []byte) bool {
			update := KVUpdate{
				Key:    []byte(key), // this introduces a small amount of allocation and copying, but is safer
				Value:  value,
				Delete: value == nil,
			}
			return yield(update)
		})
	}
}

func (store *Store) Write() {
	if !store.dirty {
		return
	}

	store.writeMap.Scan(func(key string, value []byte) bool {
		if value == nil {
			store.parent.Delete([]byte(key)) // this introduces a small amount of allocation and copying, but is safer
		} else {
			store.parent.Set([]byte(key), value) // this introduces a small amount of allocation and copying, but is safer
		}
		return true
	})

	store.writeMap.Clear()
}

func (store *Store) iterator(start, end []byte, ascending bool) storetypes.Iterator {
	if !store.dirty {
		if ascending {
			return store.parent.Iterator(start, end)
		} else {
			return store.parent.ReverseIterator(start, end)
		}
	}

	var (
		parent, writeIter storetypes.Iterator
	)

	if ascending {
		parent = store.parent.Iterator(start, end)
		writeIter = newMemIterator(start, end, &store.writeMap, true)
	} else {
		parent = store.parent.ReverseIterator(start, end)
		writeIter = newMemIterator(start, end, &store.writeMap, false)
	}

	return NewCacheMergeIterator(parent, writeIter, ascending)
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// This should be used with caution and only when the byte slice is not modified.
// But generally when we are storing a byte slice as a key in a map, this is what we should use.
func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
