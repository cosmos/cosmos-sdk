package pebbledb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/cockroachdb/pebble"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	storeerrors "cosmossdk.io/store/v2/errors"
	"cosmossdk.io/store/v2/internal/encoding"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/util"
)

const (
	VersionSize = 8
	// PruneCommitBatchSize defines the size, in number of key/value pairs, to prune
	// in a single batch.
	PruneCommitBatchSize = 50
	// batchBufferSize defines the maximum size of a batch before it is committed.
	batchBufferSize = 100_000

	StorePrefixTpl        = "s/k:%s/"         // s/k:<storeKey>
	removedStoreKeyPrefix = "s/_removed_key"  // NB: removedStoreKeys key must be lexically smaller than StorePrefixTpl
	latestVersionKey      = "s/_latest"       // NB: latestVersionKey key must be lexically smaller than StorePrefixTpl
	pruneHeightKey        = "s/_prune_height" // NB: pruneHeightKey key must be lexically smaller than StorePrefixTpl
	tombstoneVal          = "TOMBSTONE"
)

var (
	_ storage.Database         = (*Database)(nil)
	_ store.UpgradableDatabase = (*Database)(nil)
)

type Database struct {
	storage *pebble.DB

	// earliestVersion defines the earliest version set in the database, which is
	// only updated when the database is pruned.
	earliestVersion uint64

	// Sync is whether to sync writes through the OS buffer cache and down onto
	// the actual disk, if applicable. Setting Sync is required for durability of
	// individual write operations but can result in slower writes.
	//
	// If false, and the process or machine crashes, then a recent write may be
	// lost. This is due to the recently written data being buffered inside the
	// process running Pebble. This differs from the semantics of a write system
	// call in which the data is buffered in the OS buffer cache and would thus
	// survive a process crash.
	sync bool
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

	earliestVersion, err := getEarliestVersion(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get the earliest version: %w", err)
	}

	return &Database{
		storage:         db,
		earliestVersion: earliestVersion,
		sync:            true,
	}, nil
}

func NewWithDB(storage *pebble.DB, sync bool) *Database {
	earliestVersion, err := getEarliestVersion(storage)
	if err != nil {
		panic(fmt.Errorf("failed to get the earliest version: %w", err))
	}

	return &Database{
		storage:         storage,
		earliestVersion: earliestVersion,
		sync:            sync,
	}
}

func (db *Database) SetSync(sync bool) {
	db.sync = sync
}

func (db *Database) Close() error {
	err := db.storage.Close()
	db.storage = nil
	return err
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	b, err := NewBatch(db.storage, version, db.sync)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (db *Database) SetLatestVersion(version uint64) error {
	var ts [VersionSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

	return db.storage.Set([]byte(latestVersionKey), ts[:], &pebble.WriteOptions{Sync: db.sync})
}

func (db *Database) GetLatestVersion() (uint64, error) {
	bz, closer, err := db.storage.Get([]byte(latestVersionKey))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			// in case of a fresh database
			return 0, nil
		}

		return 0, err
	}

	if len(bz) == 0 {
		return 0, closer.Close()
	}

	return binary.LittleEndian.Uint64(bz), closer.Close()
}

func (db *Database) VersionExists(version uint64) (bool, error) {
	latestVersion, err := db.GetLatestVersion()
	if err != nil {
		return false, err
	}

	return latestVersion >= version && version >= db.earliestVersion, nil
}

func (db *Database) setPruneHeight(pruneVersion uint64) error {
	db.earliestVersion = pruneVersion + 1

	var ts [VersionSize]byte
	binary.LittleEndian.PutUint64(ts[:], pruneVersion)

	return db.storage.Set([]byte(pruneHeightKey), ts[:], &pebble.WriteOptions{Sync: db.sync})
}

