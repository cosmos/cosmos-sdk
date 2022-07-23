package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"cosmossdk.io/store/types"
)

func initTestStores(t *testing.T) (types.KVStore, types.KVStore) {
	db := dbm.NewMemDB()
	ms := rootmulti.NewStore(db, log.NewNopLogger())

	key1 := types.NewKVStoreKey("store1")
	key2 := types.NewKVStoreKey("store2")
	require.NotPanics(t, func() { ms.MountStoreWithDB(key1, types.StoreTypeIAVL, db) })
	require.NotPanics(t, func() { ms.MountStoreWithDB(key2, types.StoreTypeIAVL, db) })
	require.NoError(t, ms.LoadLatestVersion())
	return ms.GetKVStore(key1), ms.GetKVStore(key2)
}

func TestDiffKVStores(t *testing.T) {
	t.Parallel()
	store1, store2 := initTestStores(t)
	// Two equal stores
	k1, v1 := []byte("k1"), []byte("v1")
	store1.Set(k1, v1)
	store2.Set(k1, v1)

	kvAs, kvBs := types.DiffKVStores(store1, store2, nil)
	require.Equal(t, 0, len(kvAs))
	require.Equal(t, len(kvAs), len(kvBs))

	// delete k1 from store2, which is now empty
	store2.Delete(k1)
	kvAs, kvBs = types.DiffKVStores(store1, store2, nil)
	require.Equal(t, 1, len(kvAs))
	require.Equal(t, len(kvAs), len(kvBs))

	// set k1 in store2, different value than what store1 holds for k1
	v2 := []byte("v2")
	store2.Set(k1, v2)
	kvAs, kvBs = types.DiffKVStores(store1, store2, nil)
	require.Equal(t, 1, len(kvAs))
	require.Equal(t, len(kvAs), len(kvBs))

	// add k2 to store2
	k2 := []byte("k2")
	store2.Set(k2, v2)
	kvAs, kvBs = types.DiffKVStores(store1, store2, nil)
	require.Equal(t, 2, len(kvAs))
	require.Equal(t, len(kvAs), len(kvBs))

	// Reset stores
	store1.Delete(k1)
	store2.Delete(k1)
	store2.Delete(k2)

	// Same keys, different value. Comparisons will be nil as prefixes are skipped.
	prefix := []byte("prefix:")
	k1Prefixed := append(prefix, k1...)
	store1.Set(k1Prefixed, v1)
	store2.Set(k1Prefixed, v2)
	kvAs, kvBs = types.DiffKVStores(store1, store2, [][]byte{prefix})
	require.Equal(t, 0, len(kvAs))
	require.Equal(t, len(kvAs), len(kvBs))
}

func TestPrefixEndBytes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		prefix   []byte
		expected []byte
	}{
		{[]byte{byte(55), byte(255), byte(255), byte(0)}, []byte{byte(55), byte(255), byte(255), byte(1)}},
		{[]byte{byte(55), byte(255), byte(255), byte(15)}, []byte{byte(55), byte(255), byte(255), byte(16)}},
		{[]byte{byte(55), byte(200), byte(255)}, []byte{byte(55), byte(201)}},
		{[]byte{byte(55), byte(255), byte(255)}, []byte{byte(56)}},
		{[]byte{byte(255), byte(255), byte(255)}, nil},
		{[]byte{byte(255)}, nil},
		{nil, nil},
	}

	for _, test := range testCases {
		end := types.PrefixEndBytes(test.prefix)
		assert.DeepEqual(t, test.expected, end)
	}
}

func TestInclusiveEndBytes(t *testing.T) {
	t.Parallel()
	assert.DeepEqual(t, []byte{0x00}, types.InclusiveEndBytes(nil))
	bs := []byte("test")
	assert.DeepEqual(t, append(bs, byte(0x00)), types.InclusiveEndBytes(bs))
}
