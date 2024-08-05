package root

import (
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/commitment/mem"
	"cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/internal"
	"cosmossdk.io/store/v2/pruning"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/pebbledb"
	"cosmossdk.io/store/v2/storage/sqlite"
)

type (
	SSType int
	SCType int
)

const (
	SSTypeSQLite SSType = 0
	SSTypePebble SSType = 1
	SSTypeRocks  SSType = 2
	SCTypeIavl   SCType = 0
	SCTypeIavlV2 SCType = 1
)

// app.toml config options
type Options struct {
	SSType          SSType               `mapstructure:"ss-type" toml:"ss-type" comment:"State storage database type. Currently we support: 0 for SQLite, 1 for Pebble"`
	SCType          SCType               `mapstructure:"sc-type" toml:"sc-type" comment:"State commitment database type. Currently we support:0 for iavl, 1 for iavl v2"`
	SSPruningOption *store.PruningOption `mapstructure:"ss-pruning-option" toml:"ss-pruning-option" comment:"Pruning options for state storage"`
	SCPruningOption *store.PruningOption `mapstructure:"sc-pruning-option" toml:"sc-pruning-option" comment:"Pruning options for state commitment"`
	IavlConfig      *iavl.Config         `mapstructure:"iavl-config" toml:"iavl-config"`
}

type FactoryOptions struct {
	Logger    log.Logger
	RootDir   string
	Options   Options
	StoreKeys []string
	SCRawDB   corestore.KVStoreWithBatch
}

func DefaultStoreOptions() Options {
	return Options{
		SSType: 0,
		SCType: 0,
		SCPruningOption: &store.PruningOption{
			KeepRecent: 2,
			Interval:   1,
		},
		SSPruningOption: &store.PruningOption{
			KeepRecent: 2,
			Interval:   1,
		},
		IavlConfig: &iavl.Config{
			CacheSize:              100_000,
			SkipFastStorageUpgrade: true,
		},
	}
}

// CreateRootStore is a convenience function to create a root store based on the
// provided FactoryOptions. Strictly speaking app developers can create the root
// store directly by calling root.New, so this function is not
// necessary, but demonstrates the required steps and configuration to create a root store.
func CreateRootStore(opts *FactoryOptions) (store.RootStore, error) {
	var (
		ssDb      storage.Database
		ss        *storage.StorageStore
		sc        *commitment.CommitStore
		err       error
		ensureDir = func(dir string) error {
			if err := os.MkdirAll(dir, 0o0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			return nil
		}
	)

	storeOpts := opts.Options
	switch storeOpts.SSType {
	case SSTypeSQLite:
		dir := fmt.Sprintf("%s/data/ss/sqlite", opts.RootDir)
		if err = ensureDir(dir); err != nil {
			return nil, err
		}
		ssDb, err = sqlite.New(dir)
	case SSTypePebble:
		dir := fmt.Sprintf("%s/data/ss/pebble", opts.RootDir)
		if err = ensureDir(dir); err != nil {
			return nil, err
		}
		ssDb, err = pebbledb.New(dir)
	case SSTypeRocks:
		// TODO: rocksdb requires build tags so is not supported here by default
		return nil, errors.New("rocksdb not supported")
	}
	if err != nil {
		return nil, err
	}
	ss = storage.NewStorageStore(ssDb, opts.Logger)

	metadata := commitment.NewMetadataStore(opts.SCRawDB)
	latestVersion, err := metadata.GetLatestVersion()
	if err != nil {
		return nil, err
	}
	if len(opts.StoreKeys) == 0 {
		lastCommitInfo, err := metadata.GetCommitInfo(latestVersion)
		if err != nil {
			return nil, err
		}
		if lastCommitInfo == nil {
			return nil, fmt.Errorf("tried to construct a root store with no store keys specified but no commit info found for version %d", latestVersion)
		}
		for _, si := range lastCommitInfo.StoreInfos {
			opts.StoreKeys = append(opts.StoreKeys, string(si.Name))
		}
	}
	removedStoreKeys, err := metadata.GetRemovedStoreKeys(latestVersion)
	if err != nil {
		return nil, err
	}

	newTreeFn := func(key string) (commitment.Tree, error) {
		if internal.IsMemoryStoreKey(key) {
			return mem.New(), nil
		} else {
			switch storeOpts.SCType {
			case SCTypeIavl:
				return iavl.NewIavlTree(db.NewPrefixDB(opts.SCRawDB, []byte(key)), opts.Logger, storeOpts.IavlConfig), nil
			case SCTypeIavlV2:
				return nil, fmt.Errorf("iavl v2 not supported")
			default:
				return nil, fmt.Errorf("unsupported commitment store type")
			}
		}
	}

	trees := make(map[string]commitment.Tree, len(opts.StoreKeys))
	for _, key := range opts.StoreKeys {
		tree, err := newTreeFn(key)
		if err != nil {
			return nil, err
		}
		trees[key] = tree
	}
	oldTrees := make(map[string]commitment.Tree, len(opts.StoreKeys))
	for _, key := range removedStoreKeys {
		tree, err := newTreeFn(string(key))
		if err != nil {
			return nil, err
		}
		oldTrees[string(key)] = tree
	}
	sc, err = commitment.NewCommitStore(trees, oldTrees, opts.SCRawDB, opts.Logger)
	if err != nil {
		return nil, err
	}

	pm := pruning.NewManager(sc, ss, storeOpts.SCPruningOption, storeOpts.SSPruningOption)

	return New(opts.Logger, ss, sc, pm, nil, nil)
}
