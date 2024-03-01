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

func (db *RocksDB) NewBatchWithSize(size int) store.RawBatch {
	panic("not implemented!")
}

var _ store.RawBatch = (*rocksDBBatch)(nil)

type rocksDBBatch struct {
	db    *RocksDB
	batch *grocksdb.WriteBatch
}

func (b *rocksDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if value == nil {
		return store.ErrValueNil
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}

	b.batch.Put(key, value)
	return nil
}

func (b *rocksDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}

	b.batch.Delete(key)
	return nil
}

func (b *rocksDBBatch) Write() error {
	writeOpts := grocksdb.NewDefaultWriteOptions()
	writeOpts.SetSync(false)

	if err := b.db.storage.Write(writeOpts, b.batch); err != nil {
		return fmt.Errorf("failed to write RocksDB batch: %w", err)
	}

	return nil
}

func (b *rocksDBBatch) WriteSync() error {
	writeOpts := grocksdb.NewDefaultWriteOptions()
	writeOpts.SetSync(true)

	if err := b.db.storage.Write(writeOpts, b.batch); err != nil {
		return fmt.Errorf("failed to write RocksDB batch: %w", err)
	}

	return nil
}

func (b *rocksDBBatch) Close() error {
	b.batch.Destroy()
	return nil
}

func (b *rocksDBBatch) GetByteSize() (int, error) {
	return len(b.batch.Data()), nil
}
