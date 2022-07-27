package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3 "github.com/cosmos/cosmos-sdk/x/gov/legacy/v3"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func TestGovStoreMigrationToV3ConsensusVersion(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	govKey := sdk.NewKVStoreKey("gov")
	transientTestKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(govKey, transientTestKey)
	paramstore := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, govKey, transientTestKey, "gov")

	paramstore = paramstore.WithKeyTable(types.ParamKeyTable())

	// We assume that all deposit params are set besdides the MinInitialDepositRatio
	originalDepositParams := types.DefaultDepositParams()
	originalDepositParams.MinInitialDepositRatio = sdk.ZeroDec()
	paramstore.Set(ctx, types.ParamStoreKeyDepositParams, originalDepositParams)

	// Run migrations.
	err := v3.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new param is set.
	var depositParams types.DepositParams
	paramstore.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	require.Equal(t, v3.MinInitialDepositRatio, depositParams.MinInitialDepositRatio)
}
