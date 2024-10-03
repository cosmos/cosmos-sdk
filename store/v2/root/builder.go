package root

import (
	"fmt"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/db"
)

type Builder interface {
	Build(log.Logger, *Config) (store.RootStore, error)
	RegisterKey(key string)
	Get() store.RootStore
}

var _ Builder = (*builder)(nil)

// builder is a builder for a store/v2 RootStore satisfying the Store interface
// which is primarily used by depinject to assemble the store/v2 RootStore.  Users not using
// depinject should use the Factory called in Build directly.
type builder struct {
	// input
	storeKeys map[string]struct{}

	// output
	store store.RootStore
}

func NewBuilder() Builder {
	return &builder{storeKeys: make(map[string]struct{})}
}

// Build creates a new store/v2 RootStore.
func (sb *builder) Build(
	logger log.Logger,
	config *Config,
) (store.RootStore, error) {
	if sb.store != nil {
		return sb.store, nil
	}
	if config.Home == "" {
		return nil, fmt.Errorf("home directory is required")
	}
	scRawDb, err := db.NewDB(
		db.DBType(config.AppDBBackend),
		"application",
		filepath.Join(config.Home, "data"),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCRawDB: %w", err)
	}

	var storeKeys []string
	for key := range sb.storeKeys {
		storeKeys = append(storeKeys, key)
	}

	factoryOptions := &FactoryOptions{
		Logger:    logger,
		RootDir:   config.Home,
		Options:   config.Options,
		StoreKeys: storeKeys,
		SCRawDB:   scRawDb,
	}

	rs, err := CreateRootStore(factoryOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create root store: %w", err)
	}
	sb.store = rs
	return sb.store, nil
}

// Get returns the Store.  Build must be called before calling Get or the result will be nil.
func (sb *builder) Get() store.RootStore {
	return sb.store
}

func (sb *builder) RegisterKey(key string) {
	sb.storeKeys[key] = struct{}{}
}
