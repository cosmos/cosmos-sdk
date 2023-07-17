package v5_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

	var params v1.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, "", params.ExpeditedThreshold)
	require.Equal(t, (*time.Duration)(nil), params.ExpeditedVotingPeriod)

	// Run migrations.
	storeService := runtime.NewKVStoreService(govKey)
	err := v5.MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, v1.DefaultParams().ExpeditedMinDeposit, params.ExpeditedMinDeposit)
	require.Equal(t, v1.DefaultParams().ExpeditedThreshold, params.ExpeditedThreshold)
	require.Equal(t, v1.DefaultParams().ExpeditedVotingPeriod, params.ExpeditedVotingPeriod)
}
