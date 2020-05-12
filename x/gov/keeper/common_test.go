package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	TestProposal = types.NewTextProposal("Test", "description")
)

func createValidators(ctx sdk.Context, app *simapp.SimApp, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 5, sdk.NewInt(30000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(5)

	appCodec, _ := simapp.MakeCodecs()
	app.StakingKeeper = staking.NewKeeper(
		appCodec,
		app.GetKey(staking.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(staking.ModuleName),
	)

	val1 := staking.NewValidator(valAddrs[0], pks[0], staking.Description{})
	val2 := staking.NewValidator(valAddrs[1], pks[1], staking.Description{})
	val3 := staking.NewValidator(valAddrs[2], pks[2], staking.Description{})

	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetValidator(ctx, val3)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val3)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val2)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val3)

	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers[0]), sdk.Unbonded, val1, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[1], sdk.TokensFromConsensusPower(powers[1]), sdk.Unbonded, val2, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[2], sdk.TokensFromConsensusPower(powers[2]), sdk.Unbonded, val3, true)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return addrs, valAddrs
}
