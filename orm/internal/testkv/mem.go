package testkv

import (
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	dbm "github.com/tendermint/tm-db"
)

func NewMemIndexCommitmentStore() kvstore.IndexCommitmentStore {
	return &indexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}
