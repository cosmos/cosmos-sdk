package v5_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	v5 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v5"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)
	storeService := runtime.NewKVStoreService(govKey)
	sb := collections.NewSchemaBuilder(storeService)
	constitutionCollection := collections.NewItem(sb, v5.ConstitutionKey, "constitution", collections.StringValue)

	var params v1.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, "", params.ExpeditedThreshold)
	require.Equal(t, (*time.Duration)(nil), params.ExpeditedVotingPeriod)

	// Run migrations.
	err := v5.MigrateStore(ctx, storeService, cdc, constitutionCollection)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	// After migration, expect `ExpeditedMinDeposit` to equal previous value of `MinDeposit`.
	require.Equal(t, params.MinDeposit, params.ExpeditedMinDeposit)
	require.Equal(t, "0.750000000000000000", params.ExpeditedThreshold)
	require.Equal(t, "1.000000000000000000", params.ProposalCancelRatio)
	require.Equal(t, "", params.ProposalCancelDest)
	require.Equal(t, 24*time.Hour, *params.ExpeditedVotingPeriod)
	require.Equal(t, "0.010000000000000000", params.MinDepositRatio)

	// Check constitution
	result, err := constitutionCollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "This chain has no constitution.", result)
}
