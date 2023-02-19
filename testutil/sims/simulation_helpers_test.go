package sims

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGetSimulationLog(t *testing.T) {
	legacyAmino := codec.NewLegacyAmino()
	decoders := make(simulation.StoreDecoderRegistry)
	decoders[authtypes.StoreKey] = func(kvAs, kvBs kv.Pair) string { return "10" }

	tests := []struct {
		store       string
		kvPairs     []kv.Pair
		expectedLog string
	}{
		{
			"Empty",
			[]kv.Pair{{}},
			"",
		},
		{
			authtypes.StoreKey,
			[]kv.Pair{{Key: authtypes.GlobalAccountNumberKey, Value: legacyAmino.MustMarshal(uint64(10))}},
			"10",
		},
		{
			"OtherStore",
			[]kv.Pair{{Key: []byte("key"), Value: []byte("value")}},
			fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", []byte("key"), []byte("value"), []byte("key"), []byte("value")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.store, func(t *testing.T) {
			require.Equal(t, tt.expectedLog, GetSimulationLog(tt.store, decoders, tt.kvPairs, tt.kvPairs), tt.store)
		})
	}
}

func TestDiffKVStores(t *testing.T) {
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
	k1Prefixed := append(prefix, k1...) //nolint:gocritic // append is fine here
	store1.Set(k1Prefixed, v1)
	store2.Set(k1Prefixed, v2)
	checkDiffResults(t, store1, store2)
}

func checkDiffResults(t *testing.T, store1, store2 storetypes.KVStore) {
	kvAs1, kvBs1 := DiffKVStores(store1, store2, nil)
	kvAs2, kvBs2 := DiffKVStores(store1, store2, nil)
	assert.DeepEqual(t, kvAs1, kvAs2)
	assert.DeepEqual(t, kvBs1, kvBs2)
}

func initTestStores(t *testing.T) (storetypes.KVStore, storetypes.KVStore) {
	db := dbm.NewMemDB()
	ms := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	key1 := storetypes.NewKVStoreKey("store1")
	key2 := storetypes.NewKVStoreKey("store2")
	require.NotPanics(t, func() { ms.MountStoreWithDB(key1, storetypes.StoreTypeIAVL, db) })
	require.NotPanics(t, func() { ms.MountStoreWithDB(key2, storetypes.StoreTypeIAVL, db) })
	require.NotPanics(t, func() { ms.LoadLatestVersion() })
	return ms.GetKVStore(key1), ms.GetKVStore(key2)
}
