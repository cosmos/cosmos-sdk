//go:build rocksdb
// +build rocksdb

package db

import (
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

var (
	_ store.RawDB = (*RocksDB)(nil)

	defaultReadOpts = grocksdb.NewDefaultReadOptions()
)

// RocksDB implements RawDB using RocksDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct {
	storage *grocksdb.DB
}

func NewRocksDB(dataDir string) (*RocksDB, error) {
	opts := grocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)

	storage, err := grocksdb.OpenDb(opts, dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	return &RocksDB{
		storage: storage,
	}, nil
}

func NewRocksDBWithOpts(dataDir string, opts *grocksdb.Options) (*RocksDB, error) {
	storage, err := grocksdb.OpenDb(opts, dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	return &RocksDB{
		storage: storage,
	}, nil
}

func (db *RocksDB) Close() error {
	db.storage.Close()
	db.storage = nil
	return nil
}

func (db *RocksDB) Get(key []byte) ([]byte, error) {
	bz, err := db.storage.GetBytes(defaultReadOpts, key)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func (db *RocksDB) Has(key []byte) (bool, error) {
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}

	return bz != nil, nil
}

func (db *RocksDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *RocksDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *RocksDB) NewBatch() store.RawBatch {
	panic("not implemented!")
}

func (db *RocksDB) NewBatchWithSize(int) store.RawBatch {
	panic("not implemented!")
}
