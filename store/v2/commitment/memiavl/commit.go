package memiavl

import (
	stderrors "errors"
	"fmt"
	"io"
	"math"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
	protoio "github.com/cosmos/gogoproto/io"
	ics23 "github.com/cosmos/ics23/go"

	"github.com/crypto-org-chain/cronos/memiavl"
)

var (
	_ store.Committer             = (*CommitStore)(nil)
	_ store.UpgradeableStore      = (*CommitStore)(nil)
	_ snapshots.CommitSnapshotter = (*CommitStore)(nil)
	_ store.PausablePruner        = (*CommitStore)(nil)
)

type CommitStore struct {
	dir       string
	db        *memiavl.DB
	logger    corelog.Logger
	opts      memiavl.Options
	storeKeys []string

	lastCommitInfo                  *proof.CommitInfo
	supportExportNonSnapshotVersion bool
}

func NewCommitStore(dir string, storeKeys []string, logger corelog.Logger, opts memiavl.Options) *CommitStore {
	return &CommitStore{
		dir:       dir,
		opts:      opts,
		storeKeys: storeKeys,
		logger:    logger,
	}
}

func (c *CommitStore) WriteChangeset(cs *corestore.Changeset) error {
	return c.db.ApplyChangeSets(convertChangeSet(cs))
}

func (c *CommitStore) LoadVersion(targetVersion uint64) error {
	return c.LoadVersionAndUpgrade(targetVersion, nil)
}

func (c *CommitStore) LoadVersionForOverwriting(targetVersion uint64) error {
	panic("not implemented, TODO")
}

func (c *CommitStore) LoadVersionAndUpgrade(targetVersion uint64, upgrades *corestore.StoreUpgrades) error {
	opts := c.opts
	opts.CreateIfMissing = true
	opts.InitialStores = c.storeKeys
	opts.TargetVersion = uint32(targetVersion)
	db, err := memiavl.Load(c.dir, opts)
	if err != nil {
		return err
	}

	var treeUpgrades []*memiavl.TreeNameUpgrade
	if upgrades != nil {
		for _, name := range upgrades.Deleted {
			treeUpgrades = append(treeUpgrades, &memiavl.TreeNameUpgrade{Name: name, Delete: true})
		}
		for _, name := range upgrades.Added {
			treeUpgrades = append(treeUpgrades, &memiavl.TreeNameUpgrade{Name: name})
		}
	}

	if len(treeUpgrades) > 0 {
		if err := db.ApplyUpgrades(treeUpgrades); err != nil {
			return err
		}
	}

	c.db = db
	return nil
}

func (c *CommitStore) GetLatestVersion() (uint64, error) {
	return uint64(c.db.Version()), nil
}

func (c *CommitStore) Commit(version uint64) (*proof.CommitInfo, error) {
	cversion, err := c.db.Commit()
	if err != nil {
		return nil, err
	}

	if uint64(cversion) != version {
		return nil, fmt.Errorf("commit version %d does not match the target version %d", cversion, version)
	}

	c.lastCommitInfo = convertCommitInfo(c.db.LastCommitInfo())
	return c.lastCommitInfo, nil
}

func (c *CommitStore) SetInitialVersion(version uint64) error {
	return c.db.SetInitialVersion(int64(version))
}

func (c *CommitStore) withDB(version uint64, cb func(db *memiavl.DB) error) error {
	// If the request's height is the latest height we've committed, then utilize
	// the store's lastCommitInfo as this commit info may not be flushed to disk.
	// Otherwise, we query for the commit info from disk.
	db := c.db
	if version != uint64(db.LastCommitInfo().Version) {
		var err error
		db, err = memiavl.Load(c.dir, memiavl.Options{TargetVersion: uint32(version), ReadOnly: true})
		if err != nil {
			return err
		}
		defer db.Close()
	}

	return cb(db)
}

