package db

import (
	"errors"
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/pebble"
)

var _ store.RawDB = (*PebbleDB)(nil)

// PebbleDB implements RawDB using PebbleDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use PebbleDB as
// the IAVL v0/v1 backend.
type PebbleDB struct {
	storage *pebble.DB

	// Sync is whether to sync writes through the OS buffer cache and down onto
	// the actual disk, if applicable. Setting Sync is required for durability of
	// individual write operations but can result in slower writes.
	//
	// If false, and the process or machine crashes, then a recent write may be
	// lost. This is due to the recently written data being buffered inside the
	// process running Pebble. This differs from the semantics of a write system
	// call in which the data is buffered in the OS buffer cache and would thus
	// survive a process crash.
	sync bool
}

func NewPebbleDB(dataDir string, sync bool) (*PebbleDB, error) {
	opts := &pebble.Options{}
	opts = opts.EnsureDefaults()

	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &PebbleDB{storage: db, sync: sync}, nil
}

func NewPebbleDBWithOpts(dataDir string, sync bool, opts *pebble.Options) (*PebbleDB, error) {
	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &PebbleDB{storage: db, sync: sync}, nil
}

func (db *PebbleDB) Close() error {
	err := db.storage.Close()
	db.storage = nil
	return err
}

func (db *PebbleDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, store.ErrKeyEmpty
	}

	bz, closer, err := db.storage.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			// in case of a fresh database
			return nil, nil
		}

		return nil, fmt.Errorf("failed to perform PebbleDB read: %w", err)
	}

	if len(bz) == 0 {
		return nil, closer.Close()
	}

	return bz, closer.Close()
}

func (db *PebbleDB) Has(key []byte) (bool, error) {
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}

	return bz != nil, nil
}

func (db *PebbleDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *PebbleDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not implemented!")
}

func (db *PebbleDB) NewBatch() store.RawBatch {
	panic("not implemented!")
}

func (db *PebbleDB) NewBatchWithSize(int) store.RawBatch {
	panic("not implemented!")
}
