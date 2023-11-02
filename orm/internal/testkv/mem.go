package testkv

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/orm/model/ormtable"
)

// NewSplitMemBackend returns a Backend instance
// which uses two separate memory stores to simulate behavior when there
// are really two separate backing stores.
func NewSplitMemBackend() ormtable.Backend {
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: dbm.NewMemDB(),
		IndexStore:      dbm.NewMemDB(),
	})
}

// NewSharedMemBackend returns a Backend instance
// which uses a single backing memory store to simulate legacy scenarios
// where only a single KV-store is available to modules.
func NewSharedMemBackend() ormtable.Backend {
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: dbm.NewMemDB(),
		// commit store is automatically used as the index store
	})
}
