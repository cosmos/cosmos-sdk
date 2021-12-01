package root

import (
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
)

// GetKVStore implements BasicRootStore.
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

// Write implements CacheRootStore.
func (cs *cacheStore) Write() {
	for skey, sub := range cs.substores {
		sub.Write()
		delete(cs.substores, skey)
	}
}

// CacheRootStore implements BasicRootStore.
// This recursively wraps the CacheRootStore in another cache store.
func (cs *cacheStore) CacheRootStore() types.CacheRootStore {
	return &cacheStore{
		source:           cs,
		substores:        map[string]types.CacheKVStore{},
		traceListenMixin: newTraceListenMixin(),
	}
}
