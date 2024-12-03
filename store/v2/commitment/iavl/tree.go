package iavl

import (
	"fmt"

	"github.com/cosmos/iavl"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
)

var (
	_ commitment.Tree      = (*IavlTree)(nil)
	_ commitment.Reader    = (*IavlTree)(nil)
	_ store.PausablePruner = (*IavlTree)(nil)
)

// IavlTree is a wrapper around iavl.MutableTree.
type IavlTree struct {
	tree *iavl.MutableTree
}

// NewIavlTree creates a new IavlTree instance.
func NewIavlTree(db corestore.KVStoreWithBatch, logger log.Logger, cfg *Config) *IavlTree {
	tree := iavl.NewMutableTree(db, cfg.CacheSize, cfg.SkipFastStorageUpgrade, logger, iavl.AsyncPruningOption(true))
	return &IavlTree{
		tree: tree,
	}
}

// Remove removes the given key from the tree.
func (t *IavlTree) Remove(key []byte) error {
	_, _, err := t.tree.Remove(key)
	if err != nil {
		return err
	}
	return nil
}

// Set sets the given key-value pair in the tree.
func (t *IavlTree) Set(key, value []byte) error {
	_, err := t.tree.Set(key, value)
	return err
}

// Hash returns the hash of the latest saved version of the tree.
func (t *IavlTree) Hash() []byte {
	return t.tree.Hash()
}

// Version returns the current version of the tree.
func (t *IavlTree) Version() uint64 {
	return uint64(t.tree.Version())
}

// WorkingHash returns the working hash of the tree.
// Danger! iavl.MutableTree.WorkingHash() is a mutating operation!
// It advances the tree version by 1.
func (t *IavlTree) WorkingHash() []byte {
	return t.tree.WorkingHash()
}

// LoadVersion loads the state at the given version.
func (t *IavlTree) LoadVersion(version uint64) error {
	_, err := t.tree.LoadVersion(int64(version))
	return err
}

// LoadVersionForOverwriting loads the state at the given version.
// Any versions greater than targetVersion will be deleted.
func (t *IavlTree) LoadVersionForOverwriting(version uint64) error {
	return t.tree.LoadVersionForOverwriting(int64(version))
}

// Commit commits the current state to the tree.
func (t *IavlTree) Commit() ([]byte, uint64, error) {
	hash, v, err := t.tree.SaveVersion()
	return hash, uint64(v), err
}

// GetProof returns a proof for the given key and version.
func (t *IavlTree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	// the mutable tree is empty at genesis & when the storekey is removed, but the immutable tree is not but the immutable tree is not empty when the storekey is removed
	// by checking the latest version we can determine if we are in genesis or have a key that has been removed
	lv, err := t.tree.GetLatestVersion()
	if err != nil {
		return nil, err
	}
	if lv == 0 {
		return t.tree.GetProof(key)
	}

	immutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tree at version %d: %w", version, err)
	}

	return immutableTree.GetProof(key)
}

// Get implements the Reader interface.
func (t *IavlTree) Get(version uint64, key []byte) ([]byte, error) {
	// the mutable tree is empty at genesis & when the storekey is removed, but the immutable tree is not but the immutable tree is not empty when the storekey is removed
	// by checking the latest version we can determine if we are in genesis or have a key that has been removed
	lv, err := t.tree.GetLatestVersion()
	if err != nil {
		return nil, err
	}
	if lv == 0 {
		return t.tree.Get(key)
	}

	immutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tree at version %d: %w", version, err)
	}

	return immutableTree.Get(key)
}

// Iterator implements the Reader interface.
func (t *IavlTree) Iterator(version uint64, start, end []byte, ascending bool) (corestore.Iterator, error) {
	// the mutable tree is empty at genesis & when the storekey is removed, but the immutable tree is not empty when the storekey is removed
	// by checking the latest version we can determine if we are in genesis or have a key that has been removed
	lv, err := t.tree.GetLatestVersion()
	if err != nil {
		return nil, err
	}
	if lv == 0 {
		return t.tree.Iterator(start, end, ascending)
	}

	immutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tree at version %d: %w", version, err)
	}

	return immutableTree.Iterator(start, end, ascending)
}

// GetLatestVersion returns the latest version of the tree.
func (t *IavlTree) GetLatestVersion() (uint64, error) {
	v, err := t.tree.GetLatestVersion()
	return uint64(v), err
}

// SetInitialVersion sets the initial version of the database.
func (t *IavlTree) SetInitialVersion(version uint64) error {
	t.tree.SetInitialVersion(version)
	return nil
}

// Prune prunes all versions up to and including the provided version.
func (t *IavlTree) Prune(version uint64) error {
	return t.tree.DeleteVersionsTo(int64(version))
}

// PausePruning pauses the pruning process.
func (t *IavlTree) PausePruning(pause bool) {
	if pause {
		t.tree.SetCommitting()
	} else {
		t.tree.UnsetCommitting()
	}
}

// Export exports the tree exporter at the given version.
func (t *IavlTree) Export(version uint64) (commitment.Exporter, error) {
	tree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, err
	}
	exporter, err := tree.Export()
	if err != nil {
		return nil, err
	}

	return &Exporter{
		exporter: exporter,
	}, nil
}

// Import imports the tree importer at the given version.
func (t *IavlTree) Import(version uint64) (commitment.Importer, error) {
	importer, err := t.tree.Import(int64(version))
	if err != nil {
		return nil, err
	}

	return &Importer{
		importer: importer,
	}, nil
}

// Close closes the iavl tree.
func (t *IavlTree) Close() error {
	return t.tree.Close()
}
