package db

import (
	idb "github.com/cosmos/iavl/db"

	corestore "cosmossdk.io/core/store"
)

// Wrapper wraps a `corestore.KVStoreWithBatch` to implement iavl.DB which is used by iavl.Tree.
type Wrapper struct {
	corestore.KVStoreWithBatch
}

var _ idb.DB = (*Wrapper)(nil)

// NewWrapper returns a new Wrapper.
func NewWrapper(db corestore.KVStoreWithBatch) *Wrapper {
	return &Wrapper{KVStoreWithBatch: db}
}

// Iterator implements iavl.DB.
func (db *Wrapper) Iterator(start, end []byte) (idb.Iterator, error) {
	return db.KVStoreWithBatch.Iterator(start, end)
}

// ReverseIterator implements iavl.DB.
func (db *Wrapper) ReverseIterator(start, end []byte) (idb.Iterator, error) {
	return db.KVStoreWithBatch.ReverseIterator(start, end)
}

// NewBatch implements iavl.DB.
func (db *Wrapper) NewBatch() idb.Batch {
	return db.KVStoreWithBatch.NewBatch()
}

// NewBatchWithSize implements iavl.DB.
func (db *Wrapper) NewBatchWithSize(size int) idb.Batch {
	return db.KVStoreWithBatch.NewBatchWithSize(size)
}
