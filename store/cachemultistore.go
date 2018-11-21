package store

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//----------------------------------------
// cacheMultiStore

// cacheMultiStore holds many cache-wrapped stores.
// Implements MultiStore.
// NOTE: a cacheMultiStore (and MultiStores in general) should never expose the
// keys for the substores.
type cacheMultiStore struct {
	db         CacheKVStore
	stores     map[StoreKey]CacheWrap
	keysByName map[string]StoreKey

	traceWriter  io.Writer
	traceContext TraceContext
}

var _ CacheMultiStore = cacheMultiStore{}

func newCacheMultiStoreFromRMS(rms *rootMultiStore) cacheMultiStore {
	cms := cacheMultiStore{
		db:           NewCacheKVStore(dbStoreAdapter{rms.db}),
		stores:       make(map[StoreKey]CacheWrap, len(rms.stores)),
		keysByName:   rms.keysByName,
		traceWriter:  rms.traceWriter,
		traceContext: rms.traceContext,
	}

	for key, store := range rms.stores {
		if cms.TracingEnabled() {
			cms.stores[key] = store.CacheWrapWithTrace(cms.traceWriter, cms.traceContext)
		} else {
			cms.stores[key] = store.CacheWrap()
		}
	}

	return cms
}

func newCacheMultiStoreFromCMS(cms cacheMultiStore) cacheMultiStore {
	cms2 := cacheMultiStore{
		db:           NewCacheKVStore(cms.db),
		stores:       make(map[StoreKey]CacheWrap, len(cms.stores)),
		traceWriter:  cms.traceWriter,
		traceContext: cms.traceContext,
	}

	for key, store := range cms.stores {
		if cms2.TracingEnabled() {
			cms2.stores[key] = store.CacheWrapWithTrace(cms2.traceWriter, cms2.traceContext)
		} else {
			cms2.stores[key] = store.CacheWrap()
		}
	}

	return cms2
}

// WithTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms cacheMultiStore) WithTracer(w io.Writer) MultiStore {
	cms.traceWriter = w
	return cms
}

// WithTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms cacheMultiStore) WithTracingContext(tc TraceContext) MultiStore {
	if cms.traceContext != nil {
		for k, v := range tc {
			cms.traceContext[k] = v
		}
	} else {
		cms.traceContext = tc
	}

	return cms
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (cms cacheMultiStore) TracingEnabled() bool {
	return cms.traceWriter != nil
}

// ResetTraceContext resets the current tracing context.
func (cms cacheMultiStore) ResetTraceContext() MultiStore {
	cms.traceContext = nil
	return cms
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

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms cacheMultiStore) CacheWrapWithTrace(_ io.Writer, _ TraceContext) CacheWrap {
	return cms.CacheWrap()
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

// Implements MultiStore.
func (cms cacheMultiStore) GetKVStoreWithGas(meter sdk.GasMeter, config sdk.GasConfig, key StoreKey) KVStore {
	return NewGasKVStore(meter, config, cms.GetKVStore(key))
}
