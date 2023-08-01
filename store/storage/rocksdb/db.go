package rocksdb

import (
	"encoding/binary"
	"fmt"
	"math"

	"cosmossdk.io/store/v2"
	"github.com/linxGnu/grocksdb"
)

const (
	TimestampSize = 8

	StorePrefixTpl   = "s/k:%s/"
	latestVersionKey = "s/latest"
)

var (
	_ store.VersionedDatabase = (*Database)(nil)

	defaultWriteOpts = grocksdb.NewDefaultWriteOptions()
	defaultReadOpts  = grocksdb.NewDefaultReadOptions()
)

type Database struct {
	storage  *grocksdb.DB
	cfHandle *grocksdb.ColumnFamilyHandle
}

func New(dataDir string) (*Database, error) {
	db, cfHandle, err := OpenRocksDB(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	return &Database{
		storage:  db,
		cfHandle: cfHandle,
	}, nil
}

func (db *Database) Close() error {
	db.storage.Close()
	return nil
}

func (db *Database) GetSlice(storeKey string, version uint64, key []byte) (*grocksdb.Slice, error) {
	return db.storage.GetCF(
		newTSReadOptions(version),
		db.cfHandle,
		prependStoreKey(storeKey, key),
	)
}

func (db *Database) SetLatestVersion(version uint64) error {
	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

	return db.storage.Put(defaultWriteOpts, []byte(latestVersionKey), ts[:])
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	slice, err := db.GetSlice(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return slice.Exists(), nil
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	slice, err := db.GetSlice(storeKey, version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RocksDB slice: %w", err)
	}

	return copyAndFreeSlice(slice), nil
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

// newTSReadOptions returns ReadOptions used in the RocksDB column family read.
// Note, a zero version indicates a maximum version.
func newTSReadOptions(version uint64) *grocksdb.ReadOptions {
	var ver uint64
	if version == 0 {
		ver = math.MaxUint64
	} else {
		ver = uint64(version)
	}

	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], ver)

	readOpts := grocksdb.NewDefaultReadOptions()
	readOpts.SetTimestamp(ts[:])

	return readOpts
}

func storePrefix(storeKey string) []byte {
	return []byte(fmt.Sprintf(StorePrefixTpl, storeKey))
}

func prependStoreKey(storeKey string, key []byte) []byte {
	return append(storePrefix(storeKey), key...)
}

// moveSliceToBytes will free the slice and copy out a go []byte
// This function can be applied on *Slice returned from Key() and Value()
// of an Iterator, because they are marked as freed.

// copyAndFreeSlice will copy a given RocksDB slice and free it. If the slice does
// not exist, <nil> will be returned.
func copyAndFreeSlice(s *grocksdb.Slice) []byte {
	defer s.Free()
	if !s.Exists() {
		return nil
	}

	v := make([]byte, len(s.Data()))
	copy(v, s.Data())

	return v
}
