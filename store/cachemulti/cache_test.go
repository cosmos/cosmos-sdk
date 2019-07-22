package cachemulti_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cachemulti"
	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestGetOrSetStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, 100)
	store := iavlstore.UnsafeNewStore(tree, 10, 10)
	cwStore := store.CacheWrap()

	mngr := cachemulti.NewStoreCacheManager()
	sKey := types.NewKVStoreKey("test")
	cwStore2 := mngr.GetOrSetStoreCache(sKey, cwStore)

	require.NotNil(t, cwStore2)
	require.Equal(t, cwStore2, mngr.GetOrSetStoreCache(sKey, cwStore))
}

func TestStoreCache(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, 100)
	store := iavlstore.UnsafeNewStore(tree, 10, 10)
	cwStore := store.CacheWrap()

	mngr := cachemulti.NewStoreCacheManager()
	sKey := types.NewKVStoreKey("test")
	kvStore := mngr.GetOrSetStoreCache(sKey, cwStore).(types.KVStore)

	for i := 0; i < cachemulti.PersistentStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte(fmt.Sprintf("value_%d", i))

		kvStore.Set(key, value)

		res := kvStore.Get(key)
		require.Equal(t, res, value)

		res = cwStore.(types.KVStore).Get(key)
		require.Equal(t, res, value)
	}

	for i := 0; i < cachemulti.PersistentStoreCacheSize*2; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))

		kvStore.Delete(key)

		res := kvStore.Get(key)
		require.Nil(t, res)

		res = cwStore.(types.KVStore).Get(key)
		require.Nil(t, res)
	}
}
