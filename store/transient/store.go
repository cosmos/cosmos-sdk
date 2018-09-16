package transient

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

var _ types.KVStore = (*Store)(nil)

// transientStore is a wrapper for a MemDB with Commiter implementation
type Store struct {
	dbadapter.Store

	tank *types.GasTank
}

// Constructs new MemDB adapter
func NewStore() *Store {
	return &Store{dbadapter.Store{dbm.NewMemDB()}, new(types.GasTank)}
}

// Implements CommitStore
// Commit cleans up transientStore.
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