func (c *CommitStore) GetProof(storeKey []byte, version uint64, key []byte) ([]proof.CommitmentOp, error) {
	var result []proof.CommitmentOp
	if err := c.withDB(version, func(db *memiavl.DB) error {
		tree := db.TreeByName(string(storeKey))
		value := tree.Get(key)
		commitOp := getProofFromTree(tree, key, value != nil)
		cInfo := convertCommitInfo(db.LastCommitInfo())
		_, storeCommitmentOp, err := cInfo.GetStoreProof(storeKey)
		if err != nil {
			return err
		}

		result = []proof.CommitmentOp{commitOp, *storeCommitmentOp}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *CommitStore) Get(storeKey []byte, version uint64, key []byte) ([]byte, error) {
	var value []byte
	if err := c.withDB(version, func(db *memiavl.DB) error {
		tree := db.TreeByName(string(storeKey))
		value = tree.Get(key)
		return nil
	}); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *CommitStore) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	var value corestore.Iterator
	if err := c.withDB(version, func(db *memiavl.DB) error {
		tree := db.TreeByName(string(storeKey))
		value = tree.Iterator(start, end, true)
		return nil
	}); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *CommitStore) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	var value corestore.Iterator
	if err := c.withDB(version, func(db *memiavl.DB) error {
		tree := db.TreeByName(string(storeKey))
		value = tree.Iterator(start, end, false)
		return nil
	}); err != nil {
		return nil, err
	}

	return value, nil
}

// VersionExists only returns true if the version is the latest version
func (c *CommitStore) VersionExists(v uint64) (bool, error) {
	v1 := uint64(c.db.SnapshotVersion())
	v2, err := c.GetLatestVersion()
	if err != nil {
		return false, err
	}

	return v >= v1 && v <= v2, nil
}

func (c *CommitStore) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	var value bool
	if err := c.withDB(version, func(db *memiavl.DB) error {
		tree := db.TreeByName(string(storeKey))
		value = tree.Has(key)
		return nil
	}); err != nil {
		return false, err
	}

	return value, nil
}

func (c *CommitStore) GetCommitInfo(version uint64) (*proof.CommitInfo, error) {
	var value *memiavl.CommitInfo
	if err := c.withDB(version, func(db *memiavl.DB) error {
		value = db.LastCommitInfo()
		return nil
	}); err != nil {
		return nil, err
	}

	return convertCommitInfo(value), nil
}

func (c *CommitStore) Snapshot(height uint64, protoWriter protoio.Writer) (returnErr error) {
	if height > math.MaxUint32 {
		return fmt.Errorf("height overflows uint32: %d", height)
	}
	version := uint32(height)

	exporter, err := memiavl.NewMultiTreeExporter(c.dir, version, c.supportExportNonSnapshotVersion)
	if err != nil {
		return err
	}

	defer func() {
		returnErr = stderrors.Join(returnErr, exporter.Close())
	}()

	for {
		item, err := exporter.Next()
		if err != nil {
			if err == memiavl.ErrorExportDone {
				break
			}

			return err
		}

		switch item := item.(type) {
		case *memiavl.ExportNode:
			if err := protoWriter.WriteMsg(&snapshotstypes.SnapshotItem{
				Item: &snapshotstypes.SnapshotItem_IAVL{
					IAVL: &snapshotstypes.SnapshotIAVLItem{
						Key:     item.Key,
						Value:   item.Value,
						Height:  int32(item.Height),
						Version: item.Version,
					},
				},
			}); err != nil {
				return err
			}
		case string:
			if err := protoWriter.WriteMsg(&snapshotstypes.SnapshotItem{
				Item: &snapshotstypes.SnapshotItem_Store{
					Store: &snapshotstypes.SnapshotStoreItem{
						Name: item,
					},
				},
			}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown item type %T", item)
		}
	}

	return nil
}

func (c *CommitStore) Restore(version uint64, format uint32, protoReader protoio.Reader) (snapshotstypes.SnapshotItem, error) {
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return snapshotstypes.SnapshotItem{}, fmt.Errorf("failed to close db: %w", err)
		}
		c.db = nil
	}

	item, err := c.restore(version, format, protoReader)
	if err != nil {
		return snapshotstypes.SnapshotItem{}, err
	}

	return item, err
}

