package mem

import (
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/db"
)

var _ commitment.Tree = (*Tree)(nil)

// Tree is a simple in-memory implementation of commitment.Tree.
type Tree struct {
	*db.MemDB
}

func (t *Tree) Remove(key []byte) error {
	return t.MemDB.Delete(key)
}

func (t *Tree) GetLatestVersion() (uint64, error) {
	return 0, nil
}

func (t *Tree) Hash() []byte {
	return nil
}

func (t *Tree) Version() uint64 {
	return 0
}

func (t *Tree) LoadVersion(version uint64) error {
	return nil
}

func (t *Tree) LoadVersionForOverwriting(version uint64) error {
	return nil
}

func (t *Tree) Commit() ([]byte, uint64, error) {
	return nil, 0, nil
}

func (t *Tree) SetInitialVersion(version uint64) error {
	return nil
}

func (t *Tree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	return nil, nil
}

func (t *Tree) Get(version uint64, key []byte) ([]byte, error) {
	return t.MemDB.Get(key)
}

func (t *Tree) Prune(version uint64) error {
	return nil
}

func (t *Tree) Export(version uint64) (commitment.Exporter, error) {
	return nil, nil
}

func (t *Tree) Import(version uint64) (commitment.Importer, error) {
	return nil, nil
}

func New() *Tree {
	return &Tree{MemDB: db.NewMemDB()}
}

func (t *Tree) IsConcurrentSafe() bool {
	return false
}
