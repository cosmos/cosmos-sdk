package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type indexCommitmentStoreWrapper struct {
	commitment kvstore.Store
	index      kvstore.Store
	store      kvstore.IndexCommitmentStore
}

func (i indexCommitmentStoreWrapper) Commit() error {
	return i.store.Commit()
}

func (i indexCommitmentStoreWrapper) Rollback() error {
	return i.store.Rollback()
}

// NewDebugIndexCommitmentStore wraps both stores from an IndexCommitmentStore
// with a debugger.
func NewDebugIndexCommitmentStore(store kvstore.IndexCommitmentStore, debugger Debugger) kvstore.IndexCommitmentStore {
	return &indexCommitmentStoreWrapper{
		store:      store,
		commitment: NewDebugStore(store.CommitmentStore(), debugger, "commit"),
		index:      NewDebugStore(store.IndexStore(), debugger, "index"),
	}

}

var _ kvstore.IndexCommitmentStore = &indexCommitmentStoreWrapper{}

func (i indexCommitmentStoreWrapper) ReadCommitmentStore() kvstore.ReadStore {
	return i.commitment
}

func (i indexCommitmentStoreWrapper) ReadIndexStore() kvstore.ReadStore {
	return i.index
}

func (i indexCommitmentStoreWrapper) CommitmentStore() kvstore.Store {
	return i.commitment
}

func (i indexCommitmentStoreWrapper) IndexStore() kvstore.Store {
	return i.index
}
