package runtime

import (
	"fmt"
	"path/filepath"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/stf"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/root"
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

	// SetInitialVersion sets the initial version of the store.
	SetInitialVersion(uint64) error

	// WorkingHash writes the provided changeset to the state and returns
	// the working hash of the state.
	WorkingHash(changeset *store.Changeset) (store.Hash, error)

	// Commit commits the provided changeset and returns the new state root of the state.
	Commit(changeset *store.Changeset) (store.Hash, error)

	// Query is a key/value query directly to the underlying database. This skips the appmanager
	Query(storeKey []byte, version uint64, key []byte, prove bool) (storev2.QueryResult, error)

	// GetStateStorage returns the SS backend.
	GetStateStorage() storev2.VersionedDatabase

	// GetStateCommitment returns the SC backend.
	GetStateCommitment() storev2.Committer

	// LoadVersion loads the RootStore to the given version.
	LoadVersion(version uint64) error

	// LoadLatestVersion behaves identically to LoadVersion except it loads the
	// latest version implicitly.
	LoadLatestVersion() error

	// LastCommitID returns the latest commit ID
	LastCommitID() (proof.CommitID, error)
}

type StoreBuilder struct {
	store Store
}

func (sb *StoreBuilder) Build(
	logger log.Logger,
	storeKeys []string,
	config server.DynamicConfig,
	options root.Options,
) (Store, error) {
	if sb.store != nil {
		return sb.store, nil
	}
	home := config.GetString(FlagHome)
	scRawDb, err := db.NewDB(
		db.DBType(config.GetString("store.app-db-backend")),
		"application",
		filepath.Join(home, "data"),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCRawDB: %w", err)
	}

	factoryOptions := &root.FactoryOptions{
		Logger:    logger,
		RootDir:   home,
		Options:   options,
		StoreKeys: append(storeKeys, "stf"),
		SCRawDB:   scRawDb,
	}

	rs, err := root.CreateRootStore(factoryOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create root store: %w", err)
	}
	sb.store = rs
	return sb.store, nil
}

func (sb *StoreBuilder) Get() Store {
	return sb.store
}

func ProvideStoreBuilder() *StoreBuilder {
	return &StoreBuilder{}
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
