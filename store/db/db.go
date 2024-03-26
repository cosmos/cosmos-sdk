package db

import (
	"fmt"

	"cosmossdk.io/store/v2"
)

type RawDBType string

const (
	DBTypeGoLevelDB RawDBType = "goleveldb"
	DBTypeRocksDB             = "rocksdb"
	DBTypePebbleDB            = "pebbledb"
	DBTypePrefixDB            = "prefixdb"

	DBFileSuffix string = ".db"
)

func NewRawDB(dbType RawDBType, name, dataDir string, opts store.DBOptions) (store.RawDB, error) {
	switch dbType {
	case DBTypeGoLevelDB:
		return NewGoLevelDB(name, dataDir, opts)

	case DBTypeRocksDB:
		return NewRocksDB(name, dataDir)

	case DBTypePebbleDB:
		return NewPebbleDB(name, dataDir)
	}

	return nil, fmt.Errorf("unsupported db type: %s", dbType)
}
