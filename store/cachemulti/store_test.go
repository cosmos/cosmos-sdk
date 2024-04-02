package cachemulti

import (
	"errors"
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/internal"
	"cosmossdk.io/store/internal/btree"
	"cosmossdk.io/store/types"
)

func TestStoreGetKVStore(t *testing.T) {
	require := require.New(t)

	s := Store{stores: map[types.StoreKey]types.CacheWrap{}}
	key := types.NewKVStoreKey("abc")
	errMsg := fmt.Sprintf("kv store with key %v has not been registered in stores", key)

	require.PanicsWithValue(errMsg,
		func() { s.GetStore(key) })

	require.PanicsWithValue(errMsg,
		func() { s.GetKVStore(key) })
}

func TestRunAtomic(t *testing.T) {
	store := dbadapter.Store{DB: dbm.NewMemDB()}
	objStore := internal.NewBTreeStore(btree.NewBTree[any](),
		func(v any) bool { return v == nil },
		func(v any) int { return 1 },
	)
	keys := map[string]types.StoreKey{
		"abc": types.NewKVStoreKey("abc"),
		"obj": types.NewObjectStoreKey("obj"),
	}
	parent := Store{stores: map[types.StoreKey]types.CacheWrap{
		keys["abc"]: store.CacheWrap(),
		keys["obj"]: objStore.CacheWrap(),
	}}

	s := Store{stores: map[types.StoreKey]types.CacheWrap{}, parentStore: parent.getCacheWrap}
	s.RunAtomic(func(ms types.CacheMultiStore) error {
		ms.GetKVStore(keys["abc"]).Set([]byte("key"), []byte("value"))
		ms.GetObjKVStore(keys["obj"]).Set([]byte("key"), "value")
		return nil
	})
	require.Equal(t, []byte("value"), s.GetKVStore(keys["abc"]).Get([]byte("key")))
	require.Equal(t, []byte(nil), s.GetKVStore(keys["abc"]).Get([]byte("key-non-exist")))
	require.Equal(t, "value", s.GetObjKVStore(keys["obj"]).Get([]byte("key")).(string))

	require.Error(t, s.RunAtomic(func(ms types.CacheMultiStore) error {
		ms.GetKVStore(keys["abc"]).Set([]byte("key"), []byte("value2"))
		ms.GetObjKVStore(keys["obj"]).Set([]byte("key"), "value2")
		return errors.New("failure")
	}))
	require.Equal(t, []byte("value"), s.GetKVStore(keys["abc"]).Get([]byte("key")))
	require.Equal(t, "value", s.GetObjKVStore(keys["obj"]).Get([]byte("key")).(string))
}

func TestBranchStore(t *testing.T) {
	store := dbadapter.Store{DB: dbm.NewMemDB()}
	objStore := internal.NewBTreeStore(btree.NewBTree[any](),
		func(v any) bool { return v == nil },
		func(v any) int { return 1 },
	)
	keys := map[string]types.StoreKey{
		"abc": types.NewKVStoreKey("abc"),
		"obj": types.NewObjectStoreKey("obj"),
	}
	parent := Store{stores: map[types.StoreKey]types.CacheWrap{
		keys["abc"]: store.CacheWrap(),
		keys["obj"]: objStore.CacheWrap(),
	}}

	s := Store{stores: map[types.StoreKey]types.CacheWrap{}, parentStore: parent.getCacheWrap}
	s.GetKVStore(keys["abc"]).Set([]byte("key"), []byte("value"))
	snapshot := s.Clone()
	s.GetKVStore(keys["abc"]).Set([]byte("key"), []byte("value2"))
	s.GetObjKVStore(keys["obj"]).Set([]byte("key"), "value")
	s.Restore(snapshot)
	require.Equal(t, []byte("value"), s.GetKVStore(keys["abc"]).Get([]byte("key")))
	require.Equal(t, nil, s.GetObjKVStore(keys["obj"]).Get([]byte("key")))
}
