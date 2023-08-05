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
	storage, cfHandle, err := OpenRocksDB(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	return &Database{
		storage:  storage,
		cfHandle: cfHandle,
	}, nil
}

func NewWithDB(storage *grocksdb.DB, cfHandle *grocksdb.ColumnFamilyHandle) *Database {
	return &Database{
		storage:  storage,
		cfHandle: cfHandle,
	}
}

func (db *Database) Close() error {
	db.storage.Close()
	return nil
}

func (db *Database) getSlice(storeKey string, version uint64, key []byte) (*grocksdb.Slice, error) {
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

func (db *Database) GetLatestVersion() (uint64, error) {
	bz, err := db.storage.GetBytes(defaultReadOpts, []byte(latestVersionKey))
	if err != nil {
		return 0, err
	}

	if len(bz) == 0 {
		return 0, nil
	}

	return binary.LittleEndian.Uint64(bz), nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	slice, err := db.getSlice(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return slice.Exists(), nil
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	slice, err := db.getSlice(storeKey, version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RocksDB slice: %w", err)
	}

	return copyAndFreeSlice(slice), nil
}

// Set will store a given key/value pair in addition to setting the provided version
// as the latest version. Note, a caller should prefer to create a batch instead
// of calling Set() directly.
func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	batch.Put([]byte(latestVersionKey), ts[:])

	prefixedKey := prependStoreKey(storeKey, key)
	batch.PutCFWithTS(db.cfHandle, prefixedKey, ts[:], value)

	return db.storage.Write(defaultWriteOpts, batch)
}

// Delete will remove a given key/value pair in addition to setting the provided
// version as the latest version. Note, a caller should prefer to create a batch
// instead of calling Delete() directly.
func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

	batch := grocksdb.NewWriteBatch()
	defer batch.Destroy()

	batch.Put([]byte(latestVersionKey), ts[:])

	prefixedKey := prependStoreKey(storeKey, key)
	batch.DeleteCFWithTS(db.cfHandle, prefixedKey, ts[:])

	return db.storage.Write(defaultWriteOpts, batch)
}

func (db *Database) NewBatch(version uint64) store.Batch {
	return NewBatch(db, version)
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	prefix := storePrefix(storeKey)
	start, end = iterateWithPrefix(prefix, start, end)

	itr := db.storage.NewIteratorCF(newTSReadOptions(version), db.cfHandle)
	return newRocksDBIterator(itr, prefix, start, end, false), nil
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	prefix := storePrefix(storeKey)
	start, end = iterateWithPrefix(prefix, start, end)

	itr := db.storage.NewIteratorCF(newTSReadOptions(version), db.cfHandle)
	return newRocksDBIterator(itr, prefix, start, end, true), nil
}

// newTSReadOptions returns ReadOptions used in the RocksDB column family read.
// Note, a zero version indicates a maximum version.
func newTSReadOptions(version uint64) *grocksdb.ReadOptions {
	var ver uint64
	if version == 0 {
		ver = math.MaxUint64
	} else {
		ver = version
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

func cloneAppend(bz []byte, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))

	copy(res, bz)
	copy(res[len(bz):], tail)

	return res
}

func copyIncr(bz []byte) []byte {
	if len(bz) == 0 {
		panic("copyIncr expects non-zero bz length")
	}

	ret := make([]byte, len(bz))
	copy(ret, bz)

	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] < byte(0xFF) {
			ret[i]++
			return ret
		}

		ret[i] = byte(0x00)

		if i == 0 {
			// overflow
			return nil
		}
	}

	return nil
}

func iterateWithPrefix(prefix, begin, end []byte) ([]byte, []byte) {
	if len(prefix) == 0 {
		return begin, end
	}

	begin = cloneAppend(prefix, begin)

	if end == nil {
		end = copyIncr(prefix)
	} else {
		end = cloneAppend(prefix, end)
	}

	return begin, end
}
