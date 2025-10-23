package blockstm

import (
	"io"

	"github.com/tidwall/btree"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/tracekv"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/blockstm/tree"
)

type (
	MemDB    = GMemDB[[]byte]
	ObjMemDB = GMemDB[any]
)

var (
	_ storetypes.KVStore    = (*MemDB)(nil)
	_ storetypes.ObjKVStore = (*ObjMemDB)(nil)
)

// NewMemDB constructs a new in memory store over a []byte value type.
func NewMemDB() *MemDB {
	return NewGMemDB(storetypes.BytesIsZero, storetypes.BytesValueLen)
}

// NewObjMemDB constructs a new in memory store over a generic any type.
func NewObjMemDB() *ObjMemDB {
	return NewGMemDB(storetypes.AnyIsZero, storetypes.AnyValueLen)
}

// GMemDB is a generic implementation of an in memory Store backed by tidwall/btree.
type GMemDB[V any] struct {
	btree.BTreeG[memdbItem[V]]
	isZero   func(V) bool
	valueLen func(V) int
}

// NewGMemDB is the generic constructor for a GMemDB.
func NewGMemDB[V any](
	isZero func(V) bool,
	valueLen func(V) int,
) *GMemDB[V] {
	return &GMemDB[V]{
		BTreeG:   *btree.NewBTreeG[memdbItem[V]](tree.KeyItemLess),
		isZero:   isZero,
		valueLen: valueLen,
	}
}

// NewGMemDBNonConcurrent returns a new BTree which is not concurrency safe.
func NewGMemDBNonConcurrent[V any](
	isZero func(V) bool,
	valueLen func(V) int,
) *GMemDB[V] {
	return &GMemDB[V]{
		BTreeG: *btree.NewBTreeGOptions[memdbItem[V]](tree.KeyItemLess, btree.Options{
			NoLocks: true,
		}),
		isZero:   isZero,
		valueLen: valueLen,
	}
}

func (db *GMemDB[V]) Scan(cb func(key Key, value V) bool) {
	db.BTreeG.Scan(func(item memdbItem[V]) bool {
		return cb(item.key, item.value)
	})
}

func (db *GMemDB[V]) Get(key []byte) V {
	item, ok := db.BTreeG.Get(memdbItem[V]{key: key})
	if !ok {
		var empty V
		return empty
	}
	return item.value
}

func (db *GMemDB[V]) Has(key []byte) bool {
	return !db.isZero(db.Get(key))
}

func (db *GMemDB[V]) Set(key []byte, value V) {
	if db.isZero(value) {
		panic("nil value not allowed")
	}
	db.BTreeG.Set(memdbItem[V]{key: key, value: value})
}

func (db *GMemDB[V]) Delete(key []byte) {
	db.BTreeG.Delete(memdbItem[V]{key: key})
}

// OverlayGet returns a value from the btree and true if we found a value.
// When used as an overlay (e.g. WriteSet), it stores the `nil` value to represent deleted keys,
// so we return separate bool value for found status.
func (db *GMemDB[V]) OverlayGet(key Key) (V, bool) {
	item, ok := db.BTreeG.Get(memdbItem[V]{key: key})
	if !ok {
		var zero V
		return zero, false
	}
	return item.value, true
}

// OverlaySet sets a value in the btree
// When used as an overlay (e.g. WriteSet), it stores the `nil` value to represent deleted keys,
func (db *GMemDB[V]) OverlaySet(key Key, value V) {
	db.BTreeG.Set(memdbItem[V]{key: key, value: value})
}

func (db *GMemDB[V]) Iterator(start, end []byte) storetypes.GIterator[V] {
	return db.iterator(start, end, true)
}

func (db *GMemDB[V]) ReverseIterator(start, end []byte) storetypes.GIterator[V] {
	return db.iterator(start, end, false)
}

func (db *GMemDB[V]) iterator(start, end Key, ascending bool) storetypes.GIterator[V] {
	return NewMemDBIterator(start, end, db.Iter(), ascending)
}

func (db *GMemDB[V]) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

// CacheWrap implements types.KVStore.
func (db *GMemDB[V]) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewGStore(db, db.isZero, db.valueLen)
}

// CacheWrapWithTrace implements types.KVStore.
func (db *GMemDB[V]) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	if store, ok := any(db).(*GMemDB[[]byte]); ok {
		return cachekv.NewGStore(tracekv.NewStore(store, w, tc), store.isZero, store.valueLen)
	}
	return db.CacheWrap()
}

// MemDBIterator wraps a generic BTreeIteratorG over a memdbItem.
// It is used as an iterator over a GMemDB implementation.
type MemDBIterator[V any] struct {
	tree.BTreeIteratorG[memdbItem[V]]
}

var _ storetypes.Iterator = (*MemDBIterator[[]byte])(nil)

func NewMemDBIterator[V any](start, end Key, iter btree.IterG[memdbItem[V]], ascending bool) *MemDBIterator[V] {
	return &MemDBIterator[V]{*tree.NewBTreeIteratorG(
		memdbItem[V]{key: start},
		memdbItem[V]{key: end},
		iter,
		ascending,
	)}
}

// NewNoopIterator constructs a storetypes.GIterator with an invalidated wrapped iterator.
func NewNoopIterator[V any](start, end Key, ascending bool) storetypes.GIterator[V] {
	return &MemDBIterator[V]{tree.NewNoopBTreeIteratorG[memdbItem[V]](
		start,
		end,
		ascending,
		false,
	)}
}

func (it *MemDBIterator[V]) Value() V {
	return it.Item().value
}

type memdbItem[V any] struct {
	key   Key
	value V
}

var _ tree.KeyItem = memdbItem[[]byte]{}

func (item memdbItem[V]) GetKey() []byte {
	return item.key
}

type MultiMemDB struct {
	dbs map[storetypes.StoreKey]storetypes.Store
}

var _ MultiStore = (*MultiMemDB)(nil)

func NewMultiMemDB(stores map[storetypes.StoreKey]int) *MultiMemDB {
	dbs := make(map[storetypes.StoreKey]storetypes.Store, len(stores))
	for name := range stores {
		switch name.(type) {
		case *storetypes.ObjectStoreKey:
			dbs[name] = NewObjMemDB()
		default:
			dbs[name] = NewMemDB()
		}
	}
	return &MultiMemDB{
		dbs: dbs,
	}
}

func (mmdb *MultiMemDB) GetStore(store storetypes.StoreKey) storetypes.Store {
	return mmdb.dbs[store]
}

func (mmdb *MultiMemDB) GetKVStore(store storetypes.StoreKey) storetypes.KVStore {
	return mmdb.GetStore(store).(storetypes.KVStore)
}

func (mmdb *MultiMemDB) GetObjKVStore(store storetypes.StoreKey) storetypes.ObjKVStore {
	return mmdb.GetStore(store).(storetypes.ObjKVStore)
}
