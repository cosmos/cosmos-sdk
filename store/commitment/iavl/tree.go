package iavl

import (
	"fmt"

	"github.com/cosmos/iavl"
	ics23 "github.com/cosmos/ics23/go"

	log "cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
)

var _ commitment.Tree = (*IavlTree)(nil)

// IavlTree is a wrapper around iavl.MutableTree.
type IavlTree struct {
	tree *iavl.MutableTree
}

// NewIavlTree creates a new IavlTree instance.
func NewIavlTree(db store.RawDB, logger log.Logger, cfg *Config) *IavlTree {
	tree := iavl.NewMutableTree(dbm.NewWrapper(db), cfg.CacheSize, cfg.SkipFastStorageUpgrade, logger)
	return &IavlTree{
		tree: tree,
	}
}

// Remove removes the given key from the tree.
func (t *IavlTree) Remove(key []byte) error {
	_, res, err := t.tree.Remove(key)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("key %x not found", key)
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

// WorkingHash returns the working hash of the tree.
func (t *IavlTree) WorkingHash() []byte {
	return t.tree.WorkingHash()
}

// LoadVersion loads the state at the given version.
func (t *IavlTree) LoadVersion(version uint64) error {
	return t.tree.LoadVersionForOverwriting(int64(version))
}

// Commit commits the current state to the tree.
func (t *IavlTree) Commit() ([]byte, uint64, error) {
	hash, v, err := t.tree.SaveVersion()
	return hash, uint64(v), err
}

// GetProof returns a proof for the given key and version.
func (t *IavlTree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	immutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tree at version %d: %w", version, err)
	}

	return immutableTree.GetProof(key)
}

func (t *IavlTree) Get(version uint64, key []byte) ([]byte, error) {
	immutableTree, err := t.tree.GetImmutable(int64(version))
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tree at version %d: %w", version, err)
	}

	return immutableTree.Get(key)
}

// GetLatestVersion returns the latest version of the tree.
func (t *IavlTree) GetLatestVersion() uint64 {
	return uint64(t.tree.Version())
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
	return nil
}
