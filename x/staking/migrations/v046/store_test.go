package v046_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/depinject"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	v046staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v046"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestStoreMigration(t *testing.T) {
	var (
		cdc         codec.Codec
		legacyAmino *codec.LegacyAmino
	)

	err := depinject.Inject(testutil.AppConfig,
		&cdc,
		&legacyAmino,
	)
	require.NoError(t, err)

	stakingKey := sdk.NewKVStoreKey("staking")
	tStakingKey := sdk.NewTransientStoreKey("transient_test")
	ctx := sdktestutil.DefaultContext(stakingKey, tStakingKey)
	paramstore := paramtypes.NewSubspace(cdc, legacyAmino, stakingKey, tStakingKey, "staking")

	// Check no params
	require.False(t, paramstore.Has(ctx, types.KeyMinCommissionRate))

	// Run migrations.
	err = v046staking.MigrateStore(ctx, stakingKey, cdc, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyMinCommissionRate))
}
