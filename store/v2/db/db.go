package db

import (
	"fmt"

	coreserver "cosmossdk.io/core/server"
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

func NewDB(dbType DBType, name, dataDir string, opts coreserver.DynamicConfig) (corestore.KVStoreWithBatch, error) {
	switch dbType {
	case DBTypeGoLevelDB:
		return NewGoLevelDB(name, dataDir, opts)

	case DBTypePebbleDB:
		return NewPebbleDB(name, dataDir)
	}

	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
