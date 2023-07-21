package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCalculateRewardsBasic(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	distrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	// end block to bond validator and start new block
	staking.EndBlocker(ctx, stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	tstaking.Ctx = ctx

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// historical count should be 2 (once for validator init, once for delegation init)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// historical count should be 2 still
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestWithdrawTokenizeShareRecordReward(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	tstaking := stakingtestutil.NewHelper(t, ctx, app.StakingKeeper)

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

	// retrieve validator
	val = app.StakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}}
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial)}
	err := app.MintKeeper.MintCoins(ctx, coins)
	require.NoError(t, err)

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, stakingtypes.ModuleName, coins)
	require.NoError(t, err)

	// tokenize share amount
	delTokens := sdk.NewInt(1000000)
	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	resp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &stakingtypes.MsgTokenizeShares{
		DelegatorAddress:    sdk.AccAddress(valAddrs[0]).String(),
		ValidatorAddress:    valAddrs[0].String(),
		TokenizedShareOwner: sdk.AccAddress(valAddrs[1]).String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, delTokens),
	})
	require.NoError(t, err)

	// try withdrawing rewards before no reward is allocated
	coins, err = app.DistrKeeper.WithdrawAllTokenizeShareRecordReward(ctx, sdk.AccAddress(valAddrs[1]))
	require.Nil(t, err)
	require.Equal(t, coins, sdk.Coins{})

	// assert tokenize share response
	require.NoError(t, err)
	require.Equal(t, resp.Amount.Amount, delTokens)

	// end block to bond validator
	staking.EndBlocker(ctx, app.StakingKeeper)
	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// allocate some rewards
	app.DistrKeeper.AllocateTokensToValidator(ctx, val, tokens)
	// end period
	app.DistrKeeper.IncrementValidatorPeriod(ctx, val)

	beforeBalance := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)

	// withdraw rewards
	coins, err = app.DistrKeeper.WithdrawAllTokenizeShareRecordReward(ctx, sdk.AccAddress(valAddrs[1]))
	require.Nil(t, err)

	// check return value
	require.Equal(t, coins.String(), "5000stake")
	// check balance changes
	midBalance := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)
	require.Equal(t, beforeBalance.Amount.Add(coins.AmountOf(sdk.DefaultBondDenom)), midBalance.Amount)

	// allocate more rewards manually on module account and try full redeem
	record, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)

	err = app.MintKeeper.MintCoins(ctx, coins)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, record.GetModuleAddress(), coins)
	require.NoError(t, err)

	shareTokenBalance := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[0]), record.GetShareTokenDenom())

	_, err = msgServer.RedeemTokensForShares(sdk.WrapSDKContext(ctx), &stakingtypes.MsgRedeemTokensForShares{
		DelegatorAddress: sdk.AccAddress(valAddrs[0]).String(),
		Amount:           shareTokenBalance,
	})
	require.NoError(t, err)

	finalBalance := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(valAddrs[1]), sdk.DefaultBondDenom)
	require.Equal(t, midBalance.Amount.Add(coins.AmountOf(sdk.DefaultBondDenom)), finalBalance.Amount)
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	valPower := int64(100)
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk0, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// retrieve validator
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial.QuoRaw(2))}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial.QuoRaw(2))}},
		distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	valPower := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk0, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50%
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50% again
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}},
		distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(100000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del1 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	tstaking.Ctx = ctx
	tstaking.Delegate(sdk.AccAddress(valAddrs[1]), valAddrs[0], sdk.NewInt(100))
	del2 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// fetch updated validator
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	distrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	balancePower := int64(1000)
	balanceTokens := stakingKeeper.TokensFromConsensusPower(ctx, balancePower)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// set module account coins
	distrAcc := distrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balanceTokens))))
	accountKeeper.SetModuleAccount(ctx, distrAcc)

	// create validator with 50% commission
	power := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	valTokens := tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk0, power, true)

	// assert correct initial balance
	expTokens := balanceTokens.Sub(valTokens)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, expTokens)},
		bankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])

	// allocate some rewards
	initial := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (initial + latest for delegation)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw rewards
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.Nil(t, err)

	// historical count should still be 2 (added one record, cleared one)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// assert correct balance
	exp := balanceTokens.Sub(valTokens).Add(initial.QuoRaw(2))
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		bankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)

	// withdraw commission
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.Nil(t, err)

	// assert correct balance
	exp = balanceTokens.Sub(valTokens).Add(initial)
	require.Equal(t,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, exp)},
		bankKeeper.GetAllBalances(ctx, sdk.AccAddress(valAddrs[0])),
	)
}

func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// create validator with 50% commission
	valPower := int64(100)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk0, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.NewDecFromInt(stakingKeeper.TokensFromConsensusPower(ctx, 10))
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50%
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))

	// slash the validator by 50% again
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower/2, sdk.NewDecWithPrec(5, 1))

	// fetch the validator again
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, rewards)

	// commission should be the other half
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	valPower := int64(100)
	tstaking.CreateValidatorWithValPower(valAddrs[0], valConsPk0, valPower, true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del1 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	initial := sdk.NewDecFromInt(stakingKeeper.TokensFromConsensusPower(ctx, 30))
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// second delegation
	tstaking.DelegateWithPower(sdk.AccAddress(valAddrs[1]), valAddrs[0], 100)

	del2 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower, sdk.NewDecWithPrec(5, 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// fetch updated validator
	val = stakingKeeper.Validator(ctx, valAddrs[0])

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(2).Add(initial.QuoInt64(6))}}, rewards)

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
}

func TestCalculateRewardsMultiDelegatorMultWithdraw(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	distrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	initial := int64(20)

	// set module account coins
	distrAcc := distrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)))))
	accountKeeper.SetModuleAccount(ctx, distrAcc)

	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(initial))}

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator and delegation
	val := stakingKeeper.Validator(ctx, valAddrs[0])
	del1 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// allocate some rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	tstaking.Delegate(sdk.AccAddress(valAddrs[1]), valAddrs[0], sdk.NewInt(100))

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// fetch updated validator
	val = stakingKeeper.Validator(ctx, valAddrs[0])
	del2 := stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	// second delegator withdraws
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[1]), valAddrs[0])
	require.NoError(t, err)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// first delegator withdraws again
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 4)}}, rewards)

	// commission should be half initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// withdraw commission
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddrs[0])
	require.NoError(t, err)

	// end period
	endingPeriod = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, rewards)

	// commission should be zero
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
}

func Test100PercentCommissionReward(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)
	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	initial := int64(20)

	// set module account coins
	distrAcc := distrKeeper.GetDistributionAccount(ctx)
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)))))
	accountKeeper.SetModuleAccount(ctx, distrAcc)

	tokens := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(initial))}

	// create validator with 100% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(10, 1), sdk.NewDecWithPrec(10, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)
	stakingKeeper.Delegation(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])

	// end block to bond validator
	staking.EndBlocker(ctx, stakingKeeper)
	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// fetch validator
	val := stakingKeeper.Validator(ctx, valAddrs[0])

	// allocate some rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	rewards, err := distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddrs[0]), valAddrs[0])
	require.NoError(t, err)

	zeroRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())}
	require.True(t, rewards.IsEqual(zeroRewards))

	events := ctx.EventManager().Events()
	lastEvent := events[len(events)-1]

	var hasValue bool
	for _, attr := range lastEvent.Attributes {
		if attr.Key == "amount" && attr.Value == "0stake" {
			hasValue = true
		}
	}
	require.True(t, hasValue)
}
