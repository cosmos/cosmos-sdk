package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestSetWithdrawAddr(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(2)

	delegatorAddr := addrs[0]
	withdrawAddr := addrs[1]

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	bankKeeper.EXPECT().BlockedAddr(withdrawAddr).Return(false).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(distrAcc.GetAddress()).Return(true).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	params := types.DefaultParams()
	params.WithdrawAddrEnabled = false
	require.NoError(t, distrKeeper.Params.Set(ctx, params))

	err := distrKeeper.SetWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
	require.NotNil(t, err)

	params.WithdrawAddrEnabled = true
	require.NoError(t, distrKeeper.Params.Set(ctx, params))

	err = distrKeeper.SetWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
	require.Nil(t, err)

	require.Error(t, distrKeeper.SetWithdrawAddr(ctx, delegatorAddr, distrAcc.GetAddress()))
}

func TestWithdrawValidatorCommission(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(1)

	valAddr := sdk.ValAddress(addrs[0])

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// set outstanding rewards
	require.NoError(t, distrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission}))

	// set commission
	require.NoError(t, distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission}))

	// withdraw commission
	coins := sdk.NewCoins(sdk.NewCoin("mytoken", math.NewInt(1)), sdk.NewCoin("stake", math.NewInt(1)))
	// if SendCoinsFromModuleToAccount is called, we know that the withdraw was successful
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), "distribution", addrs[0], coins).Return(nil)

	_, err := distrKeeper.WithdrawValidatorCommission(ctx, valAddr)
	require.NoError(t, err)

	// check remainder
	remainderValCommission, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	remainder := remainderValCommission.Commission
	require.Equal(t, sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(1).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1).Quo(math.LegacyNewDec(2))),
	}, remainder)
}

func TestGetTotalRewards(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(2)

	valAddr0 := sdk.ValAddress(addrs[0])
	valAddr1 := sdk.ValAddress(addrs[1])

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	require.NoError(t, distrKeeper.SetValidatorOutstandingRewards(ctx, valAddr0, types.ValidatorOutstandingRewards{Rewards: valCommission}))
	require.NoError(t, distrKeeper.SetValidatorOutstandingRewards(ctx, valAddr1, types.ValidatorOutstandingRewards{Rewards: valCommission}))

	expectedRewards := valCommission.MulDec(math.LegacyNewDec(2))
	totalRewards := distrKeeper.GetTotalRewards(ctx)

	require.Equal(t, expectedRewards, totalRewards)
}

func TestFundCommunityPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(1)

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

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
	require.NoError(t, distrKeeper.FeePool.Set(ctx, types.InitialFeePool()))

	initPool, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.Empty(t, initPool.CommunityPool)

	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), addrs[0], "distribution", amount).Return(nil)
	err = distrKeeper.FundCommunityPool(ctx, amount, addrs[0])
	require.NoError(t, err)

	feePool, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), feePool.CommunityPool)
}

func TestWithdrawDelegationRewards_NoDelegation(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(2)

	delegatorAddr := addrs[0]
	valAddr := sdk.ValAddress(addrs[1])

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// Create validator
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	// Mock validator exists
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil)

	// Mock delegation returns error (no delegation exists)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), delegatorAddr, valAddr).Return(nil, stakingtypes.ErrNoDelegation)

	// Try to withdraw delegation rewards when no delegation exists and no outstanding rewards
	rewards, err := distrKeeper.WithdrawDelegationRewards(ctx, delegatorAddr, valAddr)

	// Should return an error since there's no delegation and no outstanding rewards
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrEmptyDelegationDistInfo)
	require.Nil(t, rewards)
}

func TestWithdrawDelegationRewards_NoDelegationWithOutstandingRewards(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(2)

	delegatorAddr := addrs[0]
	valAddr := sdk.ValAddress(addrs[1])

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// Create validator
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	// Set up outstanding rewards
	outstandingRewards := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	err = distrKeeper.UserOutstandingRewards.Set(ctx, collections.Join(delegatorAddr, valAddr), types.UserOutstandingRewards{
		Rewards: outstandingRewards,
	})
	require.NoError(t, err)

	// Mock validator exists
	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil)

	// Mock delegation returns error (no delegation exists)
	stakingKeeper.EXPECT().Delegation(gomock.Any(), delegatorAddr, valAddr).Return(nil, stakingtypes.ErrNoDelegation)

	// Mock GetAccount returns nil (not a vesting account)
	accountKeeper.EXPECT().GetAccount(gomock.Any(), delegatorAddr).Return(nil)

	// Expect bank keeper to send the outstanding rewards
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, delegatorAddr, outstandingRewards).Return(nil)

	// Try to withdraw delegation rewards when no delegation exists but there are outstanding rewards
	rewards, err := distrKeeper.WithdrawDelegationRewards(ctx, delegatorAddr, valAddr)

	// Should succeed and return the outstanding rewards
	require.NoError(t, err)
	require.Equal(t, outstandingRewards, rewards)

	// Verify outstanding rewards were removed
	_, err = distrKeeper.UserOutstandingRewards.Get(ctx, collections.Join(delegatorAddr, valAddr))
	require.ErrorIs(t, err, collections.ErrNotFound)
}
