package dbadapter

import (
	"io"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/trace"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Wrapper type for dbm.Db with implementation of KVStore
type Store struct {
	dbm.DB
}

// Implements Store.
func (Store) GetStoreType() types.StoreType {
	return types.StoreTypeDB
}

// Implements KVStore.
func (dsa Store) CacheWrap() types.CacheWrap {
	return cache.NewStore(dsa)
}

// CacheWrapWithTrace implements the KVStore interface.
func (dsa Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cache.NewStore(trace.NewStore(dsa, w, tc))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ types.KVStore = Store{}
