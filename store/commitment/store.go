package commitment

import (
	"errors"
	"fmt"
	"io"
	"math"

	protoio "github.com/cosmos/gogoproto/io"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

var (
	_ store.Committer             = (*CommitStore)(nil)
	_ snapshots.CommitSnapshotter = (*CommitStore)(nil)
)

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

func (c *CommitStore) SetInitialVersion(version uint64) error {
	for _, tree := range c.multiTrees {
		if err := tree.SetInitialVersion(version); err != nil {
			return err
		}
	}

	return nil
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

// Snapshot implements snapshotstypes.CommitSnapshotter.
func (c *CommitStore) Snapshot(version uint64, protoWriter protoio.Writer) error {
	if version == 0 {
		return fmt.Errorf("the snapshot version must be greater than 0")
	}

	latestVersion, err := c.GetLatestVersion()
	if err != nil {
		return err
	}
	if version > latestVersion {
		return fmt.Errorf("the snapshot version %d is greater than the latest version %d", version, latestVersion)
	}

	for storeKey, tree := range c.multiTrees {
		// TODO: check the parallelism of this loop
		if err := func() error {
			exporter, err := tree.Export(version)
			if err != nil {
				return fmt.Errorf("failed to export tree for version %d: %w", version, err)
			}
			defer exporter.Close()

			err = protoWriter.WriteMsg(&snapshotstypes.SnapshotItem{
				Item: &snapshotstypes.SnapshotItem_Store{
					Store: &snapshotstypes.SnapshotStoreItem{
						Name: storeKey,
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to write store name: %w", err)
			}

			for {
				item, err := exporter.Next()
				if errors.Is(err, ErrorExportDone) {
					break
				} else if err != nil {
					return fmt.Errorf("failed to get the next export node: %w", err)
				}

				if err = protoWriter.WriteMsg(&snapshotstypes.SnapshotItem{
					Item: &snapshotstypes.SnapshotItem_IAVL{
						IAVL: item,
					},
				}); err != nil {
					return fmt.Errorf("failed to write iavl node: %w", err)
				}
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

// Restore implements snapshotstypes.CommitSnapshotter.
func (c *CommitStore) Restore(version uint64, format uint32, protoReader protoio.Reader, chStorage chan<- *store.KVPair) (snapshotstypes.SnapshotItem, error) {
	var (
		importer     Importer
		snapshotItem snapshotstypes.SnapshotItem
		storeKey     string
	)

loop:
	for {
		snapshotItem = snapshotstypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return snapshotstypes.SnapshotItem{}, fmt.Errorf("invalid protobuf message: %w", err)
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshotstypes.SnapshotItem_Store:
			if importer != nil {
				if err := importer.Commit(); err != nil {
					return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to commit importer: %w", err)
				}
				importer.Close()
			}
			storeKey = item.Store.Name
			tree := c.multiTrees[storeKey]
			if tree == nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("store %s not found", storeKey)
			}
			importer, err = tree.Import(version)
			if err != nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to import tree for version %d: %w", version, err)
			}
			defer importer.Close()

		case *snapshotstypes.SnapshotItem_IAVL:
			if importer == nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("received IAVL node item before store item")
			}
			node := item.IAVL
			if node.Height > int32(math.MaxInt8) {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("node height %v cannot exceed %v",
					item.IAVL.Height, math.MaxInt8)
			}
			// Protobuf does not differentiate between []byte{} and nil, but fortunately IAVL does
			// not allow nil keys nor nil values for leaf nodes, so we can always set them to empty.
			if node.Key == nil {
				node.Key = []byte{}
			}
			if node.Height == 0 {
				if node.Value == nil {
					node.Value = []byte{}
				}
				// If the node is a leaf node, it will be written to the storage.
				chStorage <- &store.KVPair{
					Key:      node.Key,
					Value:    node.Value,
					StoreKey: storeKey,
				}
			}
			err := importer.Add(node)
			if err != nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to add node to importer: %w", err)
			}
		default:
			break loop
		}
	}

	if importer != nil {
		if err := importer.Commit(); err != nil {
			return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to commit importer: %w", err)
		}
	}

	return snapshotItem, c.LoadVersion(version)
}

func (c *CommitStore) Close() (ferr error) {
	for _, tree := range c.multiTrees {
		if err := tree.Close(); err != nil {
			ferr = errors.Join(ferr, err)
		}
	}

	return ferr
}