func (db *Database) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	val, err := db.Get(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (db *Database) Get(storeKey []byte, targetVersion uint64, key []byte) ([]byte, error) {
	if targetVersion < db.earliestVersion {
		return nil, storeerrors.ErrVersionPruned{EarliestVersion: db.earliestVersion, RequestedVersion: targetVersion}
	}

	prefixedVal, err := getMVCCSlice(db.storage, storeKey, key, targetVersion)
	if err != nil {
		if errors.Is(err, storeerrors.ErrRecordNotFound) {
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

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if targetVersion < tombstone {
		return valBz, nil
	}

	// the value is considered deleted
	return nil, nil
}

// Prune removes all versions of all keys that are <= the given version.
//
// Note, the implementation of this method is inefficient and can be potentially
// time consuming given the size of the database and when the last pruning occurred
// (if any). This is because the implementation iterates over all keys in the
// database in order to delete them.
//
// See: https://github.com/cockroachdb/cockroach/blob/33623e3ee420174a4fd3226d1284b03f0e3caaac/pkg/storage/mvcc.go#L3182
func (db *Database) Prune(version uint64) (err error) {
	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: []byte("s/k:")})
	if err != nil {
		return err
	}
	defer itr.Close()

	batch := db.storage.NewBatch()
	defer func() {
		err = errors.Join(err, batch.Close())
	}()

	var (
		batchCounter                              int
		prevKey, prevKeyPrefixed, prevPrefixedVal []byte
		prevKeyVersion                            uint64
	)

	for itr.First(); itr.Valid(); {
		prefixedKey := slices.Clone(itr.Key())

		keyBz, verBz, ok := SplitMVCCKey(prefixedKey)
		if !ok {
			return fmt.Errorf("invalid PebbleDB MVCC key: %s", prefixedKey)
		}

		keyVersion, err := decodeUint64Ascending(verBz)
		if err != nil {
			return fmt.Errorf("failed to decode key version: %w", err)
		}

		// seek to next key if we are at a version which is higher than prune height
		if keyVersion > version {
			itr.NextPrefix()
			continue
		}

		// Delete a key if another entry for that key exists a larger version than
		// the original but <= to the prune height. We also delete a key if it has
		// been tombstoned and its version is <= to the prune height.
		if prevKeyVersion <= version && (bytes.Equal(prevKey, keyBz) || valTombstoned(prevPrefixedVal)) {
			if err := batch.Delete(prevKeyPrefixed, nil); err != nil {
				return err
			}

			batchCounter++
			if batchCounter >= PruneCommitBatchSize {
				if err := batch.Commit(&pebble.WriteOptions{Sync: db.sync}); err != nil {
					return err
				}

				batchCounter = 0
				batch.Reset()
			}
		}

		prevKey = keyBz
		prevKeyVersion = keyVersion
		prevKeyPrefixed = prefixedKey
		value, err := itr.ValueAndErr()
		if err != nil {
			return err
		}
		prevPrefixedVal = slices.Clone(value)

		itr.Next()
	}

	// commit any leftover delete ops in batch
	if batchCounter > 0 {
		if err := batch.Commit(&pebble.WriteOptions{Sync: db.sync}); err != nil {
			return err
		}
	}

	if err := db.deleteRemovedStoreKeys(version); err != nil {
		return err
	}

	return db.setPruneHeight(version)
}

func (db *Database) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, storeerrors.ErrStartAfterEnd
	}

	lowerBound := MVCCEncode(prependStoreKey(storeKey, start), 0)

	var upperBound []byte
	if end != nil {
		upperBound = MVCCEncode(prependStoreKey(storeKey, end), 0)
	}

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: lowerBound, UpperBound: upperBound})
	if err != nil {
		return nil, err
	}

	return newPebbleDBIterator(itr, storePrefix(storeKey), start, end, version, db.earliestVersion, false), nil
}

func (db *Database) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, storeerrors.ErrStartAfterEnd
	}

	lowerBound := MVCCEncode(prependStoreKey(storeKey, start), 0)

	var upperBound []byte
	if end != nil {
		upperBound = MVCCEncode(prependStoreKey(storeKey, end), 0)
	}

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: lowerBound, UpperBound: upperBound})
	if err != nil {
		return nil, err
	}

	return newPebbleDBIterator(itr, storePrefix(storeKey), start, end, version, db.earliestVersion, true), nil
}

func (db *Database) PruneStoreKeys(storeKeys []string, version uint64) (err error) {
	batch := db.storage.NewBatch()
	defer func() {
		err = errors.Join(err, batch.Close())
	}()

	for _, storeKey := range storeKeys {
		if err := batch.Set([]byte(fmt.Sprintf("%s%s", encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version), storeKey)), []byte{}, nil); err != nil {
			return err
		}
	}

	return batch.Commit(&pebble.WriteOptions{Sync: db.sync})
}

