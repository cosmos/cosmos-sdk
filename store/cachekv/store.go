package cachekv

import (
	"io"
	"iter"
	"unsafe"

	"github.com/tidwall/btree"

	"cosmossdk.io/store/cachekv/internal"
	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
)

// GStore is an in-memory KV store that buffers writes and deletions until Write() is called.
// Unlike earlier versions of the cachekv.Store, we do not do any caching of reads,
// as that responsibility should be shifted to the base layer.
// Also, there is no mutex and the caller contract is the following:
// - concurrent writes (calls to Set/Delete) are not allowed and may cause undefined behavior
// - concurrent reads (calls to Get/Has/Iterator) are allowed, as long as there are no concurrent writes
type GStore[V any] struct {
	parent   types.GKVStore[V]
	writeMap btree.Map[string, V]
	dirty    bool

	// isZero is a function that returns true if the value is considered "zero", for []byte and pointers the zero value
	// is `nil`, the zero value is not allowed to set to a key, and it's returned if the key is not found.
	isZero func(V) bool
	// valueLen validates the value before it's set
	valueLen func(V) int
}

type Store = GStore[[]byte]

func NewGStore[V any](parent types.GKVStore[V], isZero func(V) bool, valueLen func(V) int) *GStore[V] {
	return &GStore[V]{parent: parent, isZero: isZero, valueLen: valueLen}
}

func NewStore(parent types.KVStore) *Store {
	return NewGStore[[]byte](
		parent,
		types.BytesIsZero,
		types.BytesValueLen)
}

func (store *GStore[V]) GetStoreType() types.StoreType {
	return types.StoreTypeIAVL
}

func (store *GStore[V]) CacheWrap() types.CacheWrap {
	return NewGStore[V](store, store.isZero, store.valueLen)
}

func (store *GStore[V]) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	// We need to make a type assertion here as the tracekv store requires bytes value types for serialization.
	if store, ok := any(store).(*GStore[[]byte]); ok {
		return NewStore(tracekv.NewStore(store, w, tc))
	}
	return store.CacheWrap()
}

func (store *GStore[V]) Get(key []byte) V {
	types.AssertValidKey(key)

	if !store.dirty {
		return store.parent.Get(key)
	}

	value, ok := store.getDirty(key)
	if ok {
		return value
	}

	return store.parent.Get(key)
}

func (store *GStore[V]) Has(key []byte) bool {
	types.AssertValidKey(key)

	if !store.dirty {
		return store.parent.Has(key)
	}

	value, found := store.getDirty(key)
	if found {
		return !store.isZero(value)
	}
	return store.parent.Has(key)
}

func (store *GStore[V]) getDirty(key []byte) (value V, found bool) {
	// we use unsafeBytesToString here because we are doing a lookup, we are not modifying the key, and we want to avoid unnecessary allocations
	return store.writeMap.Get(unsafeBytesToString(key))
}

func (store *GStore[V]) Set(key []byte, value V) {
	types.AssertValidKey(key)
	types.AssertValidValueGeneric(value, store.isZero, store.valueLen)

	store.dirty = true

	store.writeMap.Set(string(key), value)
}

func (store *GStore[V]) Delete(key []byte) {
	types.AssertValidKey(key)

	store.dirty = true

	var zeroValue V
	store.writeMap.Set(string(key), zeroValue)
}

func (store *GStore[V]) Iterator(start, end []byte) types.GIterator[V] {
	return store.iterator(start, end, true)
}

func (store *GStore[V]) ReverseIterator(start, end []byte) types.GIterator[V] {
	return store.iterator(start, end, false)
}

// Update is a type used to represent a pending update to the underlying store in the cache,
// it can be either a set or a delete depending on the Delete field.
// This is defined as a type alias so that other code can use the same struct without a direct import.
type Update[V any] = struct {
	// Key is the key to be updated.
	Key []byte
	// Value is the value to be set. It is ignored if Delete is true, but should be set to the zero value of V.
	Value V
	// Delete indicates whether this update is a deletion.
	Delete bool
}

// Updates returns the cached updates to be applied to the underlying commitment store.
// This should be preferred over calling Write against the underlying commitment store
// because it allows for better performance.
// Actually calling Write when this is the first cache layer on top of iavl will result in a panic.
func (store *GStore[V]) Updates() (updates iter.Seq[Update[V]], count int) {
	if !store.dirty {
		return func(yield func(Update[V]) bool) {}, 0
	}

	return func(yield func(Update[V]) bool) {
		store.writeMap.Scan(func(key string, value V) bool {
			return yield(Update[V]{
				Key:    []byte(key), // casting to []byte introduces a small amount of allocation and copying, but is safer
				Value:  value,
				Delete: store.isZero(value),
			})
		})
	}, store.writeMap.Len()
}

func (store *GStore[V]) Write() {
	if !store.dirty {
		return
	}

	store.writeMap.Scan(func(key string, value V) bool {
		if store.isZero(value) {
			store.parent.Delete([]byte(key)) // this introduces a small amount of allocation and copying, but is safer
		} else {
			store.parent.Set([]byte(key), value) // this introduces a small amount of allocation and copying, but is safer
		}
		return true
	})

	store.writeMap.Clear()
}

func (store *GStore[V]) iterator(start, end []byte, ascending bool) types.GIterator[V] {
	if !store.dirty {
		if ascending {
			return store.parent.Iterator(start, end)
		} else {
			return store.parent.ReverseIterator(start, end)
		}
	}

	var (
		parent, writeIter types.GIterator[V]
	)

	if ascending {
		parent = store.parent.Iterator(start, end)
		writeIter = newMemIterator(start, end, &store.writeMap, true)
	} else {
		parent = store.parent.ReverseIterator(start, end)
		writeIter = newMemIterator(start, end, &store.writeMap, false)
	}

	return internal.NewCacheMergeIterator(parent, writeIter, ascending, store.isZero)
}

// unsafeBytesToString converts a byte slice to a string without allocation.
// This should be used with caution and only when the byte slice is not modified.
// But generally when we are storing a byte slice as a key in a map, this is what we should use.
func unsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
