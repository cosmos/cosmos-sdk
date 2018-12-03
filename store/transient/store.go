package transient

import (
	"github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

var _ types.KVStore = (*Store)(nil)

// Store is a wrapper for a MemDB with Commiter implementation
type Store struct {
	dbadapter.Store
}

// Constructs new MemDB adapter
func NewStore() *Store {
	return &Store{dbadapter.Store{dbm.NewMemDB()}}
}

// Implements CommitStore
// Commit cleans up Store.
func (ts *Store) Commit() (id types.CommitID) {
	ts.Store = dbadapter.Store{dbm.NewMemDB()}
	return
}

// Implements CommitStore
func (ts *Store) SetPruning(pruning types.PruningStrategy) {
}

// Implements CommitStore
func (ts *Store) LastCommitID() (id types.CommitID) {
	return
}

// XXX: delete
/*
// Implements types.KVStore
func (ts *Store) Prefix(prefix []byte) types.KVStore {
	return prefixStore{ts, prefix}
}

// Implements types.KVStore
func (ts *Store) Gas(meter GasMeter, config GasConfig) types.KVStore {
	return NewGasKVStore(meter, config, ts)
}
*/
// Implements Store.
func (ts *Store) GetStoreType() types.StoreType {
	return types.StoreTypeTransient
}
