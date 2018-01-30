package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db         CacheKVStore
	stores     map[StoreKey]CacheWrap
	keysByName map[string]StoreKey
}

var _ CacheMultiStore = cacheMultiStore{}

func newCacheMultiStoreFromRMS(rms *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:         NewCacheKVStore(dbStoreAdapter{rms.db}),
		stores:     make(map[StoreKey]CacheWrap, len(rms.stores)),
		keysByName: rms.keysByName,
	}
	for key, store := range rms.stores {
		cms.stores[key] = store.CacheWrap()
	}
	return cms
}

func newCacheMultiStoreFromCMS(cms cacheMultiStore) cacheMultiStore {
	cms2 := cacheMultiStore{
		db:     NewCacheKVStore(cms.db),
		stores: make(map[StoreKey]CacheWrap, len(cms.stores)),
	}
	for key, store := range cms.stores {
		cms2.stores[key] = store.CacheWrap()
	}
	return cms2
}

// Implements Store.
func (cms cacheMultiStore) GetStoreType() StoreType {
	return sdk.StoreTypeMulti
}

// Implements CacheMultiStore.
func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for _, store := range cms.stores {
		store.Write()
	}
}

// Implements CacheWrapper.
func (cms cacheMultiStore) CacheWrap() CacheWrap {
	return cms.CacheMultiStore().(CacheWrap)
}

// Implements MultiStore.
func (cms cacheMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements MultiStore.
func (cms cacheMultiStore) GetStore(key StoreKey) Store {
	return cms.stores[key].(Store)
}

// Implements MultiStore.
func (cms cacheMultiStore) GetKVStore(key StoreKey) KVStore {
	return cms.stores[key].(KVStore)
}

// GetStoreByName will first convert the original name to
// a special key, before looking up the CommitStore.
// This is not exposed to the extensions (which will need the
// StoreKey), but is useful in main, and particularly app.Query,
// in order to convert human strings into CommitStores.
func (cms cacheMultiStore) GetStoreByName(name string) Store {
	key := cms.keysByName[name]
	if key == nil {
		return nil
	}
	return cms.stores[key].(Store)
}
