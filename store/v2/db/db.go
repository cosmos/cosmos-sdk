package db

import (
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

type DBType string

const (
	DBTypeGoLevelDB DBType = "goleveldb"
	DBTypeRocksDB   DBType = "rocksdb"
	DBTypePebbleDB  DBType = "pebbledb"
	DBTypePrefixDB  DBType = "prefixdb"

	DBFileSuffix string = ".db"
)

func NewDB(dbType DBType, name, dataDir string, opts store.DBOptions) (corestore.KVStoreWithBatch, error) {
	switch dbType {
	case DBTypeGoLevelDB:
		return NewGoLevelDB(name, dataDir, opts)

	case DBTypePebbleDB:
		return NewPebbleDB(name, dataDir)
	}

	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
