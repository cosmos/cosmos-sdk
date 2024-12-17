package internal

import (
	"io"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/internal/btree"
	"cosmossdk.io/store/types"
)

var _ types.KVStore = (*BTreeStore[[]byte])(nil)

// BTreeStore is a wrapper for a BTree with GKVStore[V] implementation
type BTreeStore[V any] struct {
	btree.BTree[V]
	isZero   func(V) bool
	valueLen func(V) int
}

// NewBTreeStore constructs new BTree adapter
func NewBTreeStore[V any](btree btree.BTree[V], isZero func(V) bool, valueLen func(V) int) *BTreeStore[V] {
	return &BTreeStore[V]{btree, isZero, valueLen}
}

// Hash Implements GKVStore.
func (ts *BTreeStore[V]) Has(key []byte) bool {
	return !ts.isZero(ts.Get(key))
}

func (ts *BTreeStore[V]) Iterator(start, end []byte) types.GIterator[V] {
	it, err := ts.BTree.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

func (ts *BTreeStore[V]) ReverseIterator(start, end []byte) types.GIterator[V] {
	it, err := ts.BTree.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

// GetStoreType returns the type of the store.
func (ts *BTreeStore[V]) GetStoreType() types.StoreType {
	return types.StoreTypeDB
}

// CacheWrap branches the underlying store.
func (ts *BTreeStore[V]) CacheWrap() types.CacheWrap {
	return cachekv.NewGStore(ts, ts.isZero, ts.valueLen)
}

// CacheWrapWithTrace branches the underlying store.
func (ts *BTreeStore[V]) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewGStore(ts, ts.isZero, ts.valueLen)
}
