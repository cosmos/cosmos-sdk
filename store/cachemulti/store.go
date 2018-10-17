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

// Store holds many cache-wrapped kvstores.
// Implements MultiStore.
// TODO: support recursive multikvstores,
// currently only using CacheKVStores
type Store struct {
	db         types.CacheKVStore
	kvstores   map[types.KVStoreKey]types.CacheKVStore
	keysByName map[string]types.KVStoreKey

	tracer *types.Tracer
	tank   *types.GasTank
}

var _ types.CacheMultiStore = Store{}

func NewStore(db dbm.DB, keysByName map[string]types.KVStoreKey, kvstores map[types.KVStoreKey]types.CommitKVStore, tracer *types.Tracer, tank *types.GasTank) Store {
	cms := Store{
		db:         cache.NewStore(dbadapter.NewStore(db)),
		kvstores:   make(map[types.KVStoreKey]types.CacheKVStore, len(kvstores)),
		keysByName: keysByName,
		tracer:     tracer,
		tank:       tank,
	}

	for key, store := range kvstores {
		if tracer.Enabled() {
			cms.kvstores[key] = cache.NewStore(trace.NewStore(store, tracer))
		} else {
			cms.kvstores[key] = cache.NewStore(store)
		}
	}

	return cms
}

func newCacheMultiStoreFromCMS(cms Store) Store {
	cms2 := Store{
		db:       cache.NewStore(cms.db),
		kvstores: make(map[types.KVStoreKey]types.CacheKVStore, len(cms.kvstores)),
		tracer:   cms.tracer,
	}

	for key, store := range cms.kvstores {
		if cms2.tracer.Enabled() {
			cms2.kvstores[key] = cache.NewStore(trace.NewStore(store, cms.tracer))
		} else {
			cms2.kvstores[key] = cache.NewStore(store)
		}
	}

	return cms2
}

// Implements MultiStore
func (cms Store) GetTracer() *types.Tracer {
	return cms.tracer
}

// Implements MultiStore
func (cms Store) GetGasTank() *types.GasTank {
	return cms.tank
}

// Implements CacheMultiStore.
func (cms Store) Write() {
	cms.db.Write()
	for _, store := range cms.kvstores {
		store.Write()
	}
}

// Implements MultiStore.
func (cms Store) CacheWrap() types.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements MultiStore.
func (cms Store) GetKVStore(key types.KVStoreKey) types.KVStore {
	return cms.kvstores[key].(types.KVStore)
}
