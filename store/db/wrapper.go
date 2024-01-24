package db

import idb "github.com/cosmos/iavl/db"

// Wrapper wraps a DB to implement iavl.DB which is used by iavl.Tree.
type Wrapper struct {
	DB
}

var _ idb.DB = (*Wrapper)(nil)

// NewWrapper returns a new Wrapper.
func NewWrapper(db DB) *Wrapper {
	return &Wrapper{DB: db}
}

// Iterator implements iavl.DB.
func (db *Wrapper) Iterator(start, end []byte) (idb.Iterator, error) {
	return db.DB.Iterator(start, end)
}

// ReverseIterator implements iavl.DB.
func (db *Wrapper) ReverseIterator(start, end []byte) (idb.Iterator, error) {
	return db.DB.ReverseIterator(start, end)
}

// NewBatch implements iavl.DB.
func (db *Wrapper) NewBatch() idb.Batch {
	return db.DB.NewBatch()
}

// NewBatchWithSize implements iavl.DB.
func (db *Wrapper) NewBatchWithSize(size int) idb.Batch {
	return db.DB.NewBatchWithSize(size)
}
