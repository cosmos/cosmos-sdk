package v3_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3 "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v3"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGovStoreMigrationToV4ConsensusVersion(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	govKey := sdk.NewKVStoreKey("gov")
	transientTestKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(govKey, transientTestKey)
	paramstore := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, govKey, transientTestKey, "gov")

	paramstore = paramstore.WithKeyTable(types.ParamKeyTable())

	paramstore.Set(ctx, types.ParamMinimumRestakeThreshold, sdk.NewDec(0))
	paramstore.Set(ctx, types.ParamRestakePeriod, sdk.NewInt(0))

	// Run migrations.
	err := v3.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	var minimumRestake sdk.Dec
	paramstore.Get(ctx, types.ParamMinimumRestakeThreshold, &minimumRestake)
	require.Equal(t, v3.MinimumRestakeThreshold, minimumRestake)

	var restakePeriod sdk.Int
	paramstore.Get(ctx, types.ParamRestakePeriod, &restakePeriod)
	require.Equal(t, v3.RestakePeriod, restakePeriod)
}
