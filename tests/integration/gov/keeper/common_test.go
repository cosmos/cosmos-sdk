package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_, _, addr   = testdata.KeyTestPubAddr()
	govAcct      = authtypes.NewModuleAddress(types.ModuleName)
	TestProposal = getTestProposal()
)

func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))),
		legacyProposalMsg,
	}
}

func createValidators(t *testing.T, f *newFixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.ctx, 5, sdk.NewInt(30000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	val1, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
	assert.NilError(t, err)
	val2, err := stakingtypes.NewValidator(valAddrs[1], pks[1], stakingtypes.Description{})
	assert.NilError(t, err)
	val3, err := stakingtypes.NewValidator(valAddrs[2], pks[2], stakingtypes.Description{})
	assert.NilError(t, err)

	f.stakingKeeper.SetValidator(f.ctx, val1)
	f.stakingKeeper.SetValidator(f.ctx, val2)
	f.stakingKeeper.SetValidator(f.ctx, val3)
	f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val1)
	f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val2)
	f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val3)
	f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val1)
	f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val2)
	f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val3)

	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[0]), stakingtypes.Unbonded, val1, true)
	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[1]), stakingtypes.Unbonded, val2, true)
	_, _ = f.stakingKeeper.Delegate(f.ctx, addrs[2], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[2]), stakingtypes.Unbonded, val3, true)

	f.stakingKeeper.EndBlocker(f.ctx)

	return addrs, valAddrs
}

// func createValidators2(t *testing.T, ctx sdk.Context, app *simapp.SimApp, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
// 	addrs := simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 5, sdk.NewInt(30000000))
// 	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
// 	pks := simtestutil.CreateTestPubKeys(5)
// 	cdc := moduletestutil.MakeTestEncodingConfig().Codec

// 	app.StakingKeeper = stakingkeeper.NewKeeper(
// 		cdc,
// 		app.GetKey(stakingtypes.StoreKey),
// 		app.AccountKeeper,
// 		app.BankKeeper,
// 		authtypes.NewModuleAddress(types.ModuleName).String(),
// 	)

// 	val1, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
// 	assert.NilError(t, err)
// 	val2, err := stakingtypes.NewValidator(valAddrs[1], pks[1], stakingtypes.Description{})
// 	assert.NilError(t, err)
// 	val3, err := stakingtypes.NewValidator(valAddrs[2], pks[2], stakingtypes.Description{})
// 	assert.NilError(t, err)

// 	app.StakingKeeper.SetValidator(ctx, val1)
// 	app.StakingKeeper.SetValidator(ctx, val2)
// 	app.StakingKeeper.SetValidator(ctx, val3)
// 	app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
// 	app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)
// 	app.StakingKeeper.SetValidatorByConsAddr(ctx, val3)
// 	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)
// 	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val2)
// 	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val3)

// 	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], app.StakingKeeper.TokensFromConsensusPower(ctx, powers[0]), stakingtypes.Unbonded, val1, true)
// 	_, _ = app.StakingKeeper.Delegate(ctx, addrs[1], app.StakingKeeper.TokensFromConsensusPower(ctx, powers[1]), stakingtypes.Unbonded, val2, true)
// 	_, _ = app.StakingKeeper.Delegate(ctx, addrs[2], app.StakingKeeper.TokensFromConsensusPower(ctx, powers[2]), stakingtypes.Unbonded, val3, true)

// 	app.StakingKeeper.EndBlocker(ctx)

// 	return addrs, valAddrs
// }
