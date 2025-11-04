package cachemulti

import (
	"fmt"
	"io"
	"maps"

	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
)

// storeNameCtxKey is the TraceContext metadata key that identifies
// the store which emitted a given trace.
const storeNameCtxKey = "store_name"

//----------------------------------------
// Store

// Store holds many branched stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	stores map[types.StoreKey]types.CacheWrap

	traceWriter  io.Writer
	traceContext types.TraceContext
	parentStore  func(types.StoreKey) types.CacheWrapper
}

var _ types.CacheMultiStore = Store{}

// NewFromKVStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects and a KVStore as the database. Each CacheWrapper store
// is a branched store.
func NewFromKVStore(
	stores map[types.StoreKey]types.CacheWrapper,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	cms := Store{
		stores:       make(map[types.StoreKey]types.CacheWrap, len(stores)),
		traceWriter:  traceWriter,
		traceContext: traceContext,
	}

	for key, store := range stores {
		cms.initStore(key, store)
	}

	return cms
}

// NewStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects. Each CacheWrapper store is a branched store.
func NewStore(
	stores map[types.StoreKey]types.CacheWrapper,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	return NewFromKVStore(stores, traceWriter, traceContext)
}

// NewFromParent constructs a cache multistore with a parent store lazily,
// the parent is usually another cache multistore or the block-stm multiversion store.
func NewFromParent(
	parentStore func(types.StoreKey) types.CacheWrapper,
	traceWriter io.Writer, traceContext types.TraceContext,
) Store {
	return Store{
		stores:       make(map[types.StoreKey]types.CacheWrap),
		traceWriter:  traceWriter,
		traceContext: traceContext,
		parentStore:  parentStore,
	}
}

func (cms Store) initStore(key types.StoreKey, store types.CacheWrapper) types.CacheWrap {
	if cms.TracingEnabled() {
		// only support tracing on KVStore.
		if kvstore, ok := store.(types.KVStore); ok {
			tctx := cms.traceContext.Clone().Merge(types.TraceContext{
				storeNameCtxKey: key.Name(),
			})
			store = tracekv.NewStore(kvstore, cms.traceWriter, tctx)
		}
	}
	cache := store.CacheWrap()
	cms.stores[key] = cache
	return cache
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms Store) SetTracer(w io.Writer) types.MultiStore {
	cms.traceWriter = w
	return cms
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms Store) SetTracingContext(tc types.TraceContext) types.MultiStore {
	if cms.traceContext != nil {
		maps.Copy(cms.traceContext, tc)
	} else {
		cms.traceContext = tc
	}

	return cms
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (cms Store) TracingEnabled() bool {
	return cms.traceWriter != nil
}

// LatestVersion returns the branch version of the store
func (cms Store) LatestVersion() int64 {
	panic("cannot get latest version from branch cached multi-store")
}

// GetStoreType returns the type of the store.
func (cms Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Write calls Write on each underlying store.
func (cms Store) Write() {
	for _, store := range cms.stores {
		store.Write()
	}
}

// CacheWrap implements CacheWrapper, returns the cache multi-store as a CacheWrap.
func (cms Store) CacheWrap() types.CacheWrap {
	return cms.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return cms.CacheWrap()
}

// CacheMultiStore implements MultiStore, returns a new CacheMultiStore from the
// underlying CacheMultiStore.
func (cms Store) CacheMultiStore() types.CacheMultiStore {
	return NewFromParent(cms.getCacheWrapper, cms.traceWriter, cms.traceContext)
}

// CacheMultiStoreWithVersion implements the MultiStore interface. It will panic
// as an already cached multi-store cannot load previous versions.
//
// TODO: The store implementation can possibly be modified to support this as it
// seems safe to load previous versions (heights).
func (cms Store) CacheMultiStoreWithVersion(_ int64) (types.CacheMultiStore, error) {
	panic("cannot branch cached multi-store with a version")
}

func (cms Store) getCacheWrapper(key types.StoreKey) types.CacheWrapper {
	store, ok := cms.stores[key]
	if !ok && cms.parentStore != nil {
		// load on demand
		store = cms.initStore(key, cms.parentStore(key))
	}
	if key == nil || store == nil {
		panic(fmt.Sprintf("kv store with key %v has not been registered in stores", key))
	}
	return store
}

// GetStore returns an underlying Store by key.
func (cms Store) GetStore(key types.StoreKey) types.Store {
	store, ok := cms.getCacheWrapper(key).(types.Store)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not Store", key))
	}
	return store
}

// GetKVStore returns an underlying KVStore by key.
func (cms Store) GetKVStore(key types.StoreKey) types.KVStore {
	store, ok := cms.getCacheWrapper(key).(types.KVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not KVStore", key))
	}
	return store
}

// GetObjKVStore returns an underlying KVStore by key.
func (cms Store) GetObjKVStore(key types.StoreKey) types.ObjKVStore {
	store, ok := cms.getCacheWrapper(key).(types.ObjKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not ObjKVStore", key))
	}
	return store
}
