package pebbledb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage/util"
	"github.com/cockroachdb/pebble"
)

const (
	VersionSize = 8

	StorePrefixTpl   = "s/k:%d/%s/" // s/k:<version>/<storeKey>/<key>
	latestVersionKey = "s/latest"
)

var (
	_ store.VersionedDatabase = (*Database)(nil)

	defaultWriteOpts = pebble.Sync
)

type Database struct {
	storage *pebble.DB
}

func New(dataDir string) (*Database, error) {
	opts := &pebble.Options{
		Comparer: MVCCComparer,
	}
	opts = opts.EnsureDefaults()

	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &Database{
		storage: db,
	}, nil
}

func NewWithDB(storage *pebble.DB) *Database {
	return &Database{
		storage: storage,
	}
}

func (db *Database) Close() error {
	return db.storage.Close()
}

func (db *Database) SetLatestVersion(version uint64) error {
	var ts [VersionSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)
	return db.storage.Set([]byte(latestVersionKey), ts[:], defaultWriteOpts)
}

func (db *Database) GetLatestVersion() (uint64, error) {
	bz, closer, err := db.storage.Get([]byte(latestVersionKey))
	if err != nil {
		return 0, err
	}

	if len(bz) == 0 {
		return 0, closer.Close()
	}

	return binary.LittleEndian.Uint64(bz), closer.Close()
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	_, closer, err := db.storage.Get(prependStoreKey(storeKey, version, key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to perform PebbleDB read: %w", err)
	}

	return true, closer.Close()
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	bz, closer, err := db.storage.Get(prependStoreKey(storeKey, version, key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to perform PebbleDB read: %w", err)
	}

	bzCopy := make([]byte, len(bz))
	copy(bzCopy, bz)

	return bzCopy, closer.Close()
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	var versionBz [VersionSize]byte
	binary.LittleEndian.PutUint64(versionBz[:], version)

	batch := db.storage.NewBatch()

	if err := batch.Set([]byte(latestVersionKey), versionBz[:], nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}
	if err := batch.Set(prependStoreKey(storeKey, version, key), value, nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	if err := batch.Commit(defaultWriteOpts); err != nil {
		return fmt.Errorf("failed to commit PebbleDB batch: %w", err)
	}

	return nil
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	var versionBz [VersionSize]byte
	binary.LittleEndian.PutUint64(versionBz[:], version)

	batch := db.storage.NewBatch()

	if err := batch.Set([]byte(latestVersionKey), versionBz[:], nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}
	if err := batch.Delete(prependStoreKey(storeKey, version, key), nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	if err := batch.Commit(defaultWriteOpts); err != nil {
		return fmt.Errorf("failed to commit PebbleDB batch: %w", err)
	}

	return nil
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	return NewBatch(db.storage, version)
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	prefix := storePrefix(storeKey, version)
	start, end = util.IterateWithPrefix(prefix, start, end)

	iter := db.storage.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	return newPebbleDBIterator(iter, prefix, start, end, false), nil
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	prefix := storePrefix(storeKey, version)
	start, end = util.IterateWithPrefix(prefix, start, end)

	iter := db.storage.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	return newPebbleDBIterator(iter, prefix, start, end, true), nil
}

func storePrefix(storeKey string, version uint64) []byte {
	return []byte(fmt.Sprintf(StorePrefixTpl, version, storeKey))
}

func prependStoreKey(storeKey string, version uint64, key []byte) []byte {
	return append(storePrefix(storeKey, version), key...)
}
