package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCalculateRewardsBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(1000))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// delegation mock
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(3)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// historical count should be 2 (once for validator init, once for delegation init)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// historical count should be 2 still
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// calculate delegation rewards
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// allocate some rewards
	initial := int64(10)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, rewards)

	// commission should be the other half
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, valCommission.Commission)
}

func TestCalculateRewardsAfterSlash(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	valPower := int64(100)
	stake := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	val, err := distrtestutil.CreateValidator(valConsPk0, stake)
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)

	// set mock calls
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

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
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial.QuoRaw(2))}}, rewards)

	// commission should be the other half
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial.QuoRaw(2))}},
		valCommission.Commission)
}

func TestCalculateRewardsAfterManySlashes(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	valPower := int64(100)
	stake := sdk.TokensFromConsensusPower(valPower, sdk.DefaultPowerReduction)
	val, err := distrtestutil.CreateValidator(valConsPk0, stake)
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// delegation mocks
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

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
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

	// expect a call for the next slash with the updated validator
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(1)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// slash the validator by 50% again
	slashedTokens = distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower/2,
		math.LegacyNewDecWithPrec(2, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)
	require.True(t, slashedTokens.IsPositive(), "expected positive slashed tokens, got: %s", slashedTokens)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial)}}, rewards)

	// commission should be the other half
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecFromInt(initial)}},
		valCommission.Commission)
}

func TestCalculateRewardsMultiDelegator(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr0 := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del0 := stakingtypes.NewDelegation(addr0.String(), valAddr.String(), val.DelegatorShares)

	// set mock calls
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(4)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr0, valAddr).Return(del0, nil).Times(1)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr0, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// second delegation
	addr1 := sdk.AccAddress(valConsAddr1)
	_, del1, err := distrtestutil.Delegate(ctx, distrKeeper, addr1, &val, math.NewInt(100), nil, stakingKeeper)
	require.NoError(t, err)

	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr1, valAddr).Return(del1, nil)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(1)

	// call necessary hooks to update a delegation
	err = distrKeeper.Hooks().AfterDelegationModified(ctx, addr1, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del0, endingPeriod)
	require.NoError(t, err)

	// rewards for del0 should be 3/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 3 / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del1, endingPeriod)
	require.NoError(t, err)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial * 1 / 4)}}, rewards)

	// commission should be equal to initial (50% twice)
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial)}}, valCommission.Commission)
}

func TestWithdrawDelegationRewardsBasic(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// delegation mock
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(5)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).Times(3)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}

	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// historical count should be 2 (initial + latest for delegation)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw rewards (the bank keeper should be called with the right amount of tokens to transfer)
	expRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial.QuoRaw(2))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expRewards)
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err)

	// historical count should still be 2 (added one record, cleared one)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// withdraw commission (the bank keeper should be called with the right amount of tokens to transfer)
	expCommission := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial.QuoRaw(2))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expCommission)
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddr)
	require.Nil(t, err)
}

func TestCalculateRewardsAfterManySlashesInSameBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// delegation mock
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(5)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be zero
	require.True(t, rewards.IsZero())

	// start out block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some rewards
	initial := math.LegacyNewDecFromInt(sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	valPower := int64(100)
	// slash the validator by 50% (simulated with manual calls; we assume the validator is bonded)
	distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower,
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)

	// slash the validator by 50% again
	// stakingKeeper.Slash(ctx, valConsAddr0, ctx.BlockHeight(), valPower/2, math.LegacyNewDecWithPrec(5, 1))
	distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower/2,
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)

	// increase block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards should be half the tokens
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, rewards)

	// commission should be the other half
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, valCommission.Commission)
}

