package cache_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/cache"
	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestGetOrSetStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree, err := iavl.NewMutableTree(db, 100)
	require.NoError(t, err)
	store := iavlstore.UnsafeNewStore(tree)
	store2 := mngr.GetStoreCache(sKey, store)

	require.NotNil(t, store2)
	require.Equal(t, store2, mngr.GetStoreCache(sKey, store))
}

func TestUnwrap(t *testing.T) {
	db := dbm.NewMemDB()
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree, err := iavl.NewMutableTree(db, 100)
	require.NoError(t, err)
	store := iavlstore.UnsafeNewStore(tree)
	_ = mngr.GetStoreCache(sKey, store)

	require.Equal(t, store, mngr.Unwrap(sKey))
	require.Nil(t, mngr.Unwrap(types.NewKVStoreKey("test2")))
}

func TestStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	mngr := cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)

	sKey := types.NewKVStoreKey("test")
	tree, err := iavl.NewMutableTree(db, 100)
	require.NoError(t, err)
	store := iavlstore.UnsafeNewStore(tree)
	kvStore := mngr.GetStoreCache(sKey, store)

	for i := uint(0); i < cache.DefaultCommitKVStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		kvStore.Set(key, value)

		res := kvStore.Get(key)
		require.Equal(t, res, value)
		require.Equal(t, res, store.Get(key))

		kvStore.Delete(key)

		require.Nil(t, kvStore.Get(key))
		require.Nil(t, store.Get(key))
	}
}
