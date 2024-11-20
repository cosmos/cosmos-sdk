package storage

import (
	"errors"
	"fmt"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
)

const (
	// TODO: it is a random number, need to be tuned
	defaultBatchBufferSize = 100000
)

var (
	_ store.VersionedWriter        = (*StorageStore)(nil)
	_ snapshots.StorageSnapshotter = (*StorageStore)(nil)
	_ store.Pruner                 = (*StorageStore)(nil)
	_ store.UpgradableDatabase     = (*StorageStore)(nil)
)

// StorageStore is a wrapper around the store.VersionedWriter interface.
type StorageStore struct {
	logger log.Logger
	db     Database
}

// NewStorageStore returns a reference to a new StorageStore.
func NewStorageStore(db Database, logger log.Logger) *StorageStore {
	return &StorageStore{
		logger: logger,
		db:     db,
	}
}

// Has returns true if the key exists in the store.
func (ss *StorageStore) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	return ss.db.Has(storeKey, version, key)
}

// Get returns the value associated with the given key.
func (ss *StorageStore) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	return ss.db.Get(storeKey, version, key)
}

// ApplyChangeset applies the given changeset to the storage.
func (ss *StorageStore) ApplyChangeset(cs *corestore.Changeset) error {
	b, err := ss.db.NewBatch(cs.Version)
	if err != nil {
		return err
	}

	for _, pairs := range cs.Changes {
		for _, kvPair := range pairs.StateChanges {
			if kvPair.Remove {
				if err := b.Delete(pairs.Actor, kvPair.Key); err != nil {
					return err
				}
			} else {
				if err := b.Set(pairs.Actor, kvPair.Key, kvPair.Value); err != nil {
					return err
				}
			}
		}
	}

	if err := b.Write(); err != nil {
		return err
	}

	return nil
}

// GetLatestVersion returns the latest version of the store.
func (ss *StorageStore) GetLatestVersion() (uint64, error) {
	return ss.db.GetLatestVersion()
}

// SetLatestVersion sets the latest version of the store.
func (ss *StorageStore) SetLatestVersion(version uint64) error {
	return ss.db.SetLatestVersion(version)
}

// VersionExists returns true if the given version exists in the store.
func (ss *StorageStore) VersionExists(version uint64) (bool, error) {
	return ss.db.VersionExists(version)
}

// Iterator returns an iterator over the specified domain and prefix.
func (ss *StorageStore) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	return ss.db.Iterator(storeKey, version, start, end)
}

// ReverseIterator returns an iterator over the specified domain and prefix in reverse.
func (ss *StorageStore) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	return ss.db.ReverseIterator(storeKey, version, start, end)
}

// Prune prunes the store up to the given version.
func (ss *StorageStore) Prune(version uint64) error {
	return ss.db.Prune(version)
}

// Restore restores the store from the given channel.
func (ss *StorageStore) Restore(version uint64, chStorage <-chan *corestore.StateChanges) error {
	latestVersion, err := ss.db.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}
	if version <= latestVersion {
		return fmt.Errorf("the snapshot version %d is not greater than latest version %d", version, latestVersion)
	}

	b, err := ss.db.NewBatch(version)
	if err != nil {
		return err
	}

	for kvPair := range chStorage {
		for _, kv := range kvPair.StateChanges {
			if err := b.Set(kvPair.Actor, kv.Key, kv.Value); err != nil {
				return err
			}
			if b.Size() > defaultBatchBufferSize {
				if err := b.Write(); err != nil {
					return err
				}
				if err := b.Reset(); err != nil {
					return err
				}
			}
		}
	}

	if b.Size() > 0 {
		if err := b.Write(); err != nil {
			return err
		}
	}

	return nil
}

// PruneStoreKeys prunes the store keys which implements the store.UpgradableDatabase
// interface.
func (ss *StorageStore) PruneStoreKeys(storeKeys []string, version uint64) error {
	gdb, ok := ss.db.(store.UpgradableDatabase)
	if !ok {
		return errors.New("db does not implement UpgradableDatabase interface")
	}

	return gdb.PruneStoreKeys(storeKeys, version)
}

// Close closes the store.
func (ss *StorageStore) Close() error {
	return ss.db.Close()
}
