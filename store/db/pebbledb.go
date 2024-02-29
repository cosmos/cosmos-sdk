package db

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/pebble"
)

var _ store.RawDB = (*PebbleDB)(nil)

// PebbleDB implements RawDB using PebbleDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use PebbleDB as
// the IAVL v0/v1 backend.
type PebbleDB struct {
	db *pebble.DB
}

func (db *PebbleDB) Get([]byte) ([]byte, error) {
	panic("not implemented!")
}

func (db *PebbleDB) Has(key []byte) (bool, error) {
	panic("not implemented!")
}

func (db *PebbleDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *PebbleDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *PebbleDB) Close() error {
	panic("not implemented!")
}

func (db *PebbleDB) NewBatch() RawBatch {
	panic("not implemented!")
}

func (db *PebbleDB) NewBatchWithSize(int) RawBatch {
	panic("not implemented!")
}
