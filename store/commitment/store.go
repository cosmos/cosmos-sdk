package commitment

import (
	"errors"
	"fmt"

	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

var _ store.Committer = (*CommitStore)(nil)

// CommitStore is a wrapper around multiple Tree objects mapped by a unique store
// key. Each store key reflects dedicated and unique usage within a module. A caller
// can construct a CommitStore with one or more store keys. It is expected that a
// RootStore use a CommitStore as an abstraction to handle multiple store keys
// and trees.
type CommitStore struct {
	logger log.Logger

	multiTrees map[string]Tree
}

// NewCommitStore creates a new CommitStore instance.
func NewCommitStore(multiTrees map[string]Tree, logger log.Logger) (*CommitStore, error) {
	return &CommitStore{
		logger:     logger,
		multiTrees: multiTrees,
	}, nil
}

func (c *CommitStore) WriteBatch(cs *store.Changeset) error {
	for storeKey, pairs := range cs.Pairs {
		tree, ok := c.multiTrees[storeKey]
		if !ok {
			return fmt.Errorf("store key %s not found in multiTrees", storeKey)
		}
		for _, kv := range pairs {
			if kv.Value == nil {
				if err := tree.Remove(kv.Key); err != nil {
					return err
				}
			} else if err := tree.Set(kv.Key, kv.Value); err != nil {
				return err
			}
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
		version := tree.GetLatestVersion()
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
				Version: tree.GetLatestVersion(),
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

func (c *CommitStore) Prune(version uint64) (ferr error) {
	for _, tree := range c.multiTrees {
		if err := tree.Prune(version); err != nil {
			ferr = errors.Join(ferr, err)
		}
	}

	return ferr
}

func (c *CommitStore) Close() (ferr error) {
	for _, tree := range c.multiTrees {
		if err := tree.Close(); err != nil {
			ferr = errors.Join(ferr, err)
		}
	}

	return ferr
}
