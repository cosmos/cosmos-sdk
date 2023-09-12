package iavl

import (
	"fmt"

	log "cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/iavl"
	ics23 "github.com/cosmos/ics23/go"
)

var _ store.Tree = (*IavlTree)(nil)

// IavlTree is a wrapper around iavl.MutableTree.
type IavlTree struct {
	tree *iavl.MutableTree
}

// NewIavlTree creates a new IavlTree instance.
func NewIavlTree(db dbm.DB, logger log.Logger, cfg *Config) *IavlTree {
	tree := iavl.NewMutableTree(db, cfg.CacheSize, cfg.SkipFastStorageUpgrade, logger)
	return &IavlTree{
		tree: tree,
	}
}

// WriteBatch writes a batch of key-value pairs to the database.
func (t *IavlTree) WriteBatch(cs *store.ChangeSet) error {
	for _, kv := range cs.Pairs {
		if kv.Value == nil {
			_, res, err := t.tree.Remove(kv.Key)
			if err != nil {
				return err
			}
			if !res {
				return fmt.Errorf("failed to delete key %X", kv.Key)
			}
		} else {
			_, err := t.tree.Set(kv.Key, kv.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

// Close closes the iavl tree.
func (t *IavlTree) Close() error {
	return nil
}
