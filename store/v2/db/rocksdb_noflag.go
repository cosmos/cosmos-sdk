//go:build !rocksdb
// +build !rocksdb

package db

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var _ corestore.KVStoreWithBatch = (*RocksDB)(nil)

// RocksDB implements `corestore.KVStoreWithBatch` using RocksDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct{}

func NewRocksDB(name, dataDir string) (*RocksDB, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func NewRocksDBWithOpts(dataDir string, opts store.DBOptions) (*RocksDB, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Close() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Get(key []byte) ([]byte, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Has(key []byte) (bool, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Set(key, value []byte) error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Delete(key []byte) error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) NewBatch() corestore.Batch {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) NewBatchWithSize(_ int) corestore.Batch {
	panic("rocksdb must be built with -tags rocksdb")
}
