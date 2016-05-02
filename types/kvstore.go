package types

import (
	"container/list"
	"fmt"

	. "github.com/tendermint/go-common"
)

type KVStore interface {
	Set(key, value []byte)
	Get(key []byte) (value []byte)
}

//----------------------------------------

type MemKVStore struct {
	m map[string][]byte
}

func NewMemKVStore() *MemKVStore {
	return &MemKVStore{
		m: make(map[string][]byte, 0),
	}
}

func (mkv *MemKVStore) Set(key []byte, value []byte) {
	mkv.m[string(key)] = value
}

func (mkv *MemKVStore) Get(key []byte) (value []byte) {
	return mkv.m[string(key)]
}

//----------------------------------------

// A Cache that enforces deterministic sync order.
type KVCache struct {
	store KVStore
	cache map[string]kvCacheValue
	keys  *list.List
}

type kvCacheValue struct {
	v []byte        // The value of some key
	e *list.Element // The KVCache.keys element
}

func NewKVCache(store KVStore) *KVCache {
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
	fmt.Println("Set [KVCache]", formatBytes(key), "=", formatBytes(value))
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
		fmt.Println("GET [KVCache, hit]", formatBytes(key), "=", formatBytes(cacheValue.v))
		return cacheValue.v
	} else {
		value := kvc.store.Get(key)
		kvc.cache[string(key)] = kvCacheValue{
			v: value,
			e: kvc.keys.PushBack(key),
		}
		fmt.Println("GET [KVCache, miss]", formatBytes(key), "=", formatBytes(value))
		return value
	}
}

func (kvc *KVCache) Sync() {
	for e := kvc.keys.Front(); e != nil; e = e.Next() {
		key := e.Value.([]byte)
		value := kvc.cache[string(key)]
		kvc.store.Set(key, value.v)
	}
	kvc.Reset()
}

//----------------------------------------

func formatBytes(data []byte) string {
	s := ""
	for _, b := range data {
		if 0x21 <= b && b < 0x7F {
			s += Green(string(b))
		} else {
			s += Blue(Fmt("%X", b))
		}
	}
	return s
}
