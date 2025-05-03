package cachemulti

import (
	"fmt"
	"io"

	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
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

	branched bool
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
	_ dbm.DB, stores map[types.StoreKey]types.CacheWrapper, _ map[string]types.StoreKey,
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

// GetStoreType returns the type of the store.
func (cms Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Write calls Write on each underlying store.
func (cms Store) Write() {
	if cms.branched {
		panic("cannot Write on branched store")
	}
	for _, store := range cms.stores {
		store.Write()
	}
}

func (cms Store) Discard() {
	for _, store := range cms.stores {
		store.Discard()
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
	return NewFromParent(cms.getCacheWrapper, cms.traceWriter, cms.traceContext)
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

func (cms Store) Clone() Store {
	stores := make(map[types.StoreKey]types.CacheWrap, len(cms.stores))
	for k, v := range cms.stores {
		stores[k] = v.(types.BranchStore).Clone().(types.CacheWrap)
	}
	return Store{
		stores:       stores,
		traceWriter:  cms.traceWriter,
		traceContext: cms.traceContext,
		parentStore:  cms.parentStore,

		branched: true,
	}
}

func (cms Store) Restore(other Store) {
	if !other.branched {
		panic("cannot restore from non-branched store")
	}

	// discard the non-exists stores
	for k, v := range cms.stores {
		if _, ok := other.stores[k]; !ok {
			// clear the cache store if it's not in the other
			v.Discard()
		}
	}

	// restore the other stores
	for k, v := range other.stores {
		store, ok := cms.stores[k]
		if !ok {
			store = cms.initStore(k, cms.parentStore(k))
		}

		store.(types.BranchStore).Restore(v.(types.BranchStore))
	}
}

func (cms Store) RunAtomic(cb func(types.CacheMultiStore) error) error {
	branch := cms.Clone()
	if err := cb(branch); err != nil {
		return err
	}

	cms.Restore(branch)
	return nil
}
