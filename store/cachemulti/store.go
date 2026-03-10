package cachemulti

import (
	"fmt"
	"io"

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

	parentStore func(types.StoreKey) types.CacheWrapper
}

var _ types.CacheMultiStore = Store{}

// NewFromKVStore creates a new Store object from a mapping of store keys to
// CacheWrapper objects and a KVStore as the database. Each CacheWrapper store
// is a branched store.
func NewFromKVStore(
	stores map[types.StoreKey]types.CacheWrapper,
) Store {
	cms := Store{
		stores: make(map[types.StoreKey]types.CacheWrap, len(stores)),
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
) Store {
	return NewFromKVStore(stores)
}

// NewFromParent constructs a cache multistore with a parent store lazily,
// the parent is usually another cache multistore or the block-stm multiversion store.
func NewFromParent(
	parentStore func(types.StoreKey) types.CacheWrapper,
) Store {
	return Store{
		stores:      make(map[types.StoreKey]types.CacheWrap),
		parentStore: parentStore,
	}
}

func (cms Store) initStore(key types.StoreKey, store types.CacheWrapper) types.CacheWrap {
	cache := store.CacheWrap()
	cms.stores[key] = cache
	return cache
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
	return NewFromParent(cms.getCacheWrapper)
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
