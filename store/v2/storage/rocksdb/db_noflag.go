//go:build !rocksdb
// +build !rocksdb

package rocksdb

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage"
)

var (
	_ storage.Database         = (*Database)(nil)
	_ store.UpgradableDatabase = (*Database)(nil)
)

type Database struct{}

func New(dataDir string) (*Database, error) {
	return &Database{}, nil
}

func (db *Database) Close() error {
	return nil
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	panic("rocksdb requires a build flag")
}

func (db *Database) SetLatestVersion(version uint64) error {
	panic("rocksdb requires a build flag")
}

func (db *Database) GetLatestVersion() (uint64, error) {
	panic("rocksdb requires a build flag")
}

func (db *Database) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	panic("rocksdb requires a build flag")
}

func (db *Database) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	panic("rocksdb requires a build flag")
}

// Prune prunes all versions up to and including the provided version argument.
// Internally, this performs a manual compaction, the data with older timestamp
// will be GCed by compaction.
func (db *Database) Prune(version uint64) error {
	panic("rocksdb requires a build flag")
}

func (db *Database) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb requires a build flag")
}

func (db *Database) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	panic("rocksdb requires a build flag")
}

// PruneStoreKeys will do nothing for RocksDB, it will be pruned by compaction
// when the version is pruned
func (db *Database) PruneStoreKeys(_ []string, _ uint64) error {
	return nil
}
