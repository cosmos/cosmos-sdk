package internal

import (
	"io"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal/cachekv"
)

type KVStoreWrapper struct {
	TreeReader
}

func (r KVStoreWrapper) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeIAVL
}

func (r KVStoreWrapper) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewCacheKVStore(r)
}

func (r KVStoreWrapper) CacheWrapWithTrace(io.Writer, storetypes.TraceContext) storetypes.CacheWrap {
	logger.Warn("CacheWrapWithTrace called on KVStoreWrapper: tracing not implemented")
	return cachekv.NewCacheKVStore(r)
}

func (r KVStoreWrapper) Get(key []byte) []byte {
	v, err := r.TreeReader.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (r KVStoreWrapper) Has(key []byte) bool {
	found, err := r.TreeReader.Has(key)
	if err != nil {
		panic(err)
	}
	return found
}

func (r KVStoreWrapper) Set(key []byte, value []byte) {
	panic("readonly store: cannot set value")
}

func (r KVStoreWrapper) Delete(key []byte) {
	panic("readonly store: cannot delete value")
}

var _ storetypes.KVStore = (*KVStoreWrapper)(nil)
