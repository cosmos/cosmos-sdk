package types

/*
// XXX: adapters are temporal solution for managing duplicated store.go types
// will be deleted in the following prs

import (
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
)

type CacheWrapAdapter struct {
	parent types.CacheWrap
}

var _ stypes.CacheWrap = CacheWrapAdapter{}

func (c CacheWrapAdapter) Write() {
	c.parent.Write()
}

func (c CacheWrapAdapter) CacheWrap() stypes.CacheWrap {
	return CacheWrapAdapter{c.parent.CacheWrap()}
}

func (c CacheWrapAdapter) CacheWrapWithTrace(w io.Writer, tc stypes.TraceContext) stypes.CacheWrap {
	return CacheWrapAdapter{c.parent.CacheWrapWithTrace(w, types.TraceContext(tc))}
}

// XXX: temporal type to wrap types.KVStore
type KVStoreAdapter struct {
	types.KVStore
}

var _ stypes.KVStore = KVStoreAdapter{}

func (store KVStoreAdapter) CacheWrap() stypes.CacheWrap {
	return CacheWrapAdapter{store.KVStore.CacheWrap()}
}

func (store KVStoreAdapter) CacheWrapWithTrace(w io.Writer, tc stypes.TraceContext) stypes.CacheWrap {
	return CacheWrapAdapter{store.KVStore.CacheWrapWithTrace(w, types.TraceContext(tc))}
}

func (store KVStoreAdapter) GetStoreType() stypes.StoreType {
	return stypes.StoreType(store.KVStore.GetStoreType())
}*/
