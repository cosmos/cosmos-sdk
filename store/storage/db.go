package storage

import (
	"sync"

	"cosmossdk.io/store/v2"
)

var _ store.Database = (*Database)(nil)

// Database defines the state storage (SS) backend.
type Database struct {
	mu    sync.RWMutex
	db    store.Database
	batch store.Batch
}

func New(db store.Database) *Database {
	return &Database{
		db:    db,
		batch: db.NewBatch(),
	}
}

func (db *Database) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.db.Close(); err != nil {
		return err
	}

	db.db = nil
	db.batch = nil

	return nil
}

func (db *Database) Has(key []byte) (bool, error) {
	panic("not implemented")
}

func (db *Database) Get(key []byte) ([]byte, error) {
	panic("not implemented")
}

func (db *Database) Set(key, value []byte) error {
	panic("not implemented")
}

func (db *Database) Delete(key []byte) error {
	panic("not implemented")
}

func (db *Database) NewBatch() store.Batch {
	panic("not implemented")
}

func (db *Database) NewIterator() store.Iterator {
	panic("not implemented")
}

func (db *Database) NewStartIterator(start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewEndIterator(start []byte) store.Iterator {
	panic("not implemented")
}

func (db *Database) NewPrefixIterator(prefix []byte) store.Iterator {
	panic("not implemented")
}
