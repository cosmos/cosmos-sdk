package blockstm

import (
	"github.com/cosmos/cosmos-sdk/store/v2/cachemulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

var (
	_ storetypes.MultiStore = msWrapper{}
	_ MultiStore            = stmMultiStoreWrapper{}
)

type msWrapper struct {
	MultiStore
}

func (ms msWrapper) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	// TODO implement me
	panic("implement me")
}

func (ms msWrapper) LatestVersion() int64 {
	// TODO implement me
	panic("implement me")
}

func (ms msWrapper) getCacheWrapper(key storetypes.StoreKey) storetypes.CacheWrapper {
	return ms.GetStore(key)
}

func (ms msWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	return ms.MultiStore.GetStore(key)
}

func (ms msWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}

func (ms msWrapper) GetObjKVStore(key storetypes.StoreKey) storetypes.ObjKVStore {
	return ms.MultiStore.GetObjKVStore(key)
}

func (ms msWrapper) CacheMultiStore() storetypes.CacheMultiStore {
	return cachemulti.NewFromParent(ms.getCacheWrapper)
}

// CacheWrap Implements CacheWrapper.
func (ms msWrapper) CacheWrap() storetypes.CacheWrap {
	return ms.CacheMultiStore().(storetypes.CacheWrap)
}

// GetStoreType returns the type of the store.
func (ms msWrapper) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

type stmMultiStoreWrapper struct {
	storetypes.MultiStore
}

func (ms stmMultiStoreWrapper) GetStore(key storetypes.StoreKey) storetypes.Store {
	return ms.MultiStore.GetStore(key)
}

func (ms stmMultiStoreWrapper) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return ms.MultiStore.GetKVStore(key)
}
