package testkv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

func NewSplitMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	return &indexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

func NewSharedMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	store := dbm.NewMemDB()
	return &indexCommitmentStore{
		commitment: store,
		index:      store,
	}
}
