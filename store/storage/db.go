package storage

import (
	"sync"

	"cosmossdk.io/store/v2"
)

var (
	_ store.Database = (*Database)(nil)
	_ store.Batch    = (*Database)(nil)
)

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
