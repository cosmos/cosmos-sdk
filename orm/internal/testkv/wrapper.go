package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/model/ormhooks"
)

type backend struct {
	commitment kvstore.Writer
	index      kvstore.Writer
	hooks      ormhooks.Hooks
}

var _ kvstore.Backend = &backend{}

func (i backend) CommitmentStoreReader() kvstore.Reader {
	return i.commitment
}

func (i backend) IndexStoreReader() kvstore.Reader {
	return i.index
}

func (i backend) CommitmentStore() kvstore.Store {
	return i.commitment
}

func (i backend) IndexStore() kvstore.Store {
	return i.index
}

func (i backend) ORMHooks() ormhooks.Hooks {
	return i.hooks
}
