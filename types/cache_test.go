package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tm-db"

	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestGetOrSetStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	mngr := types.NewStoreCacheManager()

	sKey := types.NewKVStoreKey("test")
	store := iavlstore.UnsafeNewStore(iavl.NewMutableTree(db, 100), 10, 10)
	store2 := mngr.GetOrSetKVStoreCache(sKey, store)

	require.NotNil(t, store2)
	require.Equal(t, store2, mngr.GetOrSetKVStoreCache(sKey, store))
}

func TestStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	mngr := types.NewStoreCacheManager()

	sKey := types.NewKVStoreKey("test")
	store := iavlstore.UnsafeNewStore(iavl.NewMutableTree(db, 100), 10, 10)
	kvStore := mngr.GetOrSetKVStoreCache(sKey, store)

	for i := 0; i < types.PersistentStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		kvStore.Set(key, value)

		res := kvStore.Get(key)
		require.Equal(t, res, value)

		res = store.Get(key)
		require.Equal(t, res, value)
	}

	for i := 0; i < types.PersistentStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))

		kvStore.Delete(key)

		res := kvStore.Get(key)
		require.Nil(t, res)

		res = store.Get(key)
		require.Nil(t, res)
	}
}
