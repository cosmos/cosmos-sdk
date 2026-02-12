package blockstm

import (
	"testing"

	"github.com/test-go/testify/require"

	"cosmossdk.io/store/cachekv"
	storetypes "cosmossdk.io/store/types"
)

type (
	foo struct {
		a bool
		b bool
	}
	bar struct {
		one int
		two int
	}
)

func TestObjMemDB(t *testing.T) {
	t.Parallel()
	obj1 := foo{true, true}
	obj2 := bar{1, 2}
	storeKey := storetypes.NewObjectStoreKey("foobar")

	// attach to a new multistore
	mmdb := NewMultiMemDB(map[storetypes.StoreKey]int{storeKey: 0})

	// get the memdb
	storage := mmdb.GetObjKVStore(storeKey)

	require.Equal(t, storetypes.StoreTypeIAVL, storage.GetStoreType())

	// initial value
	storage.Set([]byte("foo"), obj1)
	storage.Set([]byte("bar"), obj2)

	require.True(t, storage.Has([]byte("foo")))
	require.True(t, storage.Has([]byte("bar")))
	require.False(t, storage.Has([]byte("baz")))
	require.Equal(t, storage.Get([]byte("foo")), obj1)
	require.Equal(t, storage.Get([]byte("bar")), obj2)
}

func TestCacheWraps(t *testing.T) {
	t.Parallel()
	storeKey := storetypes.NewObjectStoreKey("foobar")

	// attach to a new multistore
	mmdb := NewMultiMemDB(map[storetypes.StoreKey]int{storeKey: 0})

	// get the memdb
	storage := mmdb.GetObjKVStore(storeKey)
	// attempt to cachewrap
	cacheWrapper := storage.CacheWrap()
	require.IsType(t, &cachekv.GStore[any]{}, cacheWrapper)

	cacheWrappedWithTrace := storage.CacheWrapWithTrace(nil, nil)
	require.IsType(t, &cachekv.GStore[any]{}, cacheWrappedWithTrace)
}
