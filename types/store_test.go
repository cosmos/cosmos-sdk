package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPrefixEndBytes(t *testing.T) {
	var testCases = []struct {
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
		end := sdk.PrefixEndBytes(test.prefix)
		require.Equal(t, test.expected, end)
	}
}

func TestCommitID(t *testing.T) {
	var empty sdk.CommitID
	require.True(t, empty.IsZero())

	var nonempty = sdk.CommitID{
		Version: 1,
		Hash:    []byte("testhash"),
	}
	require.False(t, nonempty.IsZero())
}

func TestNewKVStoreKeys(t *testing.T) {
	t.Parallel()
	require.Equal(t, map[string]*sdk.KVStoreKey{}, sdk.NewKVStoreKeys())
	require.Equal(t, 1, len(sdk.NewKVStoreKeys("one")))
}

func TestNewTransientStoreKeys(t *testing.T) {
	t.Parallel()
	require.Equal(t, map[string]*sdk.TransientStoreKey{}, sdk.NewTransientStoreKeys())
	require.Equal(t, 1, len(sdk.NewTransientStoreKeys("one")))
}

func TestNewInfiniteGasMeter(t *testing.T) {
	t.Parallel()
	gm := sdk.NewInfiniteGasMeter()
	require.NotNil(t, gm)
	_, ok := gm.(types.GasMeter)
	require.True(t, ok)
}

func TestStoreTypes(t *testing.T) {
	t.Parallel()
	require.Equal(t, sdk.InclusiveEndBytes([]byte("endbytes")), types.InclusiveEndBytes([]byte("endbytes")))
}

func TestDiffKVStores(t *testing.T) {
	t.Parallel()
	store1, store2 := initTestStores(t)
	// Two equal stores
	k1, v1 := []byte("k1"), []byte("v1")
	store1.Set(k1, v1)
	store2.Set(k1, v1)

	checkDiffResults(t, store1, store2)

	// delete k1 from store2, which is now empty
	store2.Delete(k1)
	checkDiffResults(t, store1, store2)

	// set k1 in store2, different value than what store1 holds for k1
	v2 := []byte("v2")
	store2.Set(k1, v2)
	checkDiffResults(t, store1, store2)

	// add k2 to store2
	k2 := []byte("k2")
	store2.Set(k2, v2)
	checkDiffResults(t, store1, store2)

	// Reset stores
	store1.Delete(k1)
	store2.Delete(k1)
	store2.Delete(k2)

	// Same keys, different value. Comparisons will be nil as prefixes are skipped.
	prefix := []byte("prefix:")
	k1Prefixed := append(prefix, k1...)
	store1.Set(k1Prefixed, v1)
	store2.Set(k1Prefixed, v2)
	checkDiffResults(t, store1, store2)
}

func initTestStores(t *testing.T) (types.KVStore, types.KVStore) {
	db := dbm.NewMemDB()
	ms := rootmulti.NewStore(db)

	key1 := types.NewKVStoreKey("store1")
	key2 := types.NewKVStoreKey("store2")
	require.NotPanics(t, func() { ms.MountStoreWithDB(key1, types.StoreTypeIAVL, db) })
	require.NotPanics(t, func() { ms.MountStoreWithDB(key2, types.StoreTypeIAVL, db) })
	require.NoError(t, ms.LoadLatestVersion())
	return ms.GetKVStore(key1), ms.GetKVStore(key2)
}

func checkDiffResults(t *testing.T, store1, store2 types.KVStore) {
	kvAs1, kvBs1 := sdk.DiffKVStores(store1, store2, nil)
	kvAs2, kvBs2 := types.DiffKVStores(store1, store2, nil)
	require.Equal(t, kvAs1, kvAs2)
	require.Equal(t, kvBs1, kvBs2)
}
