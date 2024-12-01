package iavlv2

import (
	"errors"
	"fmt"

	"github.com/cosmos/iavl/v2"
	ics23 "github.com/cosmos/ics23/go"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
)

var (
	_ commitment.Tree      = (*Tree)(nil)
	_ commitment.Reader    = (*Tree)(nil)
	_ store.PausablePruner = (*Tree)(nil)
)

type Tree struct {
	tree *iavl.Tree
}

func NewTree(treeOptions iavl.TreeOptions, dbOptions iavl.SqliteDbOptions, pool *iavl.NodePool) (*Tree, error) {
	sql, err := iavl.NewSqliteDb(pool, dbOptions)
	if err != nil {
		return nil, err
	}
	tree := iavl.NewTree(sql, pool, treeOptions)
	return &Tree{tree: tree}, nil
}

func (t *Tree) Set(key, value []byte) error {
	_, err := t.tree.Set(key, value)
	return err
}

func (t *Tree) Remove(key []byte) error {
	_, _, err := t.tree.Remove(key)
	return err
}

func (t *Tree) GetLatestVersion() (uint64, error) {
	return uint64(t.tree.Version()), nil
}

func (t *Tree) Hash() []byte {
	return t.tree.Hash()
}

func (t *Tree) Version() uint64 {
	return uint64(t.tree.Version())
}

func (t *Tree) LoadVersion(version uint64) error {
	if err := isHighBitSet(version); err != nil {
		return err
	}

	if version == 0 {
		return nil
	}
	return t.tree.LoadVersion(int64(version))
}

func (t *Tree) LoadVersionForOverwriting(version uint64) error {
	return t.LoadVersion(version) // TODO: implement overwriting
}

func (t *Tree) Commit() ([]byte, uint64, error) {
	h, v, err := t.tree.SaveVersion()
	return h, uint64(v), err
}

func (t *Tree) SetInitialVersion(version uint64) error {
	if err := isHighBitSet(version); err != nil {
		return err
	}
	t.tree.SetShouldCheckpoint()
	return t.tree.SetInitialVersion(int64(version))
}

func (t *Tree) GetProof(version uint64, key []byte) (*ics23.CommitmentProof, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	return t.tree.GetProof(int64(version), key)
}

func (t *Tree) Get(version uint64, key []byte) ([]byte, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	if int64(version) != t.tree.Version() {
		cloned, err := t.tree.ReadonlyClone()
		if err != nil {
			return nil, err
		}
		if err = cloned.LoadVersion(int64(version)); err != nil {
			return nil, err
		}
		return cloned.Get(key)
	} else {
		return t.tree.Get(key)
	}
}

func (t *Tree) Iterator(version uint64, start, end []byte, ascending bool) (corestore.Iterator, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	if int64(version) != t.tree.Version() {
		return nil, fmt.Errorf("loading past version not yet supported")
	}
	if ascending {
		return t.tree.Iterator(start, end, false)
	} else {
		return t.tree.ReverseIterator(start, end)
	}
}

func (t *Tree) Export(version uint64) (commitment.Exporter, error) {
	return nil, errors.New("snapshot import/export not yet supported")
}

func (t *Tree) Import(version uint64) (commitment.Importer, error) {
	return nil, errors.New("snapshot import/export not yet supported")
}

func (t *Tree) Close() error {
	return t.tree.Close()
}

func (t *Tree) Prune(version uint64) error {
	if err := isHighBitSet(version); err != nil {
		return err
	}

	return t.tree.DeleteVersionsTo(int64(version))
}

// PausePruning is unnecessary in IAVL v2 due to the advanced pruning mechanism
func (t *Tree) PausePruning(bool) {}

func (t *Tree) WorkingHash() []byte {
	return t.tree.Hash()
}

func isHighBitSet(version uint64) error {
	if version&(1<<63) != 0 {
		return fmt.Errorf("%d too large; uint64 with the highest bit set are not supported", version)
	}
	return nil
}
