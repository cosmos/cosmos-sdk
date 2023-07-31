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
	db      store.VersionedDatabase
	version uint64
	mem     map[string]dbEntry
	batch   store.Batch
}

// dbEntry defines a database entry stored in memory until the batch is written
// on commit.
type dbEntry struct {
	value   []byte
	delete  bool
	version uint64
}

func New(db store.VersionedDatabase, version uint64, memSize int) *Database {
	return &Database{
		db:      db,
		version: version,
		mem:     make(map[string]dbEntry, memSize),
		batch:   db.NewBatch(version),
	}
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.db.Close(); err != nil {
		return err
	}

	db.db = nil
	db.mem = nil
	db.batch = nil

	return nil
}

func (db *Database) Has(version uint64, key []byte) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.mem == nil {
		return false, store.ErrClosed
	}

	if entry, ok := db.mem[string(key)]; ok && entry.version == version {
		return !entry.delete, nil
	}

	return db.db.Has(version, key)
}

func (db *Database) Get(version uint64, key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.mem == nil {
		return nil, store.ErrClosed
	}

	if entry, ok := db.mem[string(key)]; ok && entry.version == version {
		if entry.delete {
			return nil, store.ErrRecordNotFound
		}

		return slices.Clone(entry.value), nil
	}

	return db.db.Get(version, key)
}

func (db *Database) Set(version uint64, key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.mem == nil {
		return store.ErrClosed
	}
	if version != db.version {
		return fmt.Errorf("invalid version; expected %d, got %d", db.version, version)
	}

	db.mem[string(key)] = dbEntry{value: slices.Clone(value), version: version}
	return nil
}

func (db *Database) Delete(version uint64, key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.mem == nil {
		return store.ErrClosed
	}
	if version != db.version {
		return fmt.Errorf("invalid version; expected %d, got %d", db.version, version)
	}

	db.mem[string(key)] = dbEntry{delete: true, version: version}

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

	for key, entry := range db.mem {
		if entry.delete {
			if err := db.batch.Delete([]byte(key)); err != nil {
				return err
			}
		} else if err := db.batch.Set([]byte(key), entry.value); err != nil {
			return err
		}
	}

	if err := db.batch.Write(); err != nil {
		return fmt.Errorf("failed to write state storage database batch: %w", err)
	}

	db.batch.Reset()

	return nil
}

func (db *Database) NewIterator(version uint64) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewStartIterator(version uint64, start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewEndIterator(version uint64, start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewPrefixIterator(version uint64, prefix []byte) store.Iterator {
	panic("not implemented")
}
