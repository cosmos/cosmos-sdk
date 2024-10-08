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
	"cosmossdk.io/store/v2/storage/rocksdb"
	"cosmossdk.io/store/v2/storage/sqlite"
)

type (
	SSType string
	SCType string
)

const (
	SSTypeSQLite SSType = "sqlite"
	SSTypePebble SSType = "pebble"
	SSTypeRocks  SSType = "rocksdb"
	SCTypeIavl   SCType = "iavl"
	SCTypeIavlV2 SCType = "iavl-v2"
)

// app.toml config options
type Options struct {
	SSType          SSType               `mapstructure:"ss-type" toml:"ss-type" comment:"SState storage database type. Currently we support: \"sqlite\", \"pebble\" and \"rocksdb\""`
	SCType          SCType               `mapstructure:"sc-type" toml:"sc-type" comment:"State commitment database type. Currently we support: \"iavl\" and \"iavl-v2\""`
	SSPruningOption *store.PruningOption `mapstructure:"ss-pruning-option" toml:"ss-pruning-option" comment:"Pruning options for state storage"`
	SCPruningOption *store.PruningOption `mapstructure:"sc-pruning-option" toml:"sc-pruning-option" comment:"Pruning options for state commitment"`
	IavlConfig      *iavl.Config         `mapstructure:"iavl-config" toml:"iavl-config"`
}

// Namespace is a store key and its commitment structure.
type Namespace struct {
	Name      string
	Structure string
}

// FactoryOptions are the options for creating a root store.
type FactoryOptions struct {
	Logger     log.Logger
	RootDir    string
	Options    Options
	Namespaces []Namespace
	SCRawDB    corestore.KVStoreWithBatch
}

// DefaultStoreOptions returns the default options for creating a root store.
func DefaultStoreOptions() Options {
	return Options{
		SSType: SSTypeSQLite,
		SCType: SCTypeIavl,
		SCPruningOption: &store.PruningOption{
			KeepRecent: 2,
			Interval:   100,
		},
		SSPruningOption: &store.PruningOption{
			KeepRecent: 2,
			Interval:   100,
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
		dir := fmt.Sprintf("%s/data/ss/rocksdb", opts.RootDir)
		if err = ensureDir(dir); err != nil {
			return nil, err
		}
		ssDb, err = rocksdb.New(dir)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", opts.Options.SSType)
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
	if len(opts.Namespaces) == 0 {
		lastCommitInfo, err := metadata.GetCommitInfo(latestVersion)
		if err != nil {
			return nil, err
		}
		if lastCommitInfo == nil {
			return nil, fmt.Errorf("tried to construct a root store with no store keys specified but no commit info found for version %d", latestVersion)
		}
		for _, si := range lastCommitInfo.StoreInfos {
			opts.Namespaces = append(opts.Namespaces, Namespace{
				Name:      string(si.Name),
				Structure: si.Structure,
			})
		}
	}
	removedStoreKeys, err := metadata.GetRemovedStoreKeys(latestVersion)
	if err != nil {
		return nil, err
	}

	newTreeFn := func(ns Namespace) (commitment.Tree, error) {
		if internal.IsMemoryStoreKey(string(ns.Name)) {
			return mem.New(), nil
		} else {
			scType := storeOpts.SCType
			if ns.Structure != "" {
				scType = SCType(ns.Structure)
			}
			switch scType {
			case SCTypeIavl:
				return iavl.NewIavlTree(db.NewPrefixDB(opts.SCRawDB, []byte(ns.Name)), opts.Logger, storeOpts.IavlConfig), nil
			case SCTypeIavlV2:
				return nil, errors.New("iavl v2 not supported")
			default:
				return nil, errors.New("unsupported commitment store type")
			}
		}
	}

	trees := make(map[string]commitment.Tree, len(opts.Namespaces))
	for _, ns := range opts.Namespaces {
		tree, err := newTreeFn(ns)
		if err != nil {
			return nil, err
		}
		trees[ns.Name] = tree
	}
	oldTrees := make(map[string]commitment.Tree, len(opts.Namespaces))
	for _, key := range removedStoreKeys {
		// TODO: removedStoreKeys must store commitment type or have a way to recover it
		tree, err := newTreeFn(Namespace{Name: string(key), Structure: "iavl"})
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
