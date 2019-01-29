package dbadapter

import (
	"io"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Wrapper type for dbm.Db with implementation of KVStore
type Store struct {
	dbm.DB
}

// GetStoreType returns the type of the store.
func (Store) GetStoreType() types.StoreType {
	return types.StoreTypeDB
}

// CacheWrap cache wraps the underlying store.
func (dsa Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(dsa)
}

// CacheWrapWithTrace implements KVStore.
func (dsa Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(dsa, w, tc))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ types.KVStore = Store{}
