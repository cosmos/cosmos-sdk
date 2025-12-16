package v6_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	v6 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v6"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(v6.ModuleName)
	storeService := runtime.NewKVStoreService(storeKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	var params types.Params
	bz := store.Get(v6.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)

	// Run migrations.
	err := v6.MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v6.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, types.DefaultParams().MaxCommissionRate, params.MaxCommissionRate)
}
