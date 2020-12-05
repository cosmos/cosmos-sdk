package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	TestProposal = types.NewTextProposal("Test", "description")
)

func createValidators(t *testing.T, ctx sdk.Context, app *simapp.SimApp, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 5, sdk.NewInt(30000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(5)

	appCodec, _ := simapp.MakeCodecs()
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		app.GetKey(stakingtypes.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)

	val1, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
	require.NoError(t, err)
	val2, err := stakingtypes.NewValidator(valAddrs[1], pks[1], stakingtypes.Description{})
	require.NoError(t, err)
	val3, err := stakingtypes.NewValidator(valAddrs[2], pks[2], stakingtypes.Description{})
	require.NoError(t, err)

	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetValidator(ctx, val3)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val3)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val2)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val3)

	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers[0]), stakingtypes.Unbonded, val1, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[1], sdk.TokensFromConsensusPower(powers[1]), stakingtypes.Unbonded, val2, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[2], sdk.TokensFromConsensusPower(powers[2]), stakingtypes.Unbonded, val3, true)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return addrs, valAddrs
}
