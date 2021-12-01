package memkv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
)

type IndexCommitmentStore struct {
	commitment dbm.DB
	index      dbm.DB
}

func NewIndexCommitmentStore() *IndexCommitmentStore {
	return &IndexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

var _ kv.IndexCommitmentStore = &IndexCommitmentStore{}

func (i IndexCommitmentStore) ReadCommitmentStore() kv.ReadStore {
	return i.commitment
}

func (i IndexCommitmentStore) ReadIndexStore() kv.ReadStore {
	return i.index
}

func (i IndexCommitmentStore) CommitmentStore() kv.Store {
	return i.commitment
}

func (i IndexCommitmentStore) IndexStore() kv.Store {
	return i.index
}
