package storage

import (
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
	_ store.VersionedDatabase      = (*StorageStore)(nil)
	_ snapshots.StorageSnapshotter = (*StorageStore)(nil)
	_ store.Pruner                 = (*StorageStore)(nil)
)

// StorageStore is a wrapper around the store.VersionedDatabase interface.
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
func (ss *StorageStore) ApplyChangeset(version uint64, cs *corestore.Changeset) error {
	b, err := ss.db.NewBatch(version)
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
	fmt.Println("ss.db.NewBatch", b, err)
	if err != nil {
		return err
	}

	for kvPair := range chStorage {
		fmt.Println("kvPair loop", string(kvPair.Actor))
		for _, kv := range kvPair.StateChanges {
			fmt.Println("kv loop", string(kv.Key))
			err := b.Set(kvPair.Actor, kv.Key, kv.Value)
			fmt.Println("b.Set err", err)
			if err != nil {
				return err
			}
			if b.Size() > defaultBatchBufferSize {
				err := b.Write()
				fmt.Println("b.Write() err", err)
				if err != nil {
					return err
				}
				err = b.Reset() 
				fmt.Println("b.Reset() err", err)
				if err != nil {
					return err
				}
			}
		}
	}

	fmt.Println("b.Size()", b.Size())
	if b.Size() > 0 {
		err := b.Write()
		fmt.Println("b.Write()", err)
		if err != nil {
			return err
		}
		// if err := b.Write(); err != nil {
		// 	return err
		// }
	}

	v, err := ss.db.GetLatestVersion()

	fmt.Println("storage version after restore", v, err)

	return nil
}

// Close closes the store.
func (ss *StorageStore) Close() error {
	return ss.db.Close()
}
