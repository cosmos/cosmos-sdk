package storage

import (
	"fmt"
	"sync"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"cosmossdk.io/store/v2"
)

// TODO: @bez
//
// - Implement iterator methods
// - Define and implement pruning method(s)
// - Import/snapshot method(s)???

// DefaultMemDBSize defines the default pre-allocation size of the in-memory
// database used to accumulate database entries prior to committing them to disk.
const DefaultMemDBSize = 1000

var (
	_ store.VersionedReaderWriter    = (*Database)(nil)
	_ store.VersionedIteratorCreator = (*Database)(nil)
	_ store.Committer                = (*Database)(nil)
)

// Database defines the state storage (SS) backend. Each open instance of a Database
// is meant to commit key/value pairs to disk upon Commit for a single version.
// However, it can be used to read key/value pairs at any version.
type Database struct {
	mu      sync.RWMutex
	vdb     store.VersionedDatabase
	version uint64
	memSize int
	mem     map[string]map[string]dbEntry
	batch   store.Batch
}

// dbEntry defines a database entry stored in memory until the batch is written
// on commit.
type dbEntry struct {
	value   []byte
	delete  bool
	version uint64
}

// New creates a new instance of a state storage (SS) backend Database and returns
// a reference to it. It takes a backing physical storage backend, the current
// version/height that will be committed, the list of store keys to support and
// the pre-allocation size for each store key.
func New(vdb store.VersionedDatabase, version uint64, storeKeys []string, memSize int) *Database {
	db := &Database{
		vdb:     vdb,
		version: version,
		memSize: memSize,
		batch:   vdb.NewBatch(version),
	}

	for _, storeKey := range storeKeys {
		db.mem[storeKey] = make(map[string]dbEntry, memSize)
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
	db.mem = nil
	db.batch = nil

	return nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.mem == nil {
		return false, store.ErrClosed
	}
	// TODO: Should we enforce reads have a store key set?

	if entry, ok := db.mem[storeKey][string(key)]; ok && entry.version == version {
		return !entry.delete, nil
	}

	return db.vdb.Has(storeKey, version, key)
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.mem == nil {
		return nil, store.ErrClosed
	}
	// TODO: Should we enforce reads have a store key set?

	if entry, ok := db.mem[storeKey][string(key)]; ok && entry.version == version {
		if entry.delete {
			return nil, store.ErrRecordNotFound
		}

		return slices.Clone(entry.value), nil
	}

	return db.vdb.Get(storeKey, version, key)
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.mem == nil {
		return store.ErrClosed
	}
	if version != db.version {
		return fmt.Errorf("expected %d, got %d: %w", store.ErrInvalidVersion, db.version, version)
	}
	if _, hasStoreKey := db.mem[storeKey]; !hasStoreKey {
		return fmt.Errorf("%s: %w", store.ErrUnknownStoreKey, storeKey)
	}

	db.mem[storeKey][string(key)] = dbEntry{value: slices.Clone(value), version: version}
	return nil
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.mem == nil {
		return store.ErrClosed
	}
	if version != db.version {
		return fmt.Errorf("expected %d, got %d: %w", store.ErrInvalidVersion, db.version, version)
	}
	if _, hasStoreKey := db.mem[storeKey]; !hasStoreKey {
		return fmt.Errorf("%s: %w", store.ErrUnknownStoreKey, storeKey)
	}

	db.mem[storeKey][string(key)] = dbEntry{delete: true, version: version}

	return nil
}

func (db *Database) Commit() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.commitBatch(); err != nil {
		return err
	}

	db.reset()

	return nil
}

func (db *Database) reset() {
	maps.Clear(db.mem)
}

func (db *Database) commitBatch() error {
	if db.mem == nil {
		return store.ErrClosed
	}

	db.batch.Reset()

	for storeKey, entries := range db.mem {
		for key, entry := range entries {
			if entry.delete {
				if err := db.batch.Delete(storeKey, []byte(key)); err != nil {
					return err
				}
			} else if err := db.batch.Set(storeKey, []byte(key), entry.value); err != nil {
				return err
			}
		}
	}

	if err := db.batch.Write(); err != nil {
		return fmt.Errorf("failed to write state storage database batch: %w", err)
	}

	db.batch.Reset()

	return nil
}

func (db *Database) GetLatestVersion() (uint64, error) {
	return db.vdb.GetLatestVersion()
}

func (db *Database) NewIterator(storekey string, version uint64) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewStartIterator(storekey string, version uint64, start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewEndIterator(storekey string, version uint64, start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewPrefixIterator(storekey string, version uint64, prefix []byte) store.Iterator {
	panic("not implemented")
}
