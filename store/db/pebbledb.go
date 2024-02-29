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
}

func NewPebbleDB(dataDir string) (*PebbleDB, error) {
	opts := &pebble.Options{}
	opts = opts.EnsureDefaults()

	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &PebbleDB{storage: db}, nil
}

func NewPebbleDBWithOpts(dataDir string, sync bool, opts *pebble.Options) (*PebbleDB, error) {
	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &PebbleDB{storage: db}, nil
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
	return &pebbleDBBatch{
		db:    db,
		batch: db.storage.NewBatch(),
	}
}

func (db *PebbleDB) NewBatchWithSize(size int) store.RawBatch {
	return &pebbleDBBatch{
		db:    db,
		batch: db.storage.NewBatchWithSize(size),
	}
}

var _ store.RawBatch = (*pebbleDBBatch)(nil)

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

func (b *pebbleDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if value == nil {
		return store.ErrValueNil
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}

	return b.batch.Set(key, value, nil)
}

func (b *pebbleDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}

	return b.batch.Delete(key, nil)
}

func (b *pebbleDBBatch) Write() error {
	err := b.batch.Commit(&pebble.WriteOptions{Sync: false})
	if err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return nil
}

func (b *pebbleDBBatch) WriteSync() error {
	err := b.batch.Commit(&pebble.WriteOptions{Sync: true})
	if err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return nil
}

func (b *pebbleDBBatch) Close() error {
	return b.batch.Close()
}

func (b *pebbleDBBatch) GetByteSize() (int, error) {
	return b.batch.Len(), nil
}
