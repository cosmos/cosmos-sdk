package cachemulti

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/trace"
	"github.com/cosmos/cosmos-sdk/store/types"
)

//----------------------------------------
// Store

// Store holds many cache-wrapped stores.
// Implements MultiStore.
type Store struct {
	db         types.CacheKVStore
	stores     map[types.StoreKey]types.CacheKVStore
	keysByName map[string]types.StoreKey

	tracer *types.Tracer
	tank   *types.GasTank
}

var _ types.CacheMultiStore = Store{}

func NewStore(db dbm.DB, keysByName map[string]types.StoreKey, stores map[types.StoreKey]types.CommitKVStore, tracer *types.Tracer, tank *types.GasTank) Store {
	cms := Store{
		db:         cache.NewStore(dbadapter.NewStore(db)),
		stores:     make(map[types.StoreKey]types.CacheKVStore, len(stores)),
		keysByName: keysByName,
		tracer:     tracer,
		tank:       tank,
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
		stores: make(map[types.StoreKey]types.CacheKVStore, len(cms.stores)),
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
func (cms Store) GetTracer() *types.Tracer {
	return cms.tracer
}

func (cms Store) GetGasTank() *types.GasTank {
	return cms.tank
}

// Implements CacheMultiStore.
func (cms Store) Write() {
	cms.db.Write()
	for _, store := range cms.stores {
		store.Write()
	}
}

// Implements MultiStore.
func (cms Store) CacheWrap() types.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements MultiStore.
func (cms Store) GetKVStore(key types.StoreKey) types.KVStore {
	return cms.stores[key].(types.KVStore)
}
