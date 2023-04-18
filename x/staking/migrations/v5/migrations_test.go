package v5_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	v2 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
	v5 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v5"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stackingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
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
	require.NoErrorf(t, v5.MigrateStore(ctx, storeKey, cdc), "v5.MigrateStore failed, seed: %d", seed)

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

func TestDelegationsByValidatorMigrations(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(v5.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	accAddrs := sims.CreateIncrementalAccounts(11)
	valAddrs := sims.ConvertAddrsToValAddrs(accAddrs[0:1])
	var addedDels []types.Delegation

	for i := 1; i < 11; i++ {
		del1 := types.NewDelegation(accAddrs[i], valAddrs[0], sdk.NewDec(100))
		store.Set(types.GetDelegationKey(accAddrs[i], valAddrs[0]), types.MustMarshalDelegation(cdc, del1))
		addedDels = append(addedDels, del1)
	}

	// before migration the state of delegations by val index should be empty
	dels := getValDelegations(ctx, cdc, storeKey, valAddrs[0])
	assert.Len(t, dels, 0)

	err := v5.MigrateStore(ctx, storeKey, cdc)
	assert.NoError(t, err)

	// after migration the state of delegations by val index should not be empty
	dels = getValDelegations(ctx, cdc, storeKey, valAddrs[0])
	assert.Len(t, dels, len(addedDels))
	assert.Equal(t, addedDels, dels)
}

func getValDelegations(ctx sdk.Context, cdc codec.Codec, storeKey storetypes.StoreKey, valAddr sdk.ValAddress) []types.Delegation {
	var delegations []types.Delegation

	store := ctx.KVStore(storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.GetDelegationsByValPrefixKey(valAddr))
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.Delegation
		valAddr, delAddr, err := types.ParseDelegationsByValKey(iterator.Key())
		if err != nil {
			panic(err)
		}

		bz := store.Get(types.GetDelegationKey(delAddr, valAddr))

		cdc.MustUnmarshal(bz, &delegation)

		delegations = append(delegations, delegation)
	}

	return delegations
}
