package commitment

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"

	protoio "github.com/cosmos/gogoproto/io"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/internal"
	"cosmossdk.io/store/v2/internal/conv"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

var (
	_ store.Committer             = (*CommitStore)(nil)
	_ store.UpgradeableStore      = (*CommitStore)(nil)
	_ snapshots.CommitSnapshotter = (*CommitStore)(nil)
	_ store.PausablePruner        = (*CommitStore)(nil)
)

// MountTreeFn is a function that mounts a tree given a store key.
// It is used to lazily mount trees when needed (e.g. during upgrade or proof generation).
type MountTreeFn func(storeKey string) (Tree, error)

// CommitStore is a wrapper around multiple Tree objects mapped by a unique store
// key. Each store key reflects dedicated and unique usage within a module. A caller
// can construct a CommitStore with one or more store keys. It is expected that a
// RootStore use a CommitStore as an abstraction to handle multiple store keys
// and trees.
type CommitStore struct {
	logger     corelog.Logger
	metadata   *MetadataStore
	multiTrees map[string]Tree
	// oldTrees is a map of store keys to old trees that have been deleted or renamed.
	// It is used to get the proof for the old store keys.
	oldTrees map[string]Tree
}

// NewCommitStore creates a new CommitStore instance.
func NewCommitStore(trees, oldTrees map[string]Tree, db corestore.KVStoreWithBatch, logger corelog.Logger) (*CommitStore, error) {
	return &CommitStore{
		logger:     logger,
		multiTrees: trees,
		oldTrees:   oldTrees,
		metadata:   NewMetadataStore(db),
	}, nil
}

