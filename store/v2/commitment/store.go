package commitment

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"

	protoio "github.com/cosmos/gogoproto/io"
	"golang.org/x/sync/errgroup"

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

	// NOTE: It is not recommended to use the CommitStore as a reader. This is only used
	// during the migration process. Generally, the SC layer does not provide a reader
	// in the store/v2.
	_ store.VersionedReader = (*CommitStore)(nil)
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
	eg := new(errgroup.Group)
	eg.SetLimit(store.MaxWriteParallelism)
	for _, pairs := range cs.Changes {
		key := conv.UnsafeBytesToStr(pairs.Actor)

		tree, ok := c.multiTrees[key]
		if !ok {
			return fmt.Errorf("store key %s not found in multiTrees", key)
		}
		if tree.IsConcurrentSafe() {
			eg.Go(func() error {
				return writeChangeset(tree, pairs)
			})
		} else {
			if err := writeChangeset(tree, pairs); err != nil {
				return err
			}
		}
	}

	return eg.Wait()
}

func writeChangeset(tree Tree, changes corestore.StateChanges) error {
	for _, kv := range changes.StateChanges {
		if kv.Remove {
			if err := tree.Remove(kv.Key); err != nil {
				return err
			}
		} else if err := tree.Set(kv.Key, kv.Value); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommitStore) LoadVersion(targetVersion uint64) error {
	storeKeys := make([]string, 0, len(c.multiTrees))
	for storeKey := range c.multiTrees {
		storeKeys = append(storeKeys, storeKey)
	}
	return c.loadVersion(targetVersion, storeKeys, false)
}

func (c *CommitStore) LoadVersionForOverwriting(targetVersion uint64) error {
	storeKeys := make([]string, 0, len(c.multiTrees))
	for storeKey := range c.multiTrees {
		storeKeys = append(storeKeys, storeKey)
	}

	return c.loadVersion(targetVersion, storeKeys, true)
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

	return c.loadVersion(targetVersion, newStoreKeys, true)
}

func (c *CommitStore) loadVersion(targetVersion uint64, storeKeys []string, overrideAfter bool) error {
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

	eg := errgroup.Group{}
	eg.SetLimit(store.MaxWriteParallelism)
	for _, storeKey := range storeKeys {
		tree := c.multiTrees[storeKey]
		if overrideAfter {
			if tree.IsConcurrentSafe() {
				eg.Go(func() error {
					return c.multiTrees[storeKey].LoadVersionForOverwriting(targetVersion)
				})
			} else {
				if err := c.multiTrees[storeKey].LoadVersionForOverwriting(targetVersion); err != nil {
					return err
				}
			}
		} else {
			if tree.IsConcurrentSafe() {
				eg.Go(func() error { return c.multiTrees[storeKey].LoadVersion(targetVersion) })
			} else {
				if err := c.multiTrees[storeKey].LoadVersion(targetVersion); err != nil {
					return err
				}
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	// If the target version is greater than the latest version, it is the snapshot
	// restore case, we should create a new commit info for the target version.
	if targetVersion > latestVersion {
		cInfo, err := c.GetCommitInfo(targetVersion)
		if err != nil {
			return err
		}
		return c.metadata.flushCommitInfo(targetVersion, cInfo)
	}

	return nil
}

func (c *CommitStore) Commit(version uint64) (*proof.CommitInfo, error) {
	storeInfos := make([]*proof.StoreInfo, 0, len(c.multiTrees))
	eg := new(errgroup.Group)
	eg.SetLimit(store.MaxWriteParallelism)

	for storeKey, tree := range c.multiTrees {
		if internal.IsMemoryStoreKey(storeKey) {
			continue
		}
		si := &proof.StoreInfo{Name: storeKey}
		storeInfos = append(storeInfos, si)

		if tree.IsConcurrentSafe() {
			eg.Go(func() error {
				err := c.commit(tree, si, version)
				if err != nil {
					return fmt.Errorf("commit fail: %s: %w", si.Name, err)
				}
				return nil
			})
		} else {
			err := c.commit(tree, si, version)
			if err != nil {
				return nil, err
			}
		}
	}

	// convert storeInfos to []proof.StoreInfo
	sideref := make([]*proof.StoreInfo, 0, len(c.multiTrees))
	sideref = append(sideref, storeInfos...)

	cInfo := &proof.CommitInfo{
		Version:    int64(version),
		StoreInfos: sideref,
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if err := c.metadata.flushCommitInfo(version, cInfo); err != nil {
		return nil, err
	}

	return cInfo, nil
}

func (c *CommitStore) commit(tree Tree, si *proof.StoreInfo, expected uint64) error {
	h, v, err := tree.Commit()
	if err != nil {
		return err
	}
	if v != expected {
		return fmt.Errorf("commit version %d does not match the target version %d", v, expected)
	}
	si.CommitId = &proof.CommitID{
		Version: int64(v),
		Hash:    h,
	}
	return nil
}

func (c *CommitStore) SetInitialVersion(version uint64) error {
	for _, tree := range c.multiTrees {
		if err := tree.SetInitialVersion(version); err != nil {
			return err
		}
	}

	return nil
}

// GetProof returns a proof for the given key and version.
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

// getReader returns a reader for the given store key. It will return an error if the
// store key does not exist or the tree does not implement the Reader interface.
// WARNING: This function is only used during the migration process. The SC layer
// generally does not provide a reader for the CommitStore.
func (c *CommitStore) getReader(storeKey string) (Reader, error) {
	var tree Tree
	if storeTree, ok := c.oldTrees[storeKey]; ok {
		tree = storeTree
	} else if storeTree, ok := c.multiTrees[storeKey]; ok {
		tree = storeTree
	} else {
		return nil, fmt.Errorf("store %s not found", storeKey)
	}

	reader, ok := tree.(Reader)
	if !ok {
		return nil, fmt.Errorf("tree for store %s does not implement Reader", storeKey)
	}

	return reader, nil
}

// VersionExists implements store.VersionedReader.
func (c *CommitStore) VersionExists(version uint64) (bool, error) {
	latestVersion, err := c.metadata.GetLatestVersion()
	if err != nil {
		return false, err
	}
	if latestVersion == 0 {
		return version == 0, nil
	}

	ci, err := c.metadata.GetCommitInfo(version)
	return ci != nil, err
}

// Get implements store.VersionedReader.
func (c *CommitStore) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	reader, err := c.getReader(conv.UnsafeBytesToStr(storeKey))
	if err != nil {
		return nil, err
	}

	bz, err := reader.Get(version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s from store %s: %w", key, storeKey, err)
	}

	return bz, nil
}

// Has implements store.VersionedReader.
func (c *CommitStore) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	val, err := c.Get(storeKey, version, key)
	return val != nil, err
}

// Iterator implements store.VersionedReader.
func (c *CommitStore) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	reader, err := c.getReader(conv.UnsafeBytesToStr(storeKey))
	if err != nil {
		return nil, err
	}

	return reader.Iterator(version, start, end, true)
}

// ReverseIterator implements store.VersionedReader.
func (c *CommitStore) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	reader, err := c.getReader(conv.UnsafeBytesToStr(storeKey))
	if err != nil {
		return nil, err
	}

	return reader.Iterator(version, start, end, false)
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
) (snapshotstypes.SnapshotItem, error) {
	var (
		importer     Importer
		snapshotItem snapshotstypes.SnapshotItem
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
			}
			if node.Version == 0 {
				node.Version = int64(version)
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
	// if the commit info is already stored, return it
	ci, err := c.metadata.GetCommitInfo(version)
	if err != nil {
		return nil, err
	}
	if ci != nil {
		return ci, nil
	}
	// otherwise built the commit info from the trees
	storeInfos := make([]*proof.StoreInfo, 0, len(c.multiTrees))
	for storeKey, tree := range c.multiTrees {
		if internal.IsMemoryStoreKey(storeKey) {
			continue
		}
		v := tree.Version()
		if v != version {
			return nil, fmt.Errorf("tree version %d does not match the target version %d", v, version)
		}
		storeInfos = append(storeInfos, &proof.StoreInfo{
			Name: storeKey,
			CommitId: &proof.CommitID{
				Version: int64(v),
				Hash:    tree.Hash(),
			},
		})
	}

	ci = &proof.CommitInfo{
		Version:    int64(version),
		StoreInfos: storeInfos,
	}
	return ci, nil
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
