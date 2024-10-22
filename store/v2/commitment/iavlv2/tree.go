package iavlv2

import (
	"fmt"

	"github.com/cosmos/iavl/v2"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
)

var (
	_ commitment.Tree      = (*Tree)(nil)
	_ store.PausablePruner = (*Tree)(nil)
)

type Tree struct {
	tree    *iavl.Tree
	saveErr error
}

func NewTree(treeOptions iavl.TreeOptions, dbOptions iavl.SqliteDbOptions, pool *iavl.NodePool) (*Tree, error) {
	sql, err := iavl.NewSqliteDb(pool, dbOptions)
	if err != nil {
		return nil, err
	}
	tree := iavl.NewTree(sql, pool, treeOptions)
	return &Tree{tree: tree}, nil
}

func (t Tree) Set(key, value []byte) error {
	_, err := t.tree.Set(key, value)
	return err
}

func (t Tree) Remove(key []byte) error {
	_, _, err := t.tree.Remove(key)
	return err
}

func (t Tree) GetLatestVersion() (uint64, error) {
	return uint64(t.tree.Version()), nil
}

func (t Tree) Hash() []byte {
	return t.tree.Hash()
}

func (t Tree) WorkingHash() []byte {
	// iavl v2 and store/v2 have different working hash semantics so we must commit here
	var hash []byte
	hash, _, t.saveErr = t.tree.SaveVersion()
	return hash
}

func (t Tree) LoadVersion(version uint64) error {
	// TODO fix this in iavl v2
	if version == 0 {
		return nil
	}
	return t.tree.LoadVersion(int64(version))
}

func (t Tree) Commit() ([]byte, uint64, error) {
	return t.tree.Hash(), uint64(t.tree.Version()), t.saveErr
}

func (t Tree) SetInitialVersion(version uint64) error {
	return nil
}

func (t Tree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	//TODO implement me
	panic("implement me")
}

func (t Tree) Get(version uint64, key []byte) ([]byte, error) {
	if int64(version) != t.tree.Version() {
		return nil, fmt.Errorf("loading past version not yet supported")
	}
	return t.tree.Get(key)
}

func (t Tree) Export(version uint64) (commitment.Exporter, error) {
	//TODO implement me
	panic("implement me")
}

func (t Tree) Import(version uint64) (commitment.Importer, error) {
	//TODO implement me
	panic("implement me")
}

func (t Tree) Close() error {
	return t.tree.Close()
}

func (t Tree) Prune(version uint64) error {
	return t.tree.DeleteVersionsTo(int64(version))
}

// PausePruning is unnecessary in IAVL v2 due to the advanced pruning mechanism
func (t Tree) PausePruning(bool) {}
