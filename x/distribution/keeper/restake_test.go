package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestRestake(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	balancePower := int64(1000)
	balanceTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, balancePower)
	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valTokens := tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, power, true)

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])

	// allocate some rewards
	initial := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	err := app.DistrKeeper.SaveAutoRestakeEntry(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// calculate delegation rewards before restake
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.True(t, rewards.AmountOf(sdk.DefaultBondDenom).IsPositive())

	err = app.DistrKeeper.PerformRestake(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	// calculate delegation rewards
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.True(t, rewards.AmountOf(sdk.DefaultBondDenom).IsZero())
}

func TestRestakeFromBeginBlocker(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	balancePower := int64(1000)
	balanceTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, balancePower)
	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valTokens := tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, power, true)

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])

	// allocate some rewards
	initial := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	err := app.DistrKeeper.SaveAutoRestakeEntry(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	// make sure this doesn't run on the wrong block
	ctx = ctx.WithBlockHeight(999)
	distribution.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.DistrKeeper)

	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.True(t, rewards.AmountOf(sdk.DefaultBondDenom).IsPositive())

	// make sure this doesn't run on the wrong block
	ctx = ctx.WithBlockHeight(1000)
	distribution.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.DistrKeeper)

	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.True(t, rewards.AmountOf(sdk.DefaultBondDenom).IsZero())
}

func TestRemoveRestakeEntry(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	balancePower := int64(1000)
	balanceTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, balancePower)
	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valTokens := tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, power, true)

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])

	// allocate some rewards
	initial := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	err := app.DistrKeeper.SaveAutoRestakeEntry(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	err = app.DistrKeeper.DeleteAutoRestakeEntry(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	ctx = ctx.WithBlockHeight(1000)
	distribution.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.DistrKeeper)

	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.True(t, rewards.AmountOf(sdk.DefaultBondDenom).IsPositive())
}
