package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type indexCommitmentStore struct {
	commitment kvstore.Store
	index      kvstore.Store
}

// NewDebugIndexCommitmentStore wraps both stores from an IndexCommitmentStore
// with a debugger.
func NewDebugIndexCommitmentStore(store kvstore.IndexCommitmentStore, debugger Debugger) kvstore.IndexCommitmentStore {
	return &indexCommitmentStore{
		commitment: NewDebugStore(store.CommitmentStore(), debugger, "commit"),
		index:      NewDebugStore(store.IndexStore(), debugger, "index"),
	}

}

var _ kvstore.IndexCommitmentStore = &indexCommitmentStore{}

func (i indexCommitmentStore) ReadCommitmentStore() kvstore.ReadStore {
	return i.commitment
}

func (i indexCommitmentStore) ReadIndexStore() kvstore.ReadStore {
	return i.index
}

func (i indexCommitmentStore) CommitmentStore() kvstore.Store {
	return i.commitment
}

func (i indexCommitmentStore) IndexStore() kvstore.Store {
	return i.index
}
