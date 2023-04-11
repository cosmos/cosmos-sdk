package v5_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	v2 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
	v5 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v5"
	stackingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestHistoricalKeysMigration(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("staking")
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	type testCase struct {
		oldKey, newKey []byte
		historicalInfo []byte
	}

	testCases := make(map[int64]testCase)

	// edge cases
	testCases[0], testCases[1], testCases[math.MaxInt32] = testCase{}, testCase{}, testCase{}

	// random cases
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < 10; i++ {
		height := r.Intn(math.MaxInt32-2) + 2

		testCases[int64(height)] = testCase{}
	}

	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	for height := range testCases {
		testCases[height] = testCase{
			oldKey:         v2.GetHistoricalInfoKey(height),
			newKey:         v5.GetHistoricalInfoKey(height),
			historicalInfo: cdc.MustMarshal(createHistoricalInfo(height, "testChainID")),
		}
	}

	// populate store using old key format
	for _, tc := range testCases {
		store.Set(tc.oldKey, tc.historicalInfo)
	}

	// migrate store to new key format
	require.NoErrorf(t, v5.MigrateStore(ctx, storeKey), "v5.MigrateStore failed, seed: %d", seed)

	// check results
	for _, tc := range testCases {
		require.Nilf(t, store.Get(tc.oldKey), "old key should be deleted, seed: %d", seed)
		require.NotNilf(t, store.Get(tc.newKey), "new key should be created, seed: %d", seed)
		require.Equalf(t, tc.historicalInfo, store.Get(tc.newKey), "seed: %d", seed)
	}
}

func createHistoricalInfo(height int64, chainID string) *stackingtypes.HistoricalInfo {
	return &stackingtypes.HistoricalInfo{Header: cmtproto.Header{ChainID: chainID, Height: height}}
}
