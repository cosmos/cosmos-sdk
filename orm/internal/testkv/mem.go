package testkv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// NewSplitMemBackend returns a Backend instance
// which uses two separate memory stores to simulate behavior when there
// are really two separate backing stores.
func NewSplitMemBackend() kvstore.Backend {
	return &backend{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

// NewSharedMemBackend returns a Backend instance
// which uses a single backing memory store to simulate legacy scenarios
// where only a single KV-store is available to modules.
func NewSharedMemBackend() kvstore.Backend {
	store := dbm.NewMemDB()
	return &backend{
		commitment: store,
		index:      store,
	}
}
