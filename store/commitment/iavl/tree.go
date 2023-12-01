package iavl

import (
	"errors"
	"fmt"
	"io"
	"math"

	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/iavl"
	ics23 "github.com/cosmos/ics23/go"

	log "cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
)

var _ commitment.Tree = (*IavlTree)(nil)

var _ snapshottypes.CommitSnapshotter = (*IavlTree)(nil)

// IavlTree is a wrapper around iavl.MutableTree.
type IavlTree struct {
	tree *iavl.MutableTree

	// storeKey is the identifier of the store.
	storeKey string
}

// NewIavlTree creates a new IavlTree instance.
func NewIavlTree(db dbm.DB, logger log.Logger, storeKey string, cfg *Config) *IavlTree {
	tree := iavl.NewMutableTree(db, cfg.CacheSize, cfg.SkipFastStorageUpgrade, logger)
	return &IavlTree{
		tree:     tree,
		storeKey: storeKey,
	}
}

// Remove removes the given key from the tree.
func (t *IavlTree) Remove(key []byte) error {
	_, res, err := t.tree.Remove(key)
	if !res {
		return fmt.Errorf("key %x not found", key)
	}
	return err
}

// Set sets the given key-value pair in the tree.
func (t *IavlTree) Set(key, value []byte) error {
	_, err := t.tree.Set(key, value)
	return err
}

// WorkingHash returns the working hash of the database.
func (t *IavlTree) WorkingHash() []byte {
	return t.tree.WorkingHash()
}

// LoadVersion loads the state at the given version.
func (t *IavlTree) LoadVersion(version uint64) error {
	return t.tree.LoadVersionForOverwriting(int64(version))
}

// Commit commits the current state to the database.
func (t *IavlTree) Commit() ([]byte, error) {
	hash, _, err := t.tree.SaveVersion()
	return hash, err
}

// GetProof returns a proof for the given key and version.
func (t *IavlTree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	imutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, err
	}

	return imutableTree.GetProof(key)
}

// GetLatestVersion returns the latest version of the database.
func (t *IavlTree) GetLatestVersion() uint64 {
	return uint64(t.tree.Version())
}

// Prune prunes all versions up to and including the provided version.
func (t *IavlTree) Prune(version uint64) error {
	return t.tree.DeleteVersionsTo(int64(version))
}

// Snapshot implements snapshottypes.CommitSnapshotter.
func (t *IavlTree) Snapshot(version uint64, protoWriter protoio.Writer) error {
	if version == 0 {
		return fmt.Errorf("the snapshot version must be greater than 0")
	}

	latestVersion := t.GetLatestVersion()
	if version > latestVersion {
		return fmt.Errorf("the snapshot version %d is greater than the latest version %d", version, latestVersion)
	}

	tree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return fmt.Errorf("failed to get immutable tree for version %d: %w", version, err)
	}

	exporter, err := tree.Export()
	if err != nil {
		return fmt.Errorf("failed to export tree for version %d: %w", version, err)
	}

	defer exporter.Close()

	err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
		Item: &snapshottypes.SnapshotItem_Store{
			Store: &snapshottypes.SnapshotStoreItem{
				Name: t.storeKey,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to write store name: %w", err)
	}

	for {
		node, err := exporter.Next()
		if errors.Is(err, iavl.ErrorExportDone) {
			break
		} else if err != nil {
			return fmt.Errorf("failed to get the next export node: %w", err)
		}

		if err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
			Item: &snapshottypes.SnapshotItem_IAVL{
				IAVL: &snapshottypes.SnapshotIAVLItem{
					Key:     node.Key,
					Value:   node.Value,
					Height:  int32(node.Height),
					Version: node.Version,
				},
			},
		}); err != nil {
			return fmt.Errorf("failed to write iavl node: %w", err)
		}
	}

	return nil
}

// Restore implements snapshottypes.CommitSnapshotter.
func (t *IavlTree) Restore(version uint64, format uint32, protoReader protoio.Reader, chStorage chan<- *store.KVPair) (snapshottypes.SnapshotItem, error) {
	var (
		importer     *iavl.Importer
		snapshotItem snapshottypes.SnapshotItem
	)

loop:
	for {
		snapshotItem = snapshottypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return snapshottypes.SnapshotItem{}, fmt.Errorf("invalid protobuf message: %w", err)
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshottypes.SnapshotItem_Store:
			t.storeKey = item.Store.Name
			importer, err = t.tree.Import(int64(version))
			if err != nil {
				return snapshottypes.SnapshotItem{}, fmt.Errorf("failed to import tree for version %d: %w", version, err)
			}
			defer importer.Close()

		case *snapshottypes.SnapshotItem_IAVL:
			if importer == nil {
				return snapshottypes.SnapshotItem{}, fmt.Errorf("received IAVL node item before store item")
			}
			if item.IAVL.Height > int32(math.MaxInt8) {
				return snapshottypes.SnapshotItem{}, fmt.Errorf("node height %v cannot exceed %v",
					item.IAVL.Height, math.MaxInt8)
			}
			node := &iavl.ExportNode{
				Key:     item.IAVL.Key,
				Value:   item.IAVL.Value,
				Height:  int8(item.IAVL.Height),
				Version: item.IAVL.Version,
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
					StoreKey: t.storeKey,
					Key:      node.Key,
					Value:    node.Value,
				}
			}
			err := importer.Add(node)
			if err != nil {
				return snapshottypes.SnapshotItem{}, fmt.Errorf("failed to add node to importer: %w", err)
			}
		default:
			break loop
		}
	}

	if importer != nil {
		err := importer.Commit()
		if err != nil {
			return snapshottypes.SnapshotItem{}, fmt.Errorf("failed to commit importer: %w", err)
		}
	}

	_, err := t.tree.LoadVersion(int64(version))

	return snapshotItem, err
}

// Close closes the iavl tree.
func (t *IavlTree) Close() error {
	return nil
}
