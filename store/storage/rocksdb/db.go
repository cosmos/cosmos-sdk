//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage/util"
	"github.com/linxGnu/grocksdb"
	"golang.org/x/exp/slices"
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

	// tsLow reflects the full_history_ts_low CF value. Since pruning is done in
	// a lazy manner, we use this value to prevent reads for versions that will
	// be purged in the next compaction.
	tsLow uint64
}

func New(dataDir string) (*Database, error) {
	storage, cfHandle, err := OpenRocksDB(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	slice, err := storage.GetFullHistoryTsLow(cfHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to get full_history_ts_low: %w", err)
	}

	var tsLow uint64
	tsLowBz := copyAndFreeSlice(slice)
	if len(tsLowBz) > 0 {
		tsLow = binary.LittleEndian.Uint64(tsLowBz)
	}

	return &Database{
		storage:  storage,
		cfHandle: cfHandle,
		tsLow:    tsLow,
	}, nil
}

func NewWithDB(storage *grocksdb.DB, cfHandle *grocksdb.ColumnFamilyHandle) (*Database, error) {
	slice, err := storage.GetFullHistoryTsLow(cfHandle)
	if err != nil {
		return nil, fmt.Errorf("failed to get full_history_ts_low: %w", err)
	}

	var tsLow uint64
	tsLowBz := copyAndFreeSlice(slice)
	if len(tsLowBz) > 0 {
		tsLow = binary.LittleEndian.Uint64(tsLowBz)
	}
	return &Database{
		storage:  storage,
		cfHandle: cfHandle,
		tsLow:    tsLow,
	}, nil
}

func (db *Database) Close() error {
	db.storage.Close()

	db.storage = nil
	db.cfHandle = nil

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
		return 0, store.ErrVersionNotFound
	}

	return binary.LittleEndian.Uint64(bz), nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	if version < db.tsLow {
		return false, nil
	}

	slice, err := db.getSlice(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return slice.Exists(), nil
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	if version < db.tsLow {
		return nil, nil
	}

	slice, err := db.getSlice(storeKey, version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get RocksDB slice: %w", err)
	}

	return copyAndFreeSlice(slice), nil
}

func (db *Database) ApplyChangeset(version uint64, cs *store.Changeset) error {
	b := NewBatch(db, version)

	for _, kvPair := range cs.Pairs {
		if kvPair.Value == nil {
			b.Delete(kvPair.StoreKey, kvPair.Key)
		} else {
			b.Set(kvPair.StoreKey, kvPair.Key, kvPair.Value)
		}
	}

	return b.Write()
}

// Prune attempts to prune all versions up to and including the provided version.
// This is done internally by updating the full_history_ts_low RocksDB value on
// the column families, s.t. all versions less than full_history_ts_low will be
// dropped.
//
// Note, this does NOT incur an immediate full compaction, i.e. this performs a
// lazy prune. Future compactions will honor the increased full_history_ts_low
// and trim history when possible.
func (db *Database) Prune(version uint64) error {
	tsLow := version + 1 // we increment by 1 to include the provided version

	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], tsLow)

	if err := db.storage.IncreaseFullHistoryTsLow(db.cfHandle, ts[:]); err != nil {
		return fmt.Errorf("failed to update column family full_history_ts_low: %w", err)
	}

	db.tsLow = tsLow
	return nil
}

func (db *Database) Iterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	prefix := storePrefix(storeKey)
	start, end = util.IterateWithPrefix(prefix, start, end)

	itr := db.storage.NewIteratorCF(newTSReadOptions(version), db.cfHandle)
	return newRocksDBIterator(itr, prefix, start, end, false), nil
}

func (db *Database) ReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	prefix := storePrefix(storeKey)
	start, end = util.IterateWithPrefix(prefix, start, end)

	itr := db.storage.NewIteratorCF(newTSReadOptions(version), db.cfHandle)
	return newRocksDBIterator(itr, prefix, start, end, true), nil
}

// newTSReadOptions returns ReadOptions used in the RocksDB column family read.
func newTSReadOptions(version uint64) *grocksdb.ReadOptions {
	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

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

	return slices.Clone(s.Data())
}

func readOnlySlice(s *grocksdb.Slice) []byte {
	if !s.Exists() {
		return nil
	}

	return s.Data()
}