func (rs *CommitStore) restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshotstypes.SnapshotItem, error) {
	importer, err := memiavl.NewMultiTreeImporter(rs.dir, height)
	if err != nil {
		return snapshotstypes.SnapshotItem{}, err
	}
	defer importer.Close()

	var snapshotItem snapshotstypes.SnapshotItem
loop:
	for {
		snapshotItem = snapshotstypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshotstypes.SnapshotItem{}, errors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshotstypes.SnapshotItem_Store:
			if err := importer.AddTree(item.Store.Name); err != nil {
				return snapshotstypes.SnapshotItem{}, err
			}
		case *snapshotstypes.SnapshotItem_IAVL:
			if item.IAVL.Height > math.MaxInt8 {
				return snapshotstypes.SnapshotItem{}, errors.Wrapf(storetypes.ErrLogic, "node height %v cannot exceed %v",
					item.IAVL.Height, math.MaxInt8)
			}
			node := &memiavl.ExportNode{
				Key:     item.IAVL.Key,
				Value:   item.IAVL.Value,
				Height:  int8(item.IAVL.Height),
				Version: item.IAVL.Version,
			}
			// Protobuf does not differentiate between []byte{} as nil, but fortunately IAVL does
			// not allow nil keys nor nil values for leaf nodes, so we can always set them to empty.
			if node.Key == nil {
				node.Key = []byte{}
			}
			if node.Height == 0 && node.Value == nil {
				node.Value = []byte{}
			}
			importer.AddNode(node)
		default:
			// unknown element, could be an extension
			break loop
		}
	}

	if err := importer.Finalize(); err != nil {
		return snapshotstypes.SnapshotItem{}, err
	}

	return snapshotItem, nil
}

func (c *CommitStore) Prune(version uint64) error {
	return nil
}

func (c *CommitStore) PausePruning(pause bool) {
}

func (c *CommitStore) Close() error {
	return c.db.Close()
}

func convertChangeSet(cs *corestore.Changeset) []*memiavl.NamedChangeSet {
	result := make([]*memiavl.NamedChangeSet, len(cs.Changes))
	for i, change := range cs.Changes {
		pairs := make([]*memiavl.KVPair, len(change.StateChanges))
		for j, pair := range change.StateChanges {
			pairs[j] = &memiavl.KVPair{
				Key:    pair.Key,
				Value:  pair.Value,
				Delete: pair.Remove,
			}
		}
		result[i] = &memiavl.NamedChangeSet{
			Name: string(change.Actor),
			Changeset: memiavl.ChangeSet{
				Pairs: pairs,
			},
		}
	}

	return result
}

func convertCommitInfo(commitInfo *memiavl.CommitInfo) *proof.CommitInfo {
	storeInfos := make([]*proof.StoreInfo, len(commitInfo.StoreInfos))
	for i, storeInfo := range commitInfo.StoreInfos {
		storeInfos[i] = &proof.StoreInfo{
			Name: []byte(storeInfo.Name),
			CommitID: &proof.CommitID{
				Version: uint64(storeInfo.CommitId.Version),
				Hash:    storeInfo.CommitId.Hash,
			},
		}
	}
	return &proof.CommitInfo{
		Version:    uint64(commitInfo.Version),
		StoreInfos: storeInfos,
	}
}

// Takes a MutableTree, a key, and a flag for creating existence or absence proof and returns the
// appropriate merkle.Proof. Since this must be called after querying for the value, this function should never error
// Thus, it will panic on error rather than returning it
func getProofFromTree(tree *memiavl.Tree, key []byte, exists bool) proof.CommitmentOp {
	var (
		commitmentProof *ics23.CommitmentProof
		err             error
	)

	if exists {
		// value was found
		commitmentProof, err = tree.GetMembershipProof(key)
		if err != nil {
			// sanity check: If value was found, membership proof must be creatable
			panic(fmt.Sprintf("unexpected value for empty proof: %s", err.Error()))
		}
	} else {
		// value wasn't found
		commitmentProof, err = tree.GetNonMembershipProof(key)
		if err != nil {
			// sanity check: If value wasn't found, nonmembership proof must be creatable
			panic(fmt.Sprintf("unexpected error for nonexistence proof: %s", err.Error()))
		}
	}

	return proof.NewIAVLCommitmentOp(key, commitmentProof)
}
