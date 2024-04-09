package commitment

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

	protoio "github.com/cosmos/gogoproto/io"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	internal "cosmossdk.io/store/v2/internal/conv"
	"cosmossdk.io/store/v2/internal/encoding"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

const (
	commitInfoKeyFmt = "c/%d" // c/<version>
	latestVersionKey = "c/latest"
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
	logger     log.Logger
	db         store.RawDB
	multiTrees map[string]Tree

	// pruneOptions is the pruning configuration.
	pruneOptions *store.PruneOptions
}

// NewCommitStore creates a new CommitStore instance.
func NewCommitStore(trees map[string]Tree, db store.RawDB, pruneOpts *store.PruneOptions, logger log.Logger) (*CommitStore, error) {
	if pruneOpts == nil {
		pruneOpts = store.DefaultPruneOptions()
	}

	return &CommitStore{
		logger:       logger,
		db:           db,
		multiTrees:   trees,
		pruneOptions: pruneOpts,
	}, nil
}

func (c *CommitStore) WriteBatch(cs *corestore.Changeset) error {
	for _, pairs := range cs.Changes {

		key := internal.UnsafeBytesToStr(pairs.Actor)

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

func (c *CommitStore) GetLatestVersion() (uint64, error) {
	value, err := c.db.Get([]byte(latestVersionKey))
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, nil
	}

	version, _, err := encoding.DecodeUvarint(value)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (c *CommitStore) LoadVersion(targetVersion uint64) error {
	// Rollback the metadata to the target version.
	latestVersion, err := c.GetLatestVersion()
	if err != nil {
		return err
	}
	if targetVersion < latestVersion {
		batch := c.db.NewBatch()
		for version := latestVersion; version > targetVersion; version-- {
			cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
			if err := batch.Delete(cInfoKey); err != nil {
				return err
			}
		}
		if err := batch.WriteSync(); err != nil {
			return err
		}
	}

	for _, tree := range c.multiTrees {
		if err := tree.LoadVersion(targetVersion); err != nil {
			return err
		}
	}

	// If the target version is greater than the latest version, it is the snapshot
	// restore case, we should create a new commit info for the target version.
	var cInfo *proof.CommitInfo
	if targetVersion > latestVersion {
		cInfo = c.WorkingCommitInfo(targetVersion)
	}

	return c.flushCommitInfo(targetVersion, cInfo)
}

func (c *CommitStore) GetCommitInfo(version uint64) (*proof.CommitInfo, error) {
	key := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
	value, err := c.db.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	cInfo := &proof.CommitInfo{}
	if err := cInfo.Unmarshal(value); err != nil {
		return nil, err
	}

	return cInfo, nil
}

func (c *CommitStore) flushCommitInfo(version uint64, cInfo *proof.CommitInfo) error {
	batch := c.db.NewBatch()
	if cInfo != nil {
		cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, version))
		value, err := cInfo.Marshal()
		if err != nil {
			return err
		}
		if err := batch.Set(cInfoKey, value); err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	buf.Grow(encoding.EncodeUvarintSize(version))
	if err := encoding.EncodeUvarint(&buf, version); err != nil {
		return err
	}
	if err := batch.Set([]byte(latestVersionKey), buf.Bytes()); err != nil {
		return err
	}

	return batch.WriteSync()
}

func (c *CommitStore) Commit(version uint64) (*proof.CommitInfo, error) {
	storeInfos := make([]proof.StoreInfo, 0, len(c.multiTrees))

	for storeKey, tree := range c.multiTrees {
		// If a commit event execution is interrupted, a new iavl store's version
		// will be larger than the RMS's metadata, when the block is replayed, we
		// should avoid committing that iavl store again.
		var commitID proof.CommitID
		if tree.GetLatestVersion() >= version {
			commitID.Version = version
			commitID.Hash = tree.Hash()
		} else {
			hash, version, err := tree.Commit()
			if err != nil {
				return nil, err
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

	if err := c.flushCommitInfo(version, cInfo); err != nil {
		return nil, err
	}

	// Prune the old versions.
	if prune, pruneVersion := c.pruneOptions.ShouldPrune(version); prune {
		if err := c.Prune(pruneVersion); err != nil {
			c.logger.Info("failed to prune SC", "prune_version", pruneVersion, "err", err)
		}
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
	tree, ok := c.multiTrees[internal.UnsafeBytesToStr(storeKey)]
	if !ok {
		return nil, fmt.Errorf("store %s not found", storeKey)
	}

	iProof, err := tree.GetProof(version, key)
	if err != nil {
		return nil, err
	}
	cInfo, err := c.GetCommitInfo(version)
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
	tree, ok := c.multiTrees[internal.UnsafeBytesToStr(storeKey)]
	if !ok {
		return nil, fmt.Errorf("store %s not found", storeKey)
	}

	bz, err := tree.Get(version, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s from store %s: %w", key, storeKey, err)
	}

	return bz, nil
}

func (c *CommitStore) Prune(version uint64) (ferr error) {
	// prune the metadata
	batch := c.db.NewBatch()
	for v := version; v > 0; v-- {
		cInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, v))
		if exist, _ := c.db.Has(cInfoKey); !exist {
			break
		}
		if err := batch.Delete(cInfoKey); err != nil {
			return err
		}
	}
	if err := batch.WriteSync(); err != nil {
		return err
	}

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
func (c *CommitStore) Restore(version uint64, format uint32, protoReader protoio.Reader, chStorage chan<- *corestore.StateChanges) (snapshotstypes.SnapshotItem, error) {
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
			tree := c.multiTrees[item.Store.Name]
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

				key := []byte(storeKey)
				// If the node is a leaf node, it will be written to the storage.
				chStorage <- &corestore.StateChanges{
					Actor: key,
					StateChanges: []corestore.KVPair{
						{
							Key:    node.Key,
							Value:  node.Value,
							Remove: false,
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

func (c *CommitStore) Close() (ferr error) {
	for _, tree := range c.multiTrees {
		if err := tree.Close(); err != nil {
			ferr = errors.Join(ferr, err)
		}
	}

	return ferr
}
