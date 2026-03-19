package blockstm

import (
	"sync"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

type (
	KVStorage    = GStorage[[]byte]
	ObjKVStorage = GStorage[any]

	KVCachedStorage    = GCachedStorage[[]byte]
	ObjKVCachedStorage = GCachedStorage[any]
)

var (
	_ KVStorage    = (*GCachedStorage[[]byte])(nil)
	_ ObjKVStorage = (*GCachedStorage[any])(nil)
)

// Storage is a common interface for KVStorage and ObjKVStorage.
type Storage interface {
	GetStoreType() storetypes.StoreType
}

// GStorage represents a read-only version of kv store,
// it provides the pre-state during the whole execution of block-stm.
// It can be implemented by either a plain KVStore, or a cached one.
type GStorage[V any] interface {
	Storage

	Get(key []byte) V

	Iterator(start, end []byte) storetypes.GIterator[V]
	ReverseIterator(start, end []byte) storetypes.GIterator[V]
}

// GCachedStorage wraps a plain kv store with a thread-safe cache, it's useful when we need to
// re-read pre-state values during validation.
//
// it'll bypass cache when doing iteration.
type GCachedStorage[V any] struct {
	GStorage[V]
	cache sync.Map
}

func NewGCachedStorage[V any](storage GStorage[V]) *GCachedStorage[V] {
	return &GCachedStorage[V]{GStorage: storage}
}

func (s *GCachedStorage[V]) Get(key []byte) V {
	if v, ok := s.cache.Load(string(key)); ok {
		return v.(V)
	}

	v := s.GStorage.Get(key)
	s.cache.Store(string(key), v)
	return v
}

// MultiStoreToCachedStorage convert MultiStore to an array of GStorage, wrap each store with a cache.
func MultiStoreToCachedStorage(ms MultiStore, stores map[storetypes.StoreKey]int) []Storage {
	storage := make([]Storage, len(stores))
	for key, i := range stores {
		store := ms.GetStore(key)
		switch v := store.(type) {
		case storetypes.KVStore:
			storage[i] = NewGCachedStorage(v)
		case storetypes.ObjKVStore:
			storage[i] = NewGCachedStorage(v)
		default:
			panic("unsupported store type")
		}
	}
	return storage
}

// MultiStoreToStorage convert MultiStore to an array of GStorage without cache.
func MultiStoreToStorage(ms MultiStore, stores map[storetypes.StoreKey]int) []Storage {
	storage := make([]Storage, len(stores))
	for key, i := range stores {
		store := ms.GetStore(key)
		switch v := store.(type) {
		case storetypes.KVStore:
			storage[i] = KVStorage(v)
		case storetypes.ObjKVStore:
			storage[i] = ObjKVStorage(v)
		default:
			panic("unsupported store type")
		}
	}
	return storage
}