func storePrefix(storeKey []byte) []byte {
	return []byte(fmt.Sprintf(StorePrefixTpl, storeKey))
}

func prependStoreKey(storeKey, key []byte) []byte {
	return []byte(fmt.Sprintf("%s%s", storePrefix(storeKey), key))
}

// getEarliestVersion returns the earliest version set in the database.
// It is calculated by prune height + 1. If the prune height is not set, it
// returns 0.
func getEarliestVersion(storage *pebble.DB) (uint64, error) {
	bz, closer, err := storage.Get([]byte(pruneHeightKey))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			// in cases where pruning was never triggered
			return 0, nil
		}

		return 0, err
	}

	if len(bz) == 0 {
		return 0, closer.Close()
	}

	return binary.LittleEndian.Uint64(bz) + 1, closer.Close()
}

func valTombstoned(value []byte) bool {
	if value == nil {
		return false
	}

	_, tombBz, ok := SplitMVCCKey(value)
	if !ok {
		// XXX: This should not happen as that would indicate we have a malformed
		// MVCC value.
		panic(fmt.Sprintf("invalid PebbleDB MVCC value: %s", value))
	}

	// If the tombstone suffix is empty, we consider this a zero value and thus it
	// is not tombstoned.
	if len(tombBz) == 0 {
		return false
	}

	return true
}

func getMVCCSlice(db *pebble.DB, storeKey, key []byte, version uint64) ([]byte, error) {
	// end domain is exclusive, so we need to increment the version by 1
	if version < math.MaxUint64 {
		version++
	}

	itr, err := db.NewIter(&pebble.IterOptions{
		LowerBound: MVCCEncode(prependStoreKey(storeKey, key), 0),
		UpperBound: MVCCEncode(prependStoreKey(storeKey, key), version),
	})
	if err != nil {
		return nil, err
	}

	defer itr.Close()

	if !itr.Last() {
		return nil, storeerrors.ErrRecordNotFound
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

	value, err := itr.ValueAndErr()
	return slices.Clone(value), err
}

func (db *Database) deleteRemovedStoreKeys(version uint64) (err error) {
	batch := db.storage.NewBatch()
	defer func() {
		err = errors.Join(err, batch.Close())
	}()

	end := encoding.BuildPrefixWithVersion(removedStoreKeyPrefix, version+1)
	storeKeyIter, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: []byte(removedStoreKeyPrefix), UpperBound: end})
	if err != nil {
		return err
	}
	defer storeKeyIter.Close()

	storeKeys := make(map[string]uint64)
	prefixLen := len(end)
	for storeKeyIter.First(); storeKeyIter.Valid(); storeKeyIter.Next() {
		verBz := storeKeyIter.Key()[len(removedStoreKeyPrefix):prefixLen]
		v, err := decodeUint64Ascending(verBz)
		if err != nil {
			return err
		}
		storeKey := string(storeKeyIter.Key()[prefixLen:])
		if ev, ok := storeKeys[storeKey]; ok {
			if ev < v {
				storeKeys[storeKey] = v
			}
		} else {
			storeKeys[storeKey] = v
		}
		if err := batch.Delete(storeKeyIter.Key(), nil); err != nil {
			return err
		}
	}

	for storeKey, v := range storeKeys {
		if err := func() error {
			storeKey := []byte(storeKey)
			itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: storePrefix(storeKey), UpperBound: storePrefix(util.CopyIncr(storeKey))})
			if err != nil {
				return err
			}
			defer itr.Close()

			for itr.First(); itr.Valid(); itr.Next() {
				itrKey := itr.Key()
				_, verBz, ok := SplitMVCCKey(itrKey)
				if !ok {
					return fmt.Errorf("invalid PebbleDB MVCC key: %s", itrKey)
				}
				keyVersion, err := decodeUint64Ascending(verBz)
				if err != nil {
					return err
				}
				if keyVersion > v {
					// skip keys that are newer than the version
					continue
				}
				if err := batch.Delete(itr.Key(), nil); err != nil {
					return err
				}
				if batch.Len() >= batchBufferSize {
					if err := batch.Commit(&pebble.WriteOptions{Sync: db.sync}); err != nil {
						return err
					}
					batch.Reset()
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	return batch.Commit(&pebble.WriteOptions{Sync: true})
}
