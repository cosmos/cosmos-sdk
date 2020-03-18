package mem

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/types"
)

var (
	_ types.KVStore   = (*Store)(nil)
	_ types.Committer = (*Store)(nil)
)

// Store implements an in-memory only KVStore. Entries are peresistent between
// commits and thus between blocks.
type Store struct {
	dbadapter.Store
}

func NewStore() *Store {
	return NewStoreWithDB(dbm.NewMemDB())
}

func NewStoreWithDB(db *dbm.MemDB) *Store {
	return &Store{Store: dbadapter.Store{DB: db}}
}

// GetStoreType returns the Store's type.
func (ts *Store) GetStoreType() types.StoreType {
	return types.StoreTypeMemory
}

// Commit performs a no-op as entries are persistent between commitments.
func (ts *Store) Commit() (id types.CommitID) { return }

// nolint
func (ts *Store) SetPruning(pruning types.PruningOptions) {}
func (ts *Store) LastCommitID() (id types.CommitID)       { return }
