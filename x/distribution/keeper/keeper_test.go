package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()

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

func TestWithdrawValidatorCommission_BlockedAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})
	addrs := simtestutil.CreateIncrementalAccounts(2)

	valAddr := sdk.ValAddress(addrs[0])
	accAddr := sdk.AccAddress(valAddr)
	blockedAddr := addrs[1]

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	valCommission := sdk.DecCoins{
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

	require.NoError(t, distrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission}))
	require.NoError(t, distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission}))

	// point validator's withdraw address at a blocked address
	require.NoError(t, distrKeeper.Params.Set(ctx, types.DefaultParams()))
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(false).Times(1) // allow SetWithdrawAddr to succeed
	require.NoError(t, distrKeeper.SetWithdrawAddr(ctx, accAddr, blockedAddr))

	// strict resolver sees the addr as blocked and errors before any state mutation
	bankKeeper.EXPECT().BlockedAddr(blockedAddr).Return(true).Times(1)

	sdkCtx := ctx.WithEventManager(sdk.NewEventManager())
	_, err := distrKeeper.WithdrawValidatorCommission(sdkCtx, valAddr)
	require.ErrorIs(t, err, types.ErrWithdrawAddrBlocked)

	// accumulated commission and outstanding rewards untouched
	accumAfter, err := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, valCommission, accumAfter.Commission)

	outstandingAfter, err := distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, valCommission, outstandingAfter.Rewards)

	// blocked event was emitted
	var sawBlockedEvent bool
	for _, ev := range sdkCtx.EventManager().Events() {
		if ev.Type == types.EventTypeWithdrawAddrBlocked {
			sawBlockedEvent = true
			break
		}
	}
	require.True(t, sawBlockedEvent, "expected withdraw_addr_blocked event")
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
