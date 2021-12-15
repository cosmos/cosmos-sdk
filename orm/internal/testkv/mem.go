package testkv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

// NewSplitMemIndexCommitmentStore returns an IndexCommitmentStore instance
// which uses two separate memory stores to simulate behavior when there
// are really two separate backing stores.
func NewSplitMemIndexCommitmentStore() kvstore.IndexCommitmentStoreWithHooks {
	return &indexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

// NewSharedMemIndexCommitmentStore returns an IndexCommitmentStore instance
// which uses a single backing memory store to simulate legacy scenarios
// where only a single KV-store is available to modules.
func NewSharedMemIndexCommitmentStore() kvstore.IndexCommitmentStoreWithHooks {
	store := dbm.NewMemDB()
	return &indexCommitmentStore{
		commitment: store,
		index:      store,
	}
}
