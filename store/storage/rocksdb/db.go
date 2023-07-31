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
	db.db.Close()
	return nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	panic("not implemented!")
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	panic("not implemented!")
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	panic("not implemented!")
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	panic("not implemented!")
}

func (db *Database) GetLatestVersion() (uint64, error) {
	panic("not implemented!")
}

func (db *Database) NewIterator(storeKey string, version uint64) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewStartIterator(storeKey string, version uint64, start []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewEndIterator(storeKey string, version uint64, start []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewPrefixIterator(storeKey string, version uint64, prefix []byte) store.Iterator {
	panic("not implemented!")
}

func (db *Database) NewBatch(version uint64) store.Batch {
	panic("not implemented!")
}
