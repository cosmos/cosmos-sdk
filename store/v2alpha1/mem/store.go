package mem

import (
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/dbadapter"
)

var (
	_ types.KVStore   = (*Store)(nil)
	_ types.Committer = (*Store)(nil)
)

// Store implements an in-memory only KVStore. Entries are persisted between
// commits and thus between blocks. State in Memory store is not committed as part of app state but maintained privately by each node
type Store struct {
	dbadapter.Store
	conn dbm.DBConnection
}

// NewStore constructs a new in-memory store.
func NewStore() *Store {
	db := memdb.NewDB()
	return &Store{
		Store: dbadapter.Store{DB: db.ReadWriter()},
		conn:  db,
	}
}

// GetStoreType returns the Store's type.
func (s Store) GetStoreType() types.StoreType {
	return types.StoreTypeMemory
}

// Commit commits to the underlying DB.
func (s *Store) Commit() (id types.CommitID) {
	return
}

func (s *Store) SetPruning(*types.PruningOptions) {}
func (s *Store) GetPruning() *types.PruningOptions        { return &types.PruningOptions{} }

func (s Store) LastCommitID() (id types.CommitID) { return }
