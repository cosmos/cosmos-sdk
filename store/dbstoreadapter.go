package store

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type dbStoreAdapter struct {
	dbm.DB
}

// Implements Store.
func (dbStoreAdapter) GetStoreType() StoreType {
	return sdk.StoreTypeDB
}

// Implements KVStore.
func (dsa dbStoreAdapter) CacheWrap() CacheWrap {
	return NewCacheKVStore(dsa)
}

// CacheWrapWithTrace implements the KVStore interface.
func (dsa dbStoreAdapter) CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(dsa, w, tc))
}

// Implements KVStore
func (dsa dbStoreAdapter) Prefix(prefix []byte) KVStore {
	return prefixStore{dsa, prefix}
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbStoreAdapter{dbm.DB(nil)}
