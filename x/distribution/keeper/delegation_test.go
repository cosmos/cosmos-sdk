package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCalculateRewardsBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())
	distrKeeper.SetParams(ctx, disttypes.DefaultParams())

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0)
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// set validator's tokens and delegator shares
	val.Tokens = sdk.NewInt(1000)
	val.DelegatorShares = math.LegacyNewDecFromInt(val.Tokens)
	del := stakingtypes.NewDelegation(addr, valAddr, val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(3)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

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
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr).Commission)
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())
	distrKeeper.SetParams(ctx, disttypes.DefaultParams())

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0)
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// set validator's tokens and delegator shares
	valPower := int64(100)
	val.Tokens = sdk.TokensFromConsensusPower(valPower, sdk.DefaultPowerReduction)
	val.DelegatorShares = math.LegacyNewDecFromInt(val.Tokens)
	del := stakingtypes.NewDelegation(addr, valAddr, val.DelegatorShares)

	// set mock calls
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50% (simulated with manual calls; we assume the validator is bonded)
	slashedTokens := distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower,
		sdk.NewDecWithPrec(2, 1),
		&val,
		&distrKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
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
		distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr).Commission)
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())
	distrKeeper.SetParams(ctx, disttypes.DefaultParams())

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0)
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// set validator's tokens and delegator shares
	valPower := int64(100)
	val.Tokens = sdk.TokensFromConsensusPower(valPower, sdk.DefaultPowerReduction)
	val.DelegatorShares = math.LegacyNewDecFromInt(val.Tokens)
	del := stakingtypes.NewDelegation(addr, valAddr, val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// slash the validator by 50% (simulated with manual calls; we assume the validator is bonded)
	slashedTokens := distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower,
		sdk.NewDecWithPrec(2, 1),
		&val,
		&distrKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

	// expect a call for the next slash with the updated validator
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromInt(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// slash the validator by 50% again
	slashedTokens = distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower/2,
		sdk.NewDecWithPrec(2, 1),
		&val,
		&distrKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

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
		distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr).Commission)
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())
	distrKeeper.SetParams(ctx, disttypes.DefaultParams())

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr0 := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0)
	require.NoError(t, err)

	val.Commission = stakingtypes.NewCommission(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del0 := stakingtypes.NewDelegation(addr0, valAddr, val.DelegatorShares)

	// set mock calls
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr0, valAddr).Return(del0).Times(1)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr0, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// second delegation
	addr1 := sdk.AccAddress(valConsAddr1)
	_, del1, err := distrtestutil.Delegate(ctx, distrKeeper, addr1, &val, sdk.NewInt(10000000), nil)
	require.NoError(t, err)

	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr1, valAddr).Return(del1)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val).Times(1)

	// call necessary hooks to update a delegation
	err = distrKeeper.Hooks().AfterDelegationModified(ctx, addr1, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// end period
	endingPeriod := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards := distrKeeper.CalculateDelegationRewards(ctx, val, del0, endingPeriod)

	// rewards for del0 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards = distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr).Commission)
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(distrtestutil.AppConfig,
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
	tstaking := teststaking.NewHelper(t, ctx, stakingKeeper)

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

	app, err := simtestutil.Setup(distrtestutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addr := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addr)
	tstaking := teststaking.NewHelper(t, ctx, stakingKeeper)

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

	app, err := simtestutil.Setup(distrtestutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, stakingKeeper)
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

	app, err := simtestutil.Setup(distrtestutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	distrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	tstaking := teststaking.NewHelper(t, ctx, stakingKeeper)
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

	app, err := simtestutil.Setup(distrtestutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	tstaking := teststaking.NewHelper(t, ctx, stakingKeeper)
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

	denom, _ := sdk.GetBaseDenom()
	zeroRewards := sdk.Coins{
		sdk.Coin{
			Denom:  denom,
			Amount: math.ZeroInt(),
		},
	}
	require.True(t, rewards.IsEqual(zeroRewards))
	events := ctx.EventManager().Events()
	lastEvent := events[len(events)-1]
	hasValue := false
	for _, attr := range lastEvent.Attributes {
		if string(attr.Key) == "amount" && string(attr.Value) == "0" {
			hasValue = true
		}
	}
	require.True(t, hasValue)
}
