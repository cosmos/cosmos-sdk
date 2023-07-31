package rocksdb

import (
	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

const (
	TimestampSize = 8

	StorePrefixTpl   = "s/k:%s/"
	latestVersionKey = "s/latest"
)

// TODO: STORE KEYS

var _ store.VersionedDatabase = (*Database)(nil)

type Database struct {
	db       *grocksdb.DB
	cfHandle *grocksdb.ColumnFamilyHandle
}

func (db *Database) Close() error {
	panic("not implemented!")
}

func (db *Database) Has(version uint64, key []byte) (bool, error) {
	panic("not implemented!")
}

func (db *Database) Get(version uint64, key []byte) ([]byte, error) {
	panic("not implemented!")
}

func (db *Database) Set(version uint64, key, value []byte) error {
	panic("not implemented!")
}

func (db *Database) Delete(version uint64, key []byte) error {
	panic("not implemented!")
}

func (db *Database) GetLatestVersion() (uint64, error) {
	panic("not implemented!")
}

func (db *Database) NewIterator(version uint64) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewStartIterator(version uint64, start []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewEndIterator(version uint64, start []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewPrefixIterator(version uint64, prefix []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewBatch(version uint64) store.Batch {
	panic("not implemented!")
}
