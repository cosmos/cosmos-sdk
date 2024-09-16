package db

import (
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

	case DBTypeRocksDB:
		return NewRocksDB(name, dataDir)
	}

	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
