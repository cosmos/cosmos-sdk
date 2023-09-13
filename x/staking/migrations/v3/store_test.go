package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	v3 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v3"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	stakingKey := storetypes.NewKVStoreKey("staking")
	tStakingKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(stakingKey, tStakingKey)
	paramstore := paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, stakingKey, tStakingKey, "staking")
	store := ctx.KVStore(stakingKey)

	// Check no params
	require.False(t, paramstore.Has(ctx, types.KeyMinCommissionRate))

	// Run migrations.
	err := v3.MigrateStore(ctx, store, encCfg.Codec, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyMinCommissionRate))
}
