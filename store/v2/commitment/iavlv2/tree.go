package iavlv2

import (
	"fmt"

	"github.com/cosmos/iavl/v2"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/core/log"
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
	log  log.Logger
	path string
}

func NewTree(
	cfg Config,
	dbOptions iavl.SqliteDbOptions,
	log log.Logger,
) (*Tree, error) {
	pool := iavl.NewNodePool()
	sql, err := iavl.NewSqliteDb(pool, dbOptions)
	if err != nil {
		return nil, err
	}
	tree := iavl.NewTree(sql, pool, cfg.ToTreeOptions())
	return &Tree{tree: tree, log: log, path: dbOptions.Path}, nil
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

	return t.tree.LoadVersion(int64(version))
}

func (t *Tree) LoadVersionForOverwriting(version uint64) error {
	return t.LoadVersion(version)
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
	v := int64(version)
	h := t.tree.Version()
	if v > h {
		return nil, fmt.Errorf("get: cannot read future version %d; h: %d path=%s", v, h, t.path)
	}
	versionFound, val, err := t.tree.GetRecent(v, key)
	if versionFound {
		return val, err
	}
	cloned, err := t.tree.ReadonlyClone()
	if err != nil {
		return nil, err
	}
	if err = cloned.LoadVersion(int64(version)); err != nil {
		return nil, err
	}
	return cloned.Get(key)
}

func (t *Tree) Has(version uint64, key []byte) (bool, error) {
	res, err := t.Get(version, key)
	return res != nil, err
}

func (t *Tree) Iterator(version uint64, start, end []byte, ascending bool) (corestore.Iterator, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	h := t.tree.Version()
	v := int64(version)
	if v > h {
		return nil, fmt.Errorf("iterator: cannot read future version %d; h: %d", v, h)
	}
	ok, itr := t.tree.IterateRecent(v, start, end, ascending)
	if ok {
		return itr, nil
	}
	cloned, err := t.tree.ReadonlyClone()
	if err != nil {
		return nil, err
	}
	if err = cloned.LoadVersion(int64(version)); err != nil {
		return nil, err
	}
	if ascending {
		// inclusive = false is IAVL v1's default behavior.
		// the read expectations of certain modules (like x/staking) will cause a panic if this is changed.
		return t.tree.Iterator(start, end, false)
	} else {
		return t.tree.ReverseIterator(start, end)
	}
}

func (t *Tree) Export(version uint64) (commitment.Exporter, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	e, err := t.tree.Export(int64(version), iavl.PostOrder)
	if err != nil {
		return nil, err
	}
	return &Exporter{e}, nil
}

func (t *Tree) Import(version uint64) (commitment.Importer, error) {
	if err := isHighBitSet(version); err != nil {
		return nil, err
	}
	importer, err := t.tree.Import(int64(version))
	if err != nil {
		return nil, err
	}
	return &Importer{importer}, nil
}

func (t *Tree) Close() error {
	return t.tree.Close()
}

func (t *Tree) Prune(version uint64) error {
	// do nothing, IAVL v2 has its own advanced pruning mechanism
	return nil
}

// PausePruning is unnecessary in IAVL v2 due to the advanced pruning mechanism
func (t *Tree) PausePruning(bool) {}

func (t *Tree) IsConcurrentSafe() bool {
	return true
}

func (t *Tree) WorkingHash() []byte {
	return t.tree.Hash()
}

func isHighBitSet(version uint64) error {
	if version&(1<<63) != 0 {
		return fmt.Errorf("%d too large; uint64 with the highest bit set are not supported", version)
	}
	return nil
}

func SetGlobalPruneLimit(limit int) {
	iavl.SetGlobalPruneLimit(limit)
}