func TestCalculateRewardsMultiDelegatorMultiSlash(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valPower := int64(100)

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, sdk.TokensFromConsensusPower(valPower, sdk.DefaultPowerReduction))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// validator and delegation mocks
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(3)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(2)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := math.LegacyNewDecFromInt(sdk.TokensFromConsensusPower(30, sdk.DefaultPowerReduction))
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// slash the validator
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower,
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// update validator mock
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(1)

	// second delegation
	_, del2, err := distrtestutil.Delegate(
		ctx,
		distrKeeper,
		sdk.AccAddress(valConsAddr1),
		&val,
		sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		nil,
		stakingKeeper,
	)
	require.NoError(t, err)

	// new delegation mock and update validator mock
	stakingKeeper.EXPECT().Delegation(gomock.Any(), sdk.AccAddress(valConsAddr1), valAddr).Return(del2, nil)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(1)

	// call necessary hooks to update a delegation
	err = distrKeeper.Hooks().AfterDelegationModified(ctx, sdk.AccAddress(valConsAddr1), valAddr)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// slash the validator again
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)
	distrtestutil.SlashValidator(
		ctx,
		valConsAddr0,
		ctx.BlockHeight(),
		valPower,
		math.LegacyNewDecWithPrec(5, 1),
		&val,
		&distrKeeper,
		stakingKeeper,
	)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 3)

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards for del1 should be 2/3 initial (half initial first period, 1/6 initial second period)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(2).Add(initial.QuoInt64(6))}}, rewards)

	// calculate delegation rewards for del2
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)
	require.NoError(t, err)

	// rewards for del2 should be initial / 3
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial.QuoInt64(3)}}, rewards)

	// commission should be equal to initial (twice 50% commission, unaffected by slashing)
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: initial}}, valCommission.Commission)
}

func TestCalculateRewardsMultiDelegatorMultiWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 50% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	// validator and delegation mocks
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(3)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).Times(5)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(2)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, math.NewInt(initial))}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// historical count should be 2 (validator init, delegation init)
	require.Equal(t, uint64(2), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// second delegation
	_, del2, err := distrtestutil.Delegate(
		ctx,
		distrKeeper,
		sdk.AccAddress(valConsAddr1),
		&val,
		math.NewInt(100),
		nil,
		stakingKeeper,
	)
	require.NoError(t, err)

	// new delegation mock and update validator mock
	stakingKeeper.EXPECT().Delegation(gomock.Any(), sdk.AccAddress(valConsAddr1), valAddr).Return(del2, nil).Times(3)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(6)

	// call necessary hooks to update a delegation
	err = distrKeeper.Hooks().AfterDelegationModified(ctx, sdk.AccAddress(valConsAddr1), valAddr)
	require.NoError(t, err)

	// historical count should be 3 (second delegation init)
	require.Equal(t, uint64(3), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// first delegator withdraws
	expRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(initial*3/4))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expRewards)
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, addr, valAddr)
	require.NoError(t, err)

	// second delegator withdraws
	expRewards = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(initial*1/4))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, sdk.AccAddress(valConsAddr1), expRewards)
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(valConsAddr1), valAddr)
	require.NoError(t, err)

	// historical count should be 3 (validator init + two delegations)
	require.Equal(t, uint64(3), distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// validator withdraws commission
	expCommission := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(initial))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expCommission)
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddr)
	require.NoError(t, err)

	// end period
	endingPeriod, _ := distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, err := distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)
	require.NoError(t, err)

	// rewards for del2 should be zero
	require.True(t, rewards.IsZero())

	// commission should be zero
	valCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.True(t, valCommission.Commission.IsZero())

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// first delegator withdraws again
	expCommission = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(initial*1/4))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expCommission)
	_, err = distrKeeper.WithdrawDelegationRewards(ctx, addr, valAddr)
	require.NoError(t, err)

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards for del1 should be zero
	require.True(t, rewards.IsZero())

	// calculate delegation rewards for del2
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)
	require.NoError(t, err)

	// rewards for del2 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 4)}}, rewards)

	// commission should be half initial
	valCommission, err = distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, valCommission.Commission)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// withdraw commission
	expCommission = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(initial))}
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, disttypes.ModuleName, addr, expCommission)
	_, err = distrKeeper.WithdrawValidatorCommission(ctx, valAddr)
	require.NoError(t, err)

	// end period
	endingPeriod, _ = distrKeeper.IncrementValidatorPeriod(ctx, val)

	// calculate delegation rewards for del1
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del, endingPeriod)
	require.NoError(t, err)

	// rewards for del1 should be 1/4 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 4)}}, rewards)

	// calculate delegation rewards for del2
	rewards, err = distrKeeper.CalculateDelegationRewards(ctx, val, del2, endingPeriod)
	require.NoError(t, err)

	// rewards for del2 should be 1/2 initial
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initial / 2)}}, rewards)

	// commission should be zero
	valCommission, err = distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.True(t, valCommission.Commission.IsZero())
}

