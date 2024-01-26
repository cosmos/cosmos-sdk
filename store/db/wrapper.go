package db

import idb "github.com/cosmos/iavl/db"

// Wrapper wraps a RawDB to implement iavl.DB which is used by iavl.Tree.
type Wrapper struct {
	RawDB
}

var _ idb.DB = (*Wrapper)(nil)

// NewWrapper returns a new Wrapper.
func NewWrapper(db RawDB) *Wrapper {
	return &Wrapper{RawDB: db}
}

// Iterator implements iavl.DB.
func (db *Wrapper) Iterator(start, end []byte) (idb.Iterator, error) {
	return db.RawDB.Iterator(start, end)
}

// ReverseIterator implements iavl.DB.
func (db *Wrapper) ReverseIterator(start, end []byte) (idb.Iterator, error) {
	return db.RawDB.ReverseIterator(start, end)
}

// NewBatch implements iavl.DB.
func (db *Wrapper) NewBatch() idb.Batch {
	return db.RawDB.NewBatch()
}

// NewBatchWithSize implements iavl.DB.
func (db *Wrapper) NewBatchWithSize(size int) idb.Batch {
	return db.RawDB.NewBatchWithSize(size)
}
