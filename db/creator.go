package db

import (
	"fmt"
	"strings"
)

type BackendType string

// These are valid backend types.
const (
	// MemDBBackend represents in-memory key value store, which is mostly used
	// for testing.
	MemDBBackend BackendType = "memdb"
	// RocksDBBackend represents rocksdb (uses github.com/cosmos/gorocksdb)
	//   - EXPERIMENTAL
	//   - requires gcc
	//   - use rocksdb build tag (go build -tags rocksdb)
	RocksDBBackend BackendType = "rocksdb"
	// BadgerDBBackend represents BadgerDB
	//   - pure Go
	//   - requires badgerdb build tag
	BadgerDBBackend BackendType = "badgerdb"
)

type DBCreator func(name string, dir string) (DBConnection, error)

var backends = map[BackendType]DBCreator{}

func RegisterCreator(backend BackendType, creator DBCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

// NewDB creates a new database of type backend with the given name.
func NewDB(name string, backend BackendType, dir string) (DBConnection, error) {
	creator, ok := backends[backend]
	if !ok {
		keys := make([]string, 0, len(backends))
		for k := range backends {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("unknown db_backend %s, expected one of %v",
			backend, strings.Join(keys, ","))
	}

	db, err := creator(name, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return db, nil
}
