package storage

import (
	"fmt"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
)

const (
	// TODO: Ensure a value is appropriately chosen.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/18937
	defaultBatchBufferSize = 100000
)

var (
	_ store.VersionedDatabase      = (*StorageStore)(nil)
	_ snapshots.StorageSnapshotter = (*StorageStore)(nil)
)

// StorageStore is a wrapper around the store.VersionedDatabase (SS) interface.
//
// We provide a wrapper that will allow us to abstract various methods that must
// be implemented on each possible SS backend type, e.g. State Sync restoring.
// Instead of implementing these methods on each SS backend, we can implement
// them here and embed the SS backend directly. This means that applications
// should use this type instead of an SS backend directly.
type StorageStore struct {
	db Database
}

// NewStorageStore returns a reference to a new StorageStore.
func NewStorageStore(db Database) *StorageStore {
	return &StorageStore{
		db: db,
	}
}

// Has returns true if the key exists in the store.
func (ss *StorageStore) Has(storeKey string, version uint64, key []byte) (bool, error) {
	return ss.db.Has(storeKey, version, key)
}

// Get returns the value associated with the given key.
func (ss *StorageStore) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	return ss.db.Get(storeKey, version, key)
}

// ApplyChangeset applies the given changeset to the storage.
func (ss *StorageStore) ApplyChangeset(version uint64, cs *store.Changeset) error {
	b, err := ss.db.NewBatch(version)
	if err != nil {
		return err
	}

	for storeKey, pairs := range cs.Pairs {
		for _, kvPair := range pairs {
			if kvPair.Value == nil {
				if err := b.Delete(storeKey, kvPair.Key); err != nil {
					return err
				}
			} else {
				if err := b.Set(storeKey, kvPair.Key, kvPair.Value); err != nil {
					return err
				}
			}
		}
	}

	return b.Write()
}

// GetLatestVersion returns the latest version of the store.
func (ss *StorageStore) GetLatestVersion() (uint64, error) {
	return ss.db.GetLatestVersion()
}

// SetLatestVersion sets the latest version of the store.
func (ss *StorageStore) SetLatestVersion(version uint64) error {
	return ss.db.SetLatestVersion(version)
}

// Iterator returns an iterator over the specified domain and prefix.
func (ss *StorageStore) Iterator(storeKey string, version uint64, start, end []byte) (corestore.Iterator, error) {
	return ss.db.Iterator(storeKey, version, start, end)
}

// ReverseIterator returns an iterator over the specified domain and prefix in reverse.
func (ss *StorageStore) ReverseIterator(storeKey string, version uint64, start, end []byte) (corestore.Iterator, error) {
	return ss.db.ReverseIterator(storeKey, version, start, end)
}

// Prune prunes the store up to the given version.
func (ss *StorageStore) Prune(version uint64) error {
	return ss.db.Prune(version)
}

// Restore restores the store from the given channel.
func (ss *StorageStore) Restore(version uint64, chStorage <-chan *store.KVPair) error {
	latestVersion, err := ss.db.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}
	if version <= latestVersion {
		return fmt.Errorf("the snapshot version %d is not greater than latest version %d", version, latestVersion)
	}

	b, err := ss.db.NewBatch(version)
	if err != nil {
		return fmt.Errorf("failed to create a new batch: %w", err)
	}

	for kvPair := range chStorage {
		if err := b.Set(kvPair.StoreKey, kvPair.Key, kvPair.Value); err != nil {
			return err
		}

		if b.Size() > defaultBatchBufferSize {
			if err := b.Write(); err != nil {
				return err
			}
		}
	}

	// flush the remaining writes
	if b.Size() > 0 {
		if err := b.Write(); err != nil {
			return err
		}
	}

	return ss.db.SetLatestVersion(version)
}

// Close closes the store.
func (ss *StorageStore) Close() error {
	return ss.db.Close()
}
