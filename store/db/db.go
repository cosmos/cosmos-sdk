package db

import (
	"errors"
	"fmt"

	corestore "cosmossdk.io/core/store"
)

type DBType string

const (
	DBTypeGoLevelDB DBType = "goleveldb"
	DBTypeRocksDB   DBType = "rocksdb"
	DBTypePebbleDB  DBType = "pebbledb"
	DBTypePrefixDB  DBType = "prefixdb"

	DBFileSuffix string = ".db"
)

var (
	ErrKeyEmpty = errors.New("key empty")
	// ErrBatchClosed is returned when a closed or written batch is used.
	ErrBatchClosed = errors.New("batch has been written or closed")

	// ErrValueNil is returned when attempting to set a nil value.
	ErrValueNil = errors.New("value nil")
)

// DBOptions defines the interface of a database options.
type DBOptions interface {
	Get(string) interface{}
}

func NewDB(name string, dbType DBType, dataDir string, opts DBOptions) (corestore.KVStoreWithBatch, error) {
	switch dbType {
	case DBTypeGoLevelDB:
		return NewGoLevelDB(name, dataDir, opts)

	case DBTypePebbleDB:
		return NewPebbleDB(name, dataDir)
	}

	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