func (c *CommitStore) WriteChangeset(cs *corestore.Changeset) error {
	for _, pairs := range cs.Changes {
		key := conv.UnsafeBytesToStr(pairs.Actor)

		tree, ok := c.multiTrees[key]
		if !ok {
			return fmt.Errorf("store key %s not found in multiTrees", key)
		}
		for _, kv := range pairs.StateChanges {
			if kv.Remove {
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

func (c *CommitStore) WorkingCommitInfo(version uint64) *proof.CommitInfo {
	storeInfos := make([]proof.StoreInfo, 0, len(c.multiTrees))
	for storeKey, tree := range c.multiTrees {
		if internal.IsMemoryStoreKey(storeKey) {
			continue
		}
		bz := []byte(storeKey)
		storeInfos = append(storeInfos, proof.StoreInfo{
			Name: bz,
			CommitID: proof.CommitID{
				Version: version,
				Hash:    tree.WorkingHash(),
			},
		})
	}

	return &proof.CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

func (c *CommitStore) LoadVersion(targetVersion uint64) error {
	storeKeys := make([]string, 0, len(c.multiTrees))
	for storeKey := range c.multiTrees {
		storeKeys = append(storeKeys, storeKey)
	}
	return c.loadVersion(targetVersion, storeKeys)
}

// LoadVersionAndUpgrade implements store.UpgradeableStore.
func (c *CommitStore) LoadVersionAndUpgrade(targetVersion uint64, upgrades *corestore.StoreUpgrades) error {
	// deterministic iteration order for upgrades (as the underlying store may change and
	// upgrades make store changes where the execution order may matter)
	storeKeys := slices.Sorted(maps.Keys(c.multiTrees))
	removeTree := func(storeKey string) error {
		if oldTree, ok := c.multiTrees[storeKey]; ok {
			if err := oldTree.Close(); err != nil {
				return err
			}
			delete(c.multiTrees, storeKey)
		}
		return nil
	}

	newStoreKeys := make([]string, 0, len(c.multiTrees))
	removedStoreKeys := make([]string, 0)
	for _, storeKey := range storeKeys {
		// If it has been deleted, remove the tree.
		if upgrades.IsDeleted(storeKey) {
			if err := removeTree(storeKey); err != nil {
				return err
			}
			removedStoreKeys = append(removedStoreKeys, storeKey)
			continue
		}

		// If it has been added, set the initial version.
		if upgrades.IsAdded(storeKey) {
			if err := c.multiTrees[storeKey].SetInitialVersion(targetVersion + 1); err != nil {
				return err
			}
			// This is the empty tree, no need to load the version.
			continue
		}

		newStoreKeys = append(newStoreKeys, storeKey)
	}

	if err := c.metadata.flushRemovedStoreKeys(targetVersion, removedStoreKeys); err != nil {
		return err
	}

	return c.loadVersion(targetVersion, newStoreKeys)
}

func (c *CommitStore) loadVersion(targetVersion uint64, storeKeys []string) error {
	// Rollback the metadata to the target version.
	latestVersion, err := c.GetLatestVersion()
	if err != nil {
		return err
	}
	if targetVersion < latestVersion {
		for version := latestVersion; version > targetVersion; version-- {
			if err = c.metadata.deleteCommitInfo(version); err != nil {
				return err
			}
		}
		if err := c.metadata.setLatestVersion(targetVersion); err != nil {
			return err
		}
	}

	for _, storeKey := range storeKeys {
		if err := c.multiTrees[storeKey].LoadVersion(targetVersion); err != nil {
			return err
		}
	}

	// If the target version is greater than the latest version, it is the snapshot
	// restore case, we should create a new commit info for the target version.
	if targetVersion > latestVersion {
		cInfo := c.WorkingCommitInfo(targetVersion)
		return c.metadata.flushCommitInfo(targetVersion, cInfo)
	}

	return nil
}

func (c *CommitStore) Commit(version uint64) (*proof.CommitInfo, error) {
	storeInfos := make([]proof.StoreInfo, 0, len(c.multiTrees))

	for storeKey, tree := range c.multiTrees {
		if internal.IsMemoryStoreKey(storeKey) {
			continue
		}
		// If a commit event execution is interrupted, a new iavl store's version
		// will be larger than the RMS's metadata, when the block is replayed, we
		// should avoid committing that iavl store again.
		var commitID proof.CommitID
		v, err := tree.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		if v >= version {
			commitID.Version = version
			commitID.Hash = tree.Hash()
		} else {
			hash, cversion, err := tree.Commit()
			if err != nil {
				return nil, err
			}
			if cversion != version {
				return nil, fmt.Errorf("commit version %d does not match the target version %d", cversion, version)
			}
			commitID = proof.CommitID{
				Version: version,
				Hash:    hash,
			}
		}
		storeInfos = append(storeInfos, proof.StoreInfo{
			Name:     []byte(storeKey),
			CommitID: commitID,
		})
	}

	cInfo := &proof.CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}

	if err := c.metadata.flushCommitInfo(version, cInfo); err != nil {
		return nil, err
	}

	return cInfo, nil
}

func (c *CommitStore) SetInitialVersion(version uint64) error {
	for _, tree := range c.multiTrees {
		if err := tree.SetInitialVersion(version); err != nil {
			return err
		}
	}

	return nil
}

func (c *CommitStore) GetProof(storeKey []byte, version uint64, key []byte) ([]proof.CommitmentOp, error) {
	rawStoreKey := conv.UnsafeBytesToStr(storeKey)
	tree, ok := c.multiTrees[rawStoreKey]
	if !ok {
		tree, ok = c.oldTrees[rawStoreKey]
		if !ok {
			return nil, fmt.Errorf("store %s not found", rawStoreKey)
		}
	}

	iProof, err := tree.GetProof(version, key)
	if err != nil {
		return nil, err
	}
	cInfo, err := c.metadata.GetCommitInfo(version)
	if err != nil {
		return nil, err
	}
	if cInfo == nil {
		return nil, fmt.Errorf("commit info not found for version %d", version)
	}
	commitOp := proof.NewIAVLCommitmentOp(key, iProof)
	_, storeCommitmentOp, err := cInfo.GetStoreProof(storeKey)
	if err != nil {
		return nil, err
	}

	return []proof.CommitmentOp{commitOp, *storeCommitmentOp}, nil
}

func (c *CommitStore) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	tree, ok := c.multiTrees[conv.UnsafeBytesToStr(storeKey)]
	if !ok {
		return nil, fmt.Errorf("store %s not found", storeKey)
	}

	bz, err := tree.Get(version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s from store %s: %w", key, storeKey, err)
	}

	return bz, nil
}

// Prune implements store.Pruner.
func (c *CommitStore) Prune(version uint64) error {
	// prune the metadata
	for v := version; v > 0; v-- {
		if err := c.metadata.deleteCommitInfo(v); err != nil {
			return err
		}
	}
	// prune the trees
	for _, tree := range c.multiTrees {
		if err := tree.Prune(version); err != nil {
			return err
		}
	}
	// prune the removed store keys
	if err := c.pruneRemovedStoreKeys(version); err != nil {
		return err
	}

	return nil
}

func (c *CommitStore) pruneRemovedStoreKeys(version uint64) error {
	clearKVStore := func(storeKey []byte, version uint64) (err error) {
		tree, ok := c.oldTrees[string(storeKey)]
		if !ok {
			return fmt.Errorf("store %s not found in oldTrees", storeKey)
		}
		return tree.Prune(version)
	}
	return c.metadata.deleteRemovedStoreKeys(version, clearKVStore)
}

// PausePruning implements store.PausablePruner.
func (c *CommitStore) PausePruning(pause bool) {
	for _, tree := range c.multiTrees {
		if pruner, ok := tree.(store.PausablePruner); ok {
			pruner.PausePruning(pause)
		}
	}
}

// Snapshot implements snapshotstypes.CommitSnapshotter.
func (c *CommitStore) Snapshot(version uint64, protoWriter protoio.Writer) error {
	if version == 0 {
		return errors.New("the snapshot version must be greater than 0")
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
func (c *CommitStore) Restore(
	version uint64,
	format uint32,
	protoReader protoio.Reader,
	chStorage chan<- *corestore.StateChanges,
) (snapshotstypes.SnapshotItem, error) {
	var (
		importer     Importer
		snapshotItem snapshotstypes.SnapshotItem
		storeKey     []byte
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
				if err := importer.Close(); err != nil {
					return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to close importer: %w", err)
				}
			}

			storeKey = []byte(item.Store.Name)
			tree := c.multiTrees[item.Store.Name]
			if tree == nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("store %s not found", item.Store.Name)
			}
			importer, err = tree.Import(version)
			if err != nil {
				return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to import tree for version %d: %w", version, err)
			}
			defer importer.Close()

		case *snapshotstypes.SnapshotItem_IAVL:
			if importer == nil {
				return snapshotstypes.SnapshotItem{}, errors.New("received IAVL node item before store item")
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
				chStorage <- &corestore.StateChanges{
					Actor: storeKey,
					StateChanges: []corestore.KVPair{
						{
							Key:   node.Key,
							Value: node.Value,
						},
					},
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

func (c *CommitStore) GetCommitInfo(version uint64) (*proof.CommitInfo, error) {
	return c.metadata.GetCommitInfo(version)
}

func (c *CommitStore) GetLatestVersion() (uint64, error) {
	return c.metadata.GetLatestVersion()
}

func (c *CommitStore) Close() error {
	for _, tree := range c.multiTrees {
		if err := tree.Close(); err != nil {
			return err
		}
	}

	return nil
}
