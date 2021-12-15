package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormhooks"
)

type indexCommitmentStore struct {
	commitment kvstore.Writer
	index      kvstore.Writer
	hooks      ormhooks.Hooks
}

var _ kvstore.IndexCommitmentStore = &indexCommitmentStore{}

func (i indexCommitmentStore) CommitmentStoreReader() kvstore.Reader {
	return i.commitment
}

func (i indexCommitmentStore) IndexStoreReader() kvstore.Reader {
	return i.index
}

func (i indexCommitmentStore) CommitmentStore() kvstore.Store {
	return i.commitment
}

func (i indexCommitmentStore) IndexStore() kvstore.Store {
	return i.index
}

func (i indexCommitmentStore) ORMHooks() ormhooks.Hooks {
	return i.hooks
}
