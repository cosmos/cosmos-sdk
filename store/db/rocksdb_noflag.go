//go:build !rocksdb
// +build !rocksdb

package db

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var (
	_ store.RawDB = (*RocksDB)(nil)
)

// RocksDB implements RawDB using RocksDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct {
}

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

func (db *RocksDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) NewBatch() store.RawBatch {
	panic("rocksdb must be built with -tags rocksdb")
}

func (db *RocksDB) NewBatchWithSize(_ int) store.RawBatch {
	return db.NewBatch()
}

var _ corestore.Iterator = (*rocksDBIterator)(nil)

type rocksDBIterator struct {
}

func (itr *rocksDBIterator) Domain() (start, end []byte) {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Valid() bool {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Key() []byte {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Value() []byte {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Next() {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Error() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) Close() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (itr *rocksDBIterator) assertIsValid() {
	panic("rocksdb must be built with -tags rocksdb")
}

type rocksDBBatch struct {
}

func (b *rocksDBBatch) Set(key, value []byte) error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (b *rocksDBBatch) Delete(key []byte) error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (b *rocksDBBatch) Write() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (b *rocksDBBatch) WriteSync() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (b *rocksDBBatch) Close() error {
	panic("rocksdb must be built with -tags rocksdb")
}

func (b *rocksDBBatch) GetByteSize() (int, error) {
	panic("rocksdb must be built with -tags rocksdb")
}
