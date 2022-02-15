package v046_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v046staking "github.com/cosmos/cosmos-sdk/x/params/migrations/v046"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	paramkey := sdk.NewKVStoreKey("params")
	tparamkey := sdk.NewTransientStoreKey("params_transient")
	ctx := testutil.DefaultContext(paramkey, tparamkey)
	paramstore := paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, paramkey, tparamkey, baseapp.Paramspace)

	// Check no params
	require.False(t, paramstore.Has(ctx, baseapp.ParamStoreKeyVersionParams))

	// Run migrations.
	err := v046staking.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, baseapp.ParamStoreKeyVersionParams))
}
