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

type FactoryOptions struct {
	Logger          log.Logger
	RootDir         string
	SSType          SSType
	SCType          SCType
	SSPruningOption *store.PruningOption
	SCPruningOption *store.PruningOption
	IavlConfig      *iavl.Config
	StoreKeys       []string
	SCRawDB         corestore.KVStoreWithBatch
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

	switch opts.SSType {
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

	if len(opts.StoreKeys) == 0 {
		metadata := commitment.NewMetadataStore(opts.SCRawDB)
		latestVersion, err := metadata.GetLatestVersion()
		if err != nil {
			return nil, err
		}
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

	trees := make(map[string]commitment.Tree)
	for _, key := range opts.StoreKeys {
		if internal.IsMemoryStoreKey(key) {
			trees[key] = mem.New()
		} else {
			switch opts.SCType {
			case SCTypeIavl:
				trees[key] = iavl.NewIavlTree(db.NewPrefixDB(opts.SCRawDB, []byte(key)), opts.Logger, opts.IavlConfig)
			case SCTypeIavlV2:
				return nil, errors.New("iavl v2 not supported")
			}
		}
	}
	sc, err = commitment.NewCommitStore(trees, opts.SCRawDB, opts.Logger)
	if err != nil {
		return nil, err
	}

	pm := pruning.NewManager(sc, ss, opts.SCPruningOption, opts.SSPruningOption)

	return New(opts.Logger, ss, sc, pm, nil, nil)
}
