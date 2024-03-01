//go:build rocksdb
// +build rocksdb

package db

import (
	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

var _ store.RawDB = (*RocksDB)(nil)

// RocksDB implements RawDB using RocksDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct {
	storage *grocksdb.DB
}
