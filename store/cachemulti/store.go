package cachemulti

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/trace"
)

//----------------------------------------
// Store

// Store holds many cache-wrapped stores.
// Implements MultiStore.
type Store struct {
	db         sdk.CacheKVStore
	stores     map[sdk.StoreKey]sdk.CacheKVStore
	keysByName map[string]sdk.StoreKey

	tracer *sdk.Tracer
}

var _ sdk.CacheMultiStore = Store{}

func NewStore(db dbm.DB, keysByName map[string]sdk.StoreKey, stores map[sdk.StoreKey]sdk.CommitKVStore, tracer *sdk.Tracer) Store {
	cms := Store{
		db:         cache.NewStore(dbadapter.NewStore(db)),
		stores:     make(map[sdk.StoreKey]sdk.CacheKVStore, len(stores)),
		keysByName: keysByName,
		tracer:     tracer,
	}

	for key, store := range stores {
		if tracer.Enabled() {
			cms.stores[key] = cache.NewStore(trace.NewStore(store, tracer))
		} else {
			cms.stores[key] = cache.NewStore(store)
		}
	}

	return cms
}

func newCacheMultiStoreFromCMS(cms Store) Store {
	cms2 := Store{
		db:     cache.NewStore(cms.db),
		stores: make(map[sdk.StoreKey]sdk.CacheKVStore, len(cms.stores)),
		tracer: cms.tracer,
	}

	for key, store := range cms.stores {
		if cms2.tracer.Enabled() {
			cms2.stores[key] = cache.NewStore(trace.NewStore(store, cms.tracer))
		} else {
			cms2.stores[key] = cache.NewStore(store)
		}
	}

	return cms2
}

// Implements MultiStore
func (cms Store) GetTracer() *sdk.Tracer {
	return cms.tracer
}

// Implements CacheMultiStore.
func (cms Store) Write() {
	cms.db.Write()
	for _, store := range cms.stores {
		store.Write()
	}
}

// Implements MultiStore.
func (cms Store) CacheWrap() sdk.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements MultiStore.
func (cms Store) GetKVStore(key sdk.StoreKey) sdk.KVStore {
	return cms.stores[key].(sdk.KVStore)
}
