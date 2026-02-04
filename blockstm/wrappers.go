package blockstm

import (
	"io"

	"cosmossdk.io/store/cachemulti"
	storetypes "cosmossdk.io/store/types"
)

var (
	_ storetypes.MultiStore = msWrapper{}
	_ MultiStore            = stmMultiStoreWrapper{}
)

type msWrapper struct {
	MultiStore
}

func (ms msWrapper) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement me
	panic("implement me")
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
	return cachemulti.NewFromParent(ms.getCacheWrapper, nil, nil)
}

// CacheWrap Implements CacheWrapper.
func (ms msWrapper) CacheWrap() storetypes.CacheWrap {
	return ms.CacheMultiStore().(storetypes.CacheWrap)
}

// GetStoreType returns the type of the store.
func (ms msWrapper) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

// SetTracer Implements interface MultiStore
func (ms msWrapper) SetTracer(io.Writer) storetypes.MultiStore {
	return nil
}

// SetTracingContext Implements interface MultiStore
func (ms msWrapper) SetTracingContext(storetypes.TraceContext) storetypes.MultiStore {
	return nil
}

// TracingEnabled Implements interface MultiStore
func (ms msWrapper) TracingEnabled() bool {
	return false
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
