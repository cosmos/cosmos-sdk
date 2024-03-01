//go:build rocksdb
// +build rocksdb

package db

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

var _ store.RawDB = (*RocksDB)(nil)

// RocksDB implements RawDB using RocksDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct {
	storage *grocksdb.DB
}

func (db *RocksDB) Get([]byte) ([]byte, error) {
	panic("not implemented!")
}

func (db *RocksDB) Has(key []byte) (bool, error) {
	panic("not implemented!")
}

func (db *RocksDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *RocksDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *RocksDB) Close() error {
	panic("not implemented!")
}

func (db *RocksDB) NewBatch() store.RawBatch {
	panic("not implemented!")
}

func (db *RocksDB) NewBatchWithSize(int) store.RawBatch {
	panic("not implemented!")
}
