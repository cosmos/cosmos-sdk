package iavlmulti

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/cachemulti"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// iavlmulti.Store works similar with rootmulti.Store
// but stores
type Store struct {
	db dbm.DB

	kvstores map[types.KVStoreKey]iavl.KVStore
}

var _ types.CommitMultiStore = (*Store)(nil)

func (store *Store) CacheWrap() types.CacheMultiStore {
	return cachemulti.NewStore(store.db, store.keysByName, store.iavlstores)
}
