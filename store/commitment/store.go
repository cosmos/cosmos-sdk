package commitment

import (
	"fmt"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/commitment/types"
)

const (
	// DefaultTreeType is the default tree type.
	DefaultTreeType = "default"
	IavlTreeType    = "iavl"
)

var _ store.Committer = (*CommitStore)(nil)

// CommitStore is a wrapper around the state Tree.
type CommitStore struct {
	multiTrees map[string]types.Tree
}

// NewCommitStore creates a new CommitStore instance.
func NewCommitStore(storeConfigs map[string]interface{}, db dbm.DB, logger log.Logger) (*CommitStore, error) {
	multiTrees := make(map[string]types.Tree)
	for storeKey, cfg := range storeConfigs {
		switch cfg {
		case cfg.(*iavl.Config):
			iavlDB := dbm.NewPrefixDB(db, []byte(storeKey))
			multiTrees[storeKey] = iavl.NewIavlTree(iavlDB, logger, cfg.(*iavl.Config))
		default:
			return nil, fmt.Errorf("unknown tree type for store %s, config: %v", storeKey, cfg)
		}
	}

	return &CommitStore{
		multiTrees: multiTrees,
	}, nil
}

func (c *CommitStore) WriteBatch(cs *store.Changeset) error {
	for _, kv := range cs.Pairs {
		if kv.Value == nil {
			if err := c.multiTrees[kv.StoreKey].Remove(kv.Key); err != nil {
				return err
			}
		} else if err := c.multiTrees[kv.StoreKey].Set(kv.Key, kv.Value); err != nil {
			return err
		}
	}

	return nil
}

func (c *CommitStore) WorkingStoreInfos(version uint64) []store.StoreInfo {
	storeInfos := make([]store.StoreInfo, 0, len(c.multiTrees))
	for storeKey, tree := range c.multiTrees {
		storeInfos = append(storeInfos, store.StoreInfo{
			Name: storeKey,
			CommitID: store.CommitID{
				Version: version,
				Hash:    tree.WorkingHash(),
			},
		})
	}

	return storeInfos
}

func (c *CommitStore) GetLatestVersion() (uint64, error) {
	latestVersion := uint64(0)
	for storeKey, tree := range c.multiTrees {
		version := uint64(tree.GetLatestVersion())
		if latestVersion != 0 && version != latestVersion {
			return 0, fmt.Errorf("store %s has version %d, not equal to latest version %d", storeKey, version, latestVersion)
		}
		latestVersion = version
	}

	return latestVersion, nil
}

func (c *CommitStore) LoadVersion(targetVersion uint64) error {
	for _, tree := range c.multiTrees {
		if err := tree.LoadVersion(targetVersion); err != nil {
			return err
		}
	}

	return nil
}

func (c *CommitStore) Commit() ([]store.StoreInfo, error) {
	storeInfos := make([]store.StoreInfo, 0, len(c.multiTrees))
	for storeKey, tree := range c.multiTrees {
		hash, err := tree.Commit()
		if err != nil {
			return nil, err
		}
		storeInfos = append(storeInfos, store.StoreInfo{
			Name: storeKey,
			CommitID: store.CommitID{
				Version: uint64(tree.GetLatestVersion()),
				Hash:    hash,
			},
		})
	}

	return storeInfos, nil
}

func (c *CommitStore) GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	tree, ok := c.multiTrees[storeKey]
	if !ok {
		return nil, fmt.Errorf("store %s not found", storeKey)
	}

	return tree.GetProof(version, key)
}

func (c *CommitStore) Prune(version uint64) error {
	for _, tree := range c.multiTrees {
		if err := tree.Prune(version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CommitStore) Close() error {
	for _, tree := range c.multiTrees {
		err := tree.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
