package dbadapter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cache"
)

// Wrapper type for dbm.Db with implementation of KVStore
type Store struct {
	dbm.DB
}

func NewStore(parent dbm.DB) Store {
	return Store{parent}
}

// Implements KVStore.
func (dsa Store) CacheWrap() sdk.CacheKVStore {
	return cache.NewStore(dsa)
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ sdk.KVStore = Store{}
