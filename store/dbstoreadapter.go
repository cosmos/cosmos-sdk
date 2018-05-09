package store

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	dbm "github.com/tendermint/tmlibs/db"
)

type dbStoreAdapter struct {
	dbm.DB
}

// Implements Store.
func (dbStoreAdapter) GetStoreType() StoreType {
	return baseapp.StoreTypeDB
}

// Implements KVStore.
func (dsa dbStoreAdapter) CacheWrap() CacheWrap {
	return NewCacheKVStore(dsa)
}

func (dsa dbStoreAdapter) SubspaceIterator(prefix []byte) Iterator {
	return dsa.Iterator(prefix, baseapp.PrefixEndBytes(prefix))
}

func (dsa dbStoreAdapter) ReverseSubspaceIterator(prefix []byte) Iterator {
	return dsa.ReverseIterator(prefix, baseapp.PrefixEndBytes(prefix))
}

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbStoreAdapter{dbm.DB(nil)}