func Test100PercentCommissionReward(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 100% commission
	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(10, 1), math.LegacyNewDecWithPrec(10, 1), math.LegacyNewDec(0))

	// validator and delegation mocks
	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(3)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).Times(3)

	// run the necessary hooks manually (given that we are not running an actual staking module)
	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).Times(2)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	initial := int64(20)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, math.NewInt(initial))}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// allocate some more rewards
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	rewards, err := distrKeeper.WithdrawDelegationRewards(ctx, addr, valAddr)
	require.NoError(t, err)

	zeroRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())}
	require.True(t, rewards.Equal(zeroRewards))

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

func TestWithdrawDelegationRewards_BlockedWithdrawAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// set withdraw address to address that we will block
	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, addr, blockedAddr))

	// allocate rewards so there's something to withdraw
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// snapshot accounting state we expect to remain untouched on the user path
	outstandingBefore, err := distrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valAddr)
	require.NoError(t, err)
	startingInfoBefore, err := distrKeeper.GetDelegatorStartingInfo(ctx, valAddr, addr)
	require.NoError(t, err)
	histRefBefore := distrKeeper.GetValidatorHistoricalReferenceCount(ctx)

	// blocked addr is now blocked in the bank keeper
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)

	// reset events so we can assert just the ones from this call
	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())

	// try and withdraw rewards
	_, err = distrKeeper.WithdrawDelegationRewards(sdkCtx, addr, valAddr)
	require.ErrorIs(t, err, disttypes.ErrWithdrawAddrBlocked)

	// accounting state must be unchanged
	outstandingAfter, err := distrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, outstandingBefore, outstandingAfter)
	startingInfoAfter, err := distrKeeper.GetDelegatorStartingInfo(ctx, valAddr, addr)
	require.NoError(t, err)
	require.Equal(t, startingInfoBefore, startingInfoAfter)
	require.Equal(t, histRefBefore, distrKeeper.GetValidatorHistoricalReferenceCount(ctx))

	// ensure the blocked address event was emitted
	var sawBlockedEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type == disttypes.EventTypeWithdrawAddrBlocked {
			sawBlockedEvent = true
			break
		}
	}
	require.True(t, sawBlockedEvent, "expected withdraw_addr_blocked event")
}

func TestBeforeDelegationSharesModified_FallbackDelegator(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// set withdraw address to address that will become blocked
	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, addr, blockedAddr))

	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	expRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial)}

	// block withdraw address but NOT  delegator address
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)
	bankKeeper.EXPECT().BlockedAddr(addr).Return(false).Times(1)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, addr, expRewards).Return(nil).Times(1)

	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())

	// hook entry point uses the fallback resolver by default
	require.NoError(t, distrKeeper.Hooks().BeforeDelegationSharesModified(sdkCtx, addr, valAddr))

	// ensure the redirect event was emitted with both original + final
	// addresses
	var sawRedirectEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type != disttypes.EventTypeWithdrawAddrRedirected {
			continue
		}
		sawRedirectEvent = true
		var foundOriginal, foundFinal bool
		for _, attr := range ev.Attributes {
			if attr.Key == disttypes.AttributeKeyOriginalWithdrawAddress && attr.Value == blockedAddr.String() {
				foundOriginal = true
			}
			if attr.Key == disttypes.AttributeKeyWithdrawAddress && attr.Value == addr.String() {
				foundFinal = true
			}
		}
		require.True(t, foundOriginal && foundFinal, "redirect event missing expected address attributes")
	}
	require.True(t, sawRedirectEvent, "expected withdraw_addr_redirected event")
}

func TestBeforeDelegationSharesModified_FallbackPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr) // delegator address — will also be blocked
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, addr, blockedAddr))

	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	expRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial)}

	// both addresses blocked — no Send call should happen.
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)
	bankKeeper.EXPECT().BlockedAddr(addr).Return(true).Times(1)

	cpBefore, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)

	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())

	// hook entry point uses the fallback resolver by default
	require.NoError(t, distrKeeper.Hooks().BeforeDelegationSharesModified(sdkCtx, addr, valAddr))

	// community-pool fallback: redirect event was emitted with the
	// "community_pool" destination value.
	var sawCPEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type != disttypes.EventTypeWithdrawAddrRedirected {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == disttypes.AttributeKeyWithdrawAddress && attr.Value == disttypes.AttributeValueCommunityPool {
				sawCPEvent = true
				break
			}
		}
	}
	require.True(t, sawCPEvent, "expected redirect-to-community-pool event")

	// rewards were credited to the community pool. The truncated finalRewards
	// plus the (here zero) decimal remainder must all land in the pool.
	cpAfter, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	diff := cpAfter.CommunityPool.Sub(cpBefore.CommunityPool)
	require.True(t,
		diff.AmountOf(sdk.DefaultBondDenom).Equal(math.LegacyNewDecFromInt(expRewards[0].Amount)),
		"community pool delta should equal the truncated reward amount; got %s want %s",
		diff, expRewards[0].Amount,
	)

	// outstanding rewards were decremented
	outstanding, err := distrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valAddr)
	require.NoError(t, err)
	require.True(t, outstanding.IsZero(), "expected outstanding rewards to be fully drained, got %s", outstanding)
}

