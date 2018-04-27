package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tmlibs/db"
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

func (dsa dbStoreAdapter) SubspaceIterator(prefix []byte) Iterator {
	return dsa.Iterator(prefix, sdk.PrefixEndBytes(prefix))
}

func (dsa dbStoreAdapter) ReverseSubspaceIterator(prefix []byte) Iterator {
	return dsa.ReverseIterator(prefix, sdk.PrefixEndBytes(prefix))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbStoreAdapter{dbm.DB(nil)}
