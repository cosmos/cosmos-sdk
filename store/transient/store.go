package transient

import (
	"cosmossdk.io/store/internal"
	"cosmossdk.io/store/internal/btree"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/types"
)

var (
	_ types.Committer = (*Store)(nil)
	_ types.KVStore   = (*Store)(nil)

	_ types.Committer  = (*ObjStore)(nil)
	_ types.ObjKVStore = (*ObjStore)(nil)
)

// GStore is a wrapper for a MemDB with Committer implementation
type GStore[V any] struct {
	internal.BTreeStore[V]
}

// NewGStore constructs new generic transient store
func NewGStore[V any](isZero func(V) bool, valueLen func(V) int) *GStore[V] {
	return &GStore[V]{*internal.NewBTreeStore(btree.NewBTree[V](), isZero, valueLen)}
}

// Store specializes GStore for []byte
type Store struct {
	GStore[[]byte]
}

func NewStore() *Store {
	return &Store{*NewGStore(
		types.BytesIsZero,
		types.BytesValueLen,
	)}
}

func (*Store) GetStoreType() types.StoreType {
	return types.StoreTypeTransient
}

// ObjStore specializes GStore for any
type ObjStore struct {
	GStore[any]
}

func NewObjStore() *ObjStore {
	return &ObjStore{*NewGStore(
		types.AnyIsZero,
		types.AnyValueLen,
	)}
}

func (*ObjStore) GetStoreType() types.StoreType {
	return types.StoreTypeObject
}

// Commit cleans up Store.
// Implements CommitStore
func (ts *GStore[V]) Commit() (id types.CommitID) {
	ts.Clear()
	return id
}

func (ts *GStore[V]) SetPruning(_ pruningtypes.PruningOptions) {}

// GetPruning is a no-op as pruning options cannot be directly set on this store.
// They must be set on the root commit multi-store.
func (ts *GStore[V]) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningUndefined)
}

// LastCommitID implements CommitStore, returns empty CommitID.
func (ts *GStore[V]) LastCommitID() types.CommitID {
	return types.CommitID{}
}

func (ts *GStore[V]) WorkingHash() []byte {
	return []byte{}
}
