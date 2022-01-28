package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// createValidator creates a validator in the given SimApp.
func createValidator(t *testing.T, ctx sdk.Context, app *simapp.SimApp, powers int64) (sdk.ValAddress, stakingtypes.Validator) {
	valTokens := sdk.TokensFromConsensusPower(powers, sdk.DefaultPowerReduction)
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, valTokens)
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(1)
	cdc := app.AppCodec() //simapp.MakeTestEncodingConfig().Marshaler

	app.StakingKeeper = stakingkeeper.NewKeeper(
		cdc,
		app.GetKey(stakingtypes.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)

	val, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
	require.NoError(t, err)

	app.StakingKeeper.SetValidator(ctx, val)
	require.NoError(t, app.StakingKeeper.SetValidatorByConsAddr(ctx, val))
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val)

	_, err = app.StakingKeeper.Delegate(ctx, addrs[0], valTokens, stakingtypes.Unbonded, val, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return valAddrs[0], val
}
