package transient

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

var _ sdk.KVStore = (*Store)(nil)

// transientStore is a wrapper for a MemDB with Commiter implementation
type Store struct {
	dbadapter.Store
}

// Constructs new MemDB adapter
func NewStore() *Store {
	return &Store{dbadapter.Store{dbm.NewMemDB()}}
}

// Implements CommitStore
// Commit cleans up transientStore.
func (ts *Store) Commit() (id sdk.CommitID) {
	ts.Store = dbadapter.Store{dbm.NewMemDB()}
	return
}

// Implements CommitStore
func (ts *Store) SetPruning(pruning sdk.PruningStrategy) {
}

// Implements CommitStore
func (ts *Store) LastCommitID() (id sdk.CommitID) {
	return
}
