package cachemulti

import (
	"io"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/cache"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

//----------------------------------------
// Store

// Store holds many cache-wrapped stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	db         types.CacheKVStore
	stores     map[types.StoreKey]types.CacheWrap
	keysByName map[string]types.StoreKey

	traceWriter  io.Writer
	traceContext types.TraceContext
}

var _ types.CacheMultiStore = Store{}

func NewFromKVStore(store types.KVStore, stores map[types.StoreKey]types.CacheWrapper, keysByName map[string]types.StoreKey, traceWriter io.Writer, traceContext types.TraceContext) Store {
	cms := Store{
		db:           cache.NewStore(store),
		stores:       make(map[types.StoreKey]types.CacheWrap, len(stores)),
		keysByName:   keysByName,
		traceWriter:  traceWriter,
		traceContext: traceContext,
	}

	for key, store := range stores {
		if cms.TracingEnabled() {
			cms.stores[key] = store.CacheWrapWithTrace(cms.traceWriter, cms.traceContext)
		} else {
			cms.stores[key] = store.CacheWrap()
		}
	}

	return cms
}

func NewStore(db dbm.DB, stores map[types.StoreKey]types.CacheWrapper, keysByName map[string]types.StoreKey, traceWriter io.Writer, traceContext types.TraceContext) Store {
	return NewFromKVStore(dbadapter.Store{db}, stores, keysByName, traceWriter, traceContext)
}

func newCacheMultiStoreFromCMS(cms Store) Store {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range cms.stores {
		stores[k] = v
	}
	return NewFromKVStore(cms.db, stores, nil, cms.traceWriter, cms.traceContext)
}

// WithTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms Store) WithTracer(w io.Writer) types.MultiStore {
	cms.traceWriter = w
	return cms
}

// WithTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms Store) WithTracingContext(tc types.TraceContext) types.MultiStore {
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
func (cms Store) TracingEnabled() bool {
	return cms.traceWriter != nil
}

// ResetTraceContext resets the current tracing context.
func (cms Store) ResetTraceContext() types.MultiStore {
	cms.traceContext = nil
	return cms
}

// Implements Store.
func (cms Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Implements CacheMultiStore.
func (cms Store) Write() {
	cms.db.Write()
	for _, store := range cms.stores {
		store.Write()
	}
}

// Implements CacheWrapper.
func (cms Store) CacheWrap() types.CacheWrap {
	return cms.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return cms.CacheWrap()
}

// Implements MultiStore.
func (cms Store) CacheMultiStore() types.CacheMultiStore {
	return newCacheMultiStoreFromCMS(cms)
}

// Implements MultiStore.
func (cms Store) GetStore(key types.StoreKey) types.Store {
	return cms.stores[key].(types.Store)
}

// Implements MultiStore.
func (cms Store) GetKVStore(key types.StoreKey) types.KVStore {
	return cms.stores[key].(types.KVStore)
}
