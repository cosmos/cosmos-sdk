package dbadapter

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/trace"
)

// Wrapper type for dbm.Db with implementation of KVStore
type Store struct {
	dbm.DB
}

func NewStore(parent dbm.DB) Store {
	return Store{parent}
}

// Implements KVStore.
func (dsa Store) CacheWrap() sdk.CacheWrap {
	return cache.NewStore(dsa)
}

// CacheWrapWithTrace implements the KVStore interface.
func (dsa Store) CacheWrapWithTrace(w io.Writer, tc sdk.TraceContext) sdk.CacheWrap {
	return cache.NewStore(trace.NewStore(dsa, w, tc))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ sdk.KVStore = Store{}
