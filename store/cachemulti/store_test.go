package cachemulti

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/types"
)

func keyFmt(i int) []byte { return []byte((fmt.Sprintf("key%0.8d", i))) }
func valFmt(i int) []byte { return []byte((fmt.Sprintf("value%0.8d", i))) }

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

func TestStoreCopy(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	cacheKv1 := cachekv.NewStore(mem)
	cacheKv2 := cachekv.NewStore(mem)
	storeKey1 := types.NewKVStoreKey("foo")
	storeKey2 := types.NewKVStoreKey("bar")
	s1 := NewFromKVStore(
		mem,
		map[types.StoreKey]types.CacheWrapper{
			storeKey1: cacheKv1,
			storeKey2: cacheKv2,
		},
		map[string]types.StoreKey{
			"foo": storeKey1,
			"bar": storeKey2,
		},
		nil,
		nil,
	)

	kv1 := s1.GetKVStore(storeKey1)
	kv2 := s1.GetKVStore(storeKey2)

	// Set some values first
	for i := range 10 {
		kv1.Set(keyFmt(i), valFmt(i))
		kv2.Set(keyFmt(i), valFmt(i))
	}

	// Copy the multi store
	s2 := s1.Copy()
	copyKv1 := s2.GetKVStore(storeKey1)
	copyKv2 := s2.GetKVStore(storeKey2)

	// Check if the underlying kv stores are equal
	for i := range 10 {
		require.Equal(t, kv1.Get(keyFmt(i)), copyKv1.Get(keyFmt(i)))
		require.Equal(t, kv2.Get(keyFmt(i)), copyKv2.Get(keyFmt(i)))
		// Then change some values and check if the underlying kv stores reflect the changes
		// Alternate deletes on the originals and the copies
		if i%2 == 0 {
			kv1.Delete(keyFmt(i))
			kv2.Delete(keyFmt(i))
			require.Empty(t, kv1.Get(keyFmt(i)))
			require.NotEmpty(t, copyKv1.Get(keyFmt(i)))
			require.Empty(t, kv2.Get(keyFmt(i)))
			require.NotEmpty(t, copyKv2.Get(keyFmt(i)))
		} else {
			copyKv1.Delete(keyFmt(i))
			copyKv2.Delete(keyFmt(i))
			require.Empty(t, copyKv1.Get(keyFmt(i)))
			require.NotEmpty(t, kv1.Get(keyFmt(i)))
			require.Empty(t, copyKv2.Get(keyFmt(i)))
			require.NotEmpty(t, kv2.Get(keyFmt(i)))
		}

	}
}
