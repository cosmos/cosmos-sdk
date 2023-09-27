package pebbledb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/cockroachdb/pebble"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage/util"
)

const (
	VersionSize = 8

	StorePrefixTpl   = "s/k:%s/"   // s/k:<storeKey>
	latestVersionKey = "s/_latest" // NB: latestVersionKey key must be lexically smaller than StorePrefixTpl
	tombstoneVal     = "TOMBSTONE"
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
	err := db.storage.Close()
	db.storage = nil
	return err
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
	val, err := db.Get(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (db *Database) Get(storeKey string, targetVersion uint64, key []byte) ([]byte, error) {
	prefixedVal, err := getMVCCSlice(db.storage, storeKey, key, targetVersion)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to perform PebbleDB read: %w", err)
	}

	valBz, tombBz, ok := SplitMVCCKey(prefixedVal)
	if !ok {
		return nil, fmt.Errorf("invalid PebbleDB MVCC value: %s", prefixedVal)
	}

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if len(tombBz) == 0 {
		return valBz, nil
	}

	tombstone, err := decodeUint64Ascending(tombBz)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value tombstone: %w", err)
	}
	if tombstone > targetVersion {
		return nil, fmt.Errorf("value tombstone too large: %d", tombstone)
	}

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if targetVersion < tombstone {
		return valBz, nil
	}

	// the value is considered deleted
	return nil, nil
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	b, err := db.NewBatch(version)
	if err != nil {
		return err
	}

	if err := b.Set(storeKey, key, value); err != nil {
		return err
	}

	return b.Write()
}

// Delete marks the key as deleted. The key will not be retrievable at or before
// the provided version. However, the key is still persisted in the underlying
// database engine as it's value is marked with a non-zero tombestone.
func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	b, err := db.NewBatch(version)
	if err != nil {
		return err
	}

	if err := b.Delete(storeKey, key); err != nil {
		return err
	}

	return b.Write()
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	return NewBatch(db.storage, version)
}

// Prune for the PebbleDB SS backend is currently not supported. It seems the only
// reliable way to prune is to iterate over the desired domain and either manually
// tombstone or delete. Either way, the operation would be timely.
//
// See: https://github.com/cockroachdb/cockroach/blob/33623e3ee420174a4fd3226d1284b03f0e3caaac/pkg/storage/mvcc.go#L3182
func (db *Database) Prune(version uint64) error {
	panic("not implemented!")
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	lowerBound := MVCCEncode(prependStoreKey(storeKey, start), 0)

	var upperBound []byte
	if end != nil {
		upperBound = MVCCEncode(prependStoreKey(storeKey, end), 0)
	}

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: lowerBound, UpperBound: upperBound})
	if err != nil {
		return nil, fmt.Errorf("failed to create PebbleDB iterator: %w", err)
	}

	return newPebbleDBIterator(itr, storePrefix(storeKey), start, end, version), nil
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	panic("not implemented!")
}

func storePrefix(storeKey string) []byte {
	return []byte(fmt.Sprintf(StorePrefixTpl, storeKey))
}

func prependStoreKey(storeKey string, key []byte) []byte {
	return append(storePrefix(storeKey), key...)
}

func getMVCCSlice(db *pebble.DB, storeKey string, key []byte, version uint64) ([]byte, error) {
	// end domain is exclusive, so we need to increment the version by 1
	if version < math.MaxUint64 {
		version++
	}

	itr, err := db.NewIter(&pebble.IterOptions{
		LowerBound: MVCCEncode(prependStoreKey(storeKey, key), 0),
		UpperBound: MVCCEncode(prependStoreKey(storeKey, key), version),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create PebbleDB iterator: %w", err)
	}
	defer func() {
		err = errors.Join(err, itr.Close())
	}()

	if !itr.Last() {
		return nil, store.ErrRecordNotFound
	}

	_, vBz, ok := SplitMVCCKey(itr.Key())
	if !ok {
		return nil, fmt.Errorf("invalid PebbleDB MVCC key: %s", itr.Key())
	}

	keyVersion, err := decodeUint64Ascending(vBz)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key version: %w", err)
	}
	if keyVersion > version {
		return nil, fmt.Errorf("key version too large: %d", keyVersion)
	}

	return util.CopyBytes(itr.Value()), nil
}
