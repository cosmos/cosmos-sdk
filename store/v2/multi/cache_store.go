package multi

import (
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
)

// Branched state
type cacheStore struct {
	source    types.MultiStore
	substores map[string]types.CacheKVStore
	*traceListenMixin
}

func newCacheStore(bs types.MultiStore) *cacheStore {
	return &cacheStore{
		source:           bs,
		substores:        map[string]types.CacheKVStore{},
		traceListenMixin: newTraceListenMixin(),
	}
}

// GetKVStore implements MultiStore.
func (cs *cacheStore) GetKVStore(skey types.StoreKey) types.KVStore {
	key := skey.Name()
	sub, has := cs.substores[key]
	if !has {
		sub = cachekv.NewStore(cs.source.GetKVStore(skey))
		cs.substores[key] = sub
	}
	// Wrap with trace/listen if needed. Note: we don't cache this, so users must get a new substore after
	// modifying tracers/listeners.
	return cs.wrapTraceListen(sub, skey)
}

// Write implements CacheMultiStore.
func (cs *cacheStore) Write() {
	for _, sub := range cs.substores {
		sub.Write()
	}
}

// CacheMultiStore implements MultiStore.
// This recursively wraps the CacheMultiStore in another cache store.
func (cs *cacheStore) CacheWrap() types.CacheMultiStore {
	return newCacheStore(cs)
}

// A non-writable cache for interface wiring purposes
type noopCacheStore struct {
	types.CacheMultiStore
}

func (noopCacheStore) Write() {}

// pretend commit store is cache store
func CommitAsCacheStore(s types.CommitMultiStore) types.CacheMultiStore {
	return noopCacheStore{newCacheStore(s)}
}
