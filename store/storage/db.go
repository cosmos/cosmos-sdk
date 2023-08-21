package storage

import (
	"sync"

	"cosmossdk.io/store/v2"
)

// TODO:
//
// - Define and implement pruning method(s)
// - Import/snapshot method(s)???
//
// Ref: https://github.com/cosmos/cosmos-sdk/issues/17279

var (
	_ store.VersionedReaderWriter    = (*Database)(nil)
	_ store.VersionedIteratorCreator = (*Database)(nil)
	_ store.VersionedBatcher         = (*Database)(nil)
)

// Database defines a thread-safe state storage (SS) engine. There should only
// be a single instantiated SS Database in use upstream. A SS Database is instantiated
// with a VersionedDatabase, which is configurable by the caller.
//
// Note, a caller should use NewBatch() to create a batch for writing key/value
// pairs. In addition, writes are NOT buffered, which is left as the responsibility
// to the upstream caller. I.e. writes should be buffered until a commit is signaled,
// which can then write a batch to disk.
type Database struct {
	mu  sync.Mutex
	vdb store.VersionedDatabase
}

func New(vdb store.VersionedDatabase) *Database {
	db := &Database{
		vdb: vdb,
	}

	return db
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.vdb.Close(); err != nil {
		return err
	}

	// reset all mutable fields
	db.vdb = nil

	return nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.vdb.Has(storeKey, version, key)
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.vdb.Get(storeKey, version, key)
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.vdb.Set(storeKey, version, key, value)
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.vdb.Delete(storeKey, version, key)
}

func (db *Database) isClosed() bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.vdb == nil
}

func (db *Database) GetLatestVersion() (uint64, error) {
	return db.vdb.GetLatestVersion()
}

func (db *Database) SetLatestVersion(version uint64) error {
	return db.vdb.SetLatestVersion(version)
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	return db.vdb.NewBatch(version)
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	return db.vdb.NewIterator(storeKey, version, start, end)
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	return db.vdb.NewReverseIterator(storeKey, version, start, end)
}
