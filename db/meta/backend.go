package meta

import (
	"fmt"
	"strings"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

type BackendType string

// These are valid backend types.
const (
	// MemDBBackend represents in-memory key value store, which is mostly used
	// for testing.
	//   - use memdb build tag (go build -tags memdb)
	MemDBBackend BackendType = "memdb"
	// RocksDBBackend represents rocksdb (uses github.com/tecbot/gorocksdb)
	//   - EXPERIMENTAL
	//   - requires gcc
	//   - use rocksdb build tag (go build -tags rocksdb)
	RocksDBBackend BackendType = "rocksdb"
	// BadgerDBBackend represents badger (uses github.com/dgraph-io/badger/v3)
	//   - EXPERIMENTAL
	//   - use badgerdb build tag (go build -tags badgerdb)
	BadgerDBBackend BackendType = "badgerdb"
)

type dbConstructor func(name string, dir string) (dbm.DBConnection, error)

var backends = map[BackendType]dbConstructor{}

func registerConstructor(backend BackendType, ctor dbConstructor, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = ctor
}

// NewDB creates a new database of type backend with the given name.
func NewDB(name string, backend BackendType, dir string) (dbm.DBConnection, error) {
	dbCreator, ok := backends[backend]
	if !ok {
		keys := make([]string, 0, len(backends))
		for k := range backends {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("unknown DB backend %s, expected one of %v",
			backend, strings.Join(keys, ","))
	}

	db, err := dbCreator(name, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return db, nil
}
