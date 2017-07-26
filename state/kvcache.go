package state

import "container/list"

// KVCache is a cache that enforces deterministic sync order.
type KVCache struct {
	store KVStore
	cache map[string]kvCacheValue
	keys  *list.List
}

type kvCacheValue struct {
	v []byte        // The value of some key
	e *list.Element // The KVCache.keys element
}

// NOTE: If store is nil, creates a new MemKVStore
func NewKVCache(store KVStore) *KVCache {
	if store == nil {
		store = NewMemKVStore()
	}
	return (&KVCache{
		store: store,
	}).Reset()
}

func (kvc *KVCache) Reset() *KVCache {
	kvc.cache = make(map[string]kvCacheValue)
	kvc.keys = list.New()
	return kvc
}

func (kvc *KVCache) Set(key []byte, value []byte) {
	cacheValue, ok := kvc.cache[string(key)]
	if ok {
		kvc.keys.MoveToBack(cacheValue.e)
	} else {
		cacheValue.e = kvc.keys.PushBack(key)
	}
	cacheValue.v = value
	kvc.cache[string(key)] = cacheValue
}

func (kvc *KVCache) Get(key []byte) (value []byte) {
	cacheValue, ok := kvc.cache[string(key)]
	if ok {
		return cacheValue.v
	} else {
		value := kvc.store.Get(key)
		kvc.cache[string(key)] = kvCacheValue{
			v: value,
			e: kvc.keys.PushBack(key),
		}
		return value
	}
}

//Update the store with the values from the cache
func (kvc *KVCache) Sync() {
	for e := kvc.keys.Front(); e != nil; e = e.Next() {
		key := e.Value.([]byte)
		value := kvc.cache[string(key)]
		kvc.store.Set(key, value.v)
	}
	kvc.Reset()
}