func TestBeforeDelegationSharesModified_Strict(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, addr, blockedAddr))

	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// strict should error on the original withdraw address being blocked
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)

	strictCtx := stakingtypes.WithStrictWithdraw(ctx.WithEventManager(sdk.NewEventManager()))
	err = distrKeeper.Hooks().BeforeDelegationSharesModified(strictCtx, addr, valAddr)
	require.ErrorIs(t, err, disttypes.ErrWithdrawAddrBlocked)
}

func TestAfterValidatorRemoved_FallbackDelegator(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	accAddr := sdk.AccAddress(valAddr) // validator owner account
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	// 50% commission rate so commission accrues
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(accAddr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), accAddr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, accAddr, valAddr))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// point validator's withdraw address at one that will be blocked
	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, accAddr, blockedAddr))

	// allocate tokens — half go to commission, half to delegator pool
	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	expCommission := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial.Quo(math.NewInt(2)))}

	// block withdraw address but not the validator owner account
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)
	bankKeeper.EXPECT().BlockedAddr(accAddr).Return(false).Times(1)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, accAddr, expCommission).Return(nil).Times(1)

	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())

	require.NoError(t, distrKeeper.Hooks().AfterValidatorRemoved(sdkCtx, valConsAddr0, valAddr))

	// redirect event was emitted with both the original blocked addr and the
	// fallback (validator owner) addr.
	var sawRedirectEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type != disttypes.EventTypeWithdrawAddrRedirected {
			continue
		}
		sawRedirectEvent = true
		var foundOriginal, foundFinal bool
		for _, attr := range ev.Attributes {
			if attr.Key == disttypes.AttributeKeyOriginalWithdrawAddress && attr.Value == blockedAddr.String() {
				foundOriginal = true
			}
			if attr.Key == disttypes.AttributeKeyWithdrawAddress && attr.Value == accAddr.String() {
				foundFinal = true
			}
		}
		require.True(t, foundOriginal && foundFinal, "redirect event missing expected address attributes")
	}
	require.True(t, sawRedirectEvent, "expected withdraw_addr_redirected event")
}

func TestAfterValidatorRemoved_FallbackPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	accAddr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(accAddr.String(), valAddr.String(), val.DelegatorShares)
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	stakingKeeper.EXPECT().Delegation(gomock.Any(), accAddr, valAddr).Return(del, nil).AnyTimes()

	require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, accAddr, valAddr))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	blockedAddr := sdk.AccAddress(valConsAddr1)
	require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, accAddr, blockedAddr))

	initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	tokens := sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// both addresses blocked — commission goes to community pool, no Send call
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)
	bankKeeper.EXPECT().BlockedAddr(accAddr).Return(true).Times(1)

	cpBefore, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)

	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())

	require.NoError(t, distrKeeper.Hooks().AfterValidatorRemoved(sdkCtx, valConsAddr0, valAddr))

	// redirect event names the community pool as the destination
	var sawCPEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type != disttypes.EventTypeWithdrawAddrRedirected {
			continue
		}
		for _, attr := range ev.Attributes {
			if attr.Key == disttypes.AttributeKeyWithdrawAddress && attr.Value == disttypes.AttributeValueCommunityPool {
				sawCPEvent = true
				break
			}
		}
	}
	require.True(t, sawCPEvent, "expected redirect-to-community-pool event")

	// community pool gained the full allocated amount: commission redirected
	// to the pool plus the leftover (dust) outstanding rewards.
	cpAfter, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	diff := cpAfter.CommunityPool.Sub(cpBefore.CommunityPool)
	require.True(t,
		diff.AmountOf(sdk.DefaultBondDenom).Equal(math.LegacyNewDecFromInt(initial)),
		"community pool delta should equal the full allocation; got %s want %s",
		diff, initial,
	)
}
