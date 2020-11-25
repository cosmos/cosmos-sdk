package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCalculateRewardsBasic(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, sdk.NewInt(100), true)

	// end block to bond validator and start new block
	staking.EndBlocker(ctx, app.StakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	tstaking.Ctx = ctx

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// historical count should be 2 (once for validator init, once for delegation init)
	require.Equal(t, uint64(2), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// historical count should be 2 still
	require.Equal(t, uint64(2), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// calculate delegation rewards
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valPower := int64(100)
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// retrieve validator
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoRaw(2).ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoRaw(2).ToDec()}},
		app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	valPower := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50% again
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.ToDec()}},
		app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, sdk.NewInt(100), true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del1 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	tstaking.Ctx = ctx
	tstaking.Delegate(sdk.AccAddress(valAddrs[1]), valAddrs[0], sdk.NewInt(100))
	del2 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// fetch updated validator
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	balancePower := int64(1000)
	balanceTokens := sdk.TokensFromConsensusPower(balancePower)
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens))))
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
	initial := sdk.TokensFromConsensusPower(10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (initial + latest for delegation)
	require.Equal(t, uint64(2), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw rewards
	_, err := app.DistrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	// historical count should still be 2 (added one record, cleared one)
	require.Equal(t, uint64(2), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// assert correct balance
	exp := balanceTokens.Sub(valTokens).Add(initial.QuoRaw(2))
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// withdraw commission
	_, err = app.DistrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.Nil(t, err)

	// assert correct balance
	exp = balanceTokens.Sub(valTokens).Add(initial)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)
}

func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator with 50% commission
	valPower := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10).ToDec()
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50%
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// slash the validator by 50% again
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	valPower := int64(100)
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk1, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del1 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(30).ToDec()
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// second delegation
	tstaking.DelegateWithPower(sdk.AccAddress(valAddrs[1]), valAddrs[0], 100)

	del2 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	app.StakingKeeper.Slash(ctx, valConsAddr1, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// fetch updated validator
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(2).Add(initial.QuoInt64(6))}}, rewards)

	// calculate delegation rewards for del2
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)
	initial := int64(20)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	err := app.BankKeeper.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000))))
	require.NoError(t, err)
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(initial))}

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk1, sdk.NewInt(100), true)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := app.StakingKeeper.Validator(ctx, valAddrs[0])
	del1 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	tstaking.Delegate(sdk.AccAddress(valAddrs[1]), valAddrs[0], sdk.NewInt(100))

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// fetch updated validator
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])
	del2 := app.StakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws
	_, err = app.DistrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	// second delegator withdraws
	_, err = app.DistrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])
	require.NoError(t, err)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), app.DistrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission
	_, err = app.DistrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod := app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := app.DistrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws again
	_, err = app.DistrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw commission
	_, err = app.DistrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod = app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = app.DistrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, app.DistrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
}
