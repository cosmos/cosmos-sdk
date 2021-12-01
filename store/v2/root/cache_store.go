package root

import (
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
)

// GetKVStore implements BasicRootStore.
func (cs *cacheStore) GetKVStore(key types.StoreKey) types.KVStore {
	ret, has := cs.substores[key.Name()]
	if has {
		return ret
	}
	ret = cachekv.NewStore(cs.source.GetKVStore(key))
	cs.substores[key.Name()] = ret
	return ret
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
		source:        cs,
		substores:     map[string]types.CacheKVStore{},
		listenerMixin: &listenerMixin{},
		traceMixin:    &traceMixin{},
	}
}
