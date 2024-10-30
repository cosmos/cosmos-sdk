package runtime

import (
	"errors"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf"
	storev2 "cosmossdk.io/store/v2"
)

// NewKVStoreService creates a new KVStoreService.
// This wrapper is kept for backwards compatibility.
// When migrating from runtime to runtime/v2, use runtimev2.NewKVStoreService(storeKey.Name()) instead of runtime.NewKVStoreService(storeKey).
func NewKVStoreService(storeKey string) store.KVStoreService {
	return stf.NewKVStoreService([]byte(storeKey))
}

type Store interface {
	// GetLatestVersion returns the latest version that consensus has been made on
	GetLatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// version. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error
}

// StoreLoader allows for custom loading of the store, this is useful when upgrading the store from a previous version
type StoreLoader func(store Store) error

// DefaultStoreLoader just calls LoadLatestVersion on the store
func DefaultStoreLoader(store Store) error {
	return store.LoadLatestVersion()
}

// UpgradeStoreLoader upgrades the store if the upgrade height matches the current version, it is used as a replacement
// for the DefaultStoreLoader when there are store upgrades
func UpgradeStoreLoader(upgradeHeight int64, storeUpgrades *store.StoreUpgrades) StoreLoader {
	// sanity checks on store upgrades
	if err := checkStoreUpgrade(storeUpgrades); err != nil {
		panic(err)
	}

	return func(store Store) error {
		latestVersion, err := store.GetLatestVersion()
		if err != nil {
			return err
		}

		if uint64(upgradeHeight) == latestVersion+1 {
			if len(storeUpgrades.Deleted) > 0 || len(storeUpgrades.Added) > 0 {
				if upgrader, ok := store.(storev2.UpgradeableStore); ok {
					return upgrader.LoadVersionAndUpgrade(latestVersion, storeUpgrades)
				}

				return fmt.Errorf("store does not support upgrades")
			}
		}

		return DefaultStoreLoader(store)
	}
}

// checkStoreUpgrade performs sanity checks on the store upgrades
func checkStoreUpgrade(storeUpgrades *store.StoreUpgrades) error {
	if storeUpgrades == nil {
		return errors.New("store upgrades cannot be nil")
	}

	// check for duplicates
	addedFilter := make(map[string]struct{})
	deletedFilter := make(map[string]struct{})

	for _, key := range storeUpgrades.Added {
		if _, ok := addedFilter[key]; ok {
			return fmt.Errorf("store upgrade has duplicate key %s in added", key)
		}
		addedFilter[key] = struct{}{}
	}
	for _, key := range storeUpgrades.Deleted {
		if _, ok := deletedFilter[key]; ok {
			return fmt.Errorf("store upgrade has duplicate key %s in deleted", key)
		}
		deletedFilter[key] = struct{}{}
	}

	for _, key := range storeUpgrades.Added {
		if _, ok := deletedFilter[key]; ok {
			return fmt.Errorf("store upgrade has key %s in both added and deleted", key)
		}
	}
	for _, key := range storeUpgrades.Deleted {
		if _, ok := addedFilter[key]; ok {
			return fmt.Errorf("store upgrade has key %s in both added and deleted", key)
		}
	}

	return nil
}
