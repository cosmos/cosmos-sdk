package v4_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/legacy/v4"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func TestGovStoreMigrationToV4ConsensusVersion(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	govKey := sdk.NewKVStoreKey("gov")
	transientTestKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(govKey, transientTestKey)
	paramstore := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, govKey, transientTestKey, "gov")

	paramstore = paramstore.WithKeyTable(types.ParamKeyTable())

	// We assume that all deposit params are set besides the MinInitialDepositRatio
	originalDepositParams := types.DefaultDepositParams()
	originalDepositParams.MinExpeditedDeposit = sdk.NewCoins()
	paramstore.Set(ctx, types.ParamStoreKeyDepositParams, originalDepositParams)

	originalVotingParams := types.DefaultVotingParams()
	originalVotingParams.ExpeditedVotingPeriod = time.Duration(0)
	paramstore.Set(ctx, types.ParamStoreKeyVotingParams, originalVotingParams)

	originalTallyParams := types.DefaultTallyParams()
	originalTallyParams.ExpeditedThreshold = sdk.ZeroDec()
	paramstore.Set(ctx, types.ParamStoreKeyTallyParams, originalTallyParams)

	// Run migrations.
	err := v4.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	var depositParams types.DepositParams
	paramstore.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	require.Equal(t, v4.MinExpeditedDeposit, depositParams.MinExpeditedDeposit)

	var votingParams types.VotingParams
	paramstore.Get(ctx, types.ParamStoreKeyVotingParams, &votingParams)
	require.Equal(t, v4.ExpeditedVotingPeriod, votingParams.ExpeditedVotingPeriod)

	var tallyParams types.TallyParams
	paramstore.Get(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
	require.Equal(t, v4.ExpeditedThreshold, tallyParams.ExpeditedThreshold)
}
