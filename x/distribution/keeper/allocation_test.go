package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/distribution/keeper"
	distrtestutil "cosmossdk.io/x/distribution/testutil"
	disttypes "cosmossdk.io/x/distribution/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModule{})
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)
	poolKeeper := distrtestutil.NewMockPoolKeeper(ctrl)

	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger())

	valCodec := address.NewBech32Codec("cosmosvaloper")

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(valCodec).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		env,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		poolKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// create validator with 50% commission
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val, nil).AnyTimes()

	// allocate tokens
	tokens := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(10)},
	}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))

	// check commission
	expected := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(5)},
	}

	valBz, err := valCodec.StringToBytes(val.GetOperator())
	require.NoError(t, err)

	valCommission, err := distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valBz)
	require.NoError(t, err)
	require.Equal(t, expected, valCommission.Commission)

	// check current rewards
	currentRewards, err := distrKeeper.ValidatorCurrentRewards.Get(ctx, valBz)
	require.NoError(t, err)
	require.Equal(t, expected, currentRewards.Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModule{})
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)
	poolKeeper := distrtestutil.NewMockPoolKeeper(ctrl)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc)
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()

	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		env,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		poolKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool & set params
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))

	// create validator with 50% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 0% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	abciValA := comet.Validator{
		Address: valConsPk0.Address(),
		Power:   100,
	}
	abciValB := comet.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	_, err = distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	feePool, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(t, feePool.DecimalPool.IsZero())

	_, err = distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound) // require no rewards

	_, err = distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound) // require no rewards

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees)
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), disttypes.ModuleName, disttypes.ProtocolPoolModuleName, sdk.Coins{{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(2)}}) // 2 community pool coins
	poolKeeper.EXPECT().SetToDistribute(ctx, gomock.Any(), gomock.Any())

	votes := []comet.VoteInfo{
		{
			Validator: abciValA,
		},
		{
			Validator: abciValB,
		},
	}

	require.NoError(t, distrKeeper.AllocateTokens(ctx, 200, votes))

	// 98 outstanding rewards (100 less 2 to community pool)
	val0OutstandingRewards, err := distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val0OutstandingRewards.Rewards)

	val1OutstandingRewards, err := distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1OutstandingRewards.Rewards)

	feePool, err = distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(t, feePool.DecimalPool.IsZero())

	// 50% commission for first proposer, (0.5 * 98%) * 100 / 2 = 23.25
	val0Commission, err := distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0Commission.Commission)

	// zero commission for second proposer
	val1Commission, err := distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	// just staking.proportional for first proposer less commission = (0.5 * 98%) * 100 / 2 = 24.50
	val0CurrentRewards, err := distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0CurrentRewards.Rewards)

	// proposer reward + staking.proportional for second proposer = (0.5 * (98%)) * 100 = 49
	val1CurrentRewards, err := distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1CurrentRewards.Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModule{})
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)
	poolKeeper := distrtestutil.NewMockPoolKeeper(ctrl)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc)
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()

	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		env,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		poolKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// reset fee pool
	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 10% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 10% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// create third validator with 10% commission
	valAddr2 := sdk.ValAddress(valConsAddr2)
	val2, err := stakingtypes.NewValidator(sdk.ValAddress(valConsAddr2).String(), valConsPk1, stakingtypes.Description{})
	require.NoError(t, err)
	val2.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).Return(val2, nil).AnyTimes()

	abciValA := comet.Validator{
		Address: valConsPk0.Address(),
		Power:   11,
	}
	abciValB := comet.Validator{
		Address: valConsPk1.Address(),
		Power:   10,
	}
	abciValC := comet.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	_, err = distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	feePool, err := distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(t, feePool.DecimalPool.IsZero())

	_, err = distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorsAccumulatedCommission.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	_, err = distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr0)
	require.ErrorIs(t, err, collections.ErrNotFound) // require no rewards

	_, err = distrKeeper.ValidatorCurrentRewards.Get(ctx, valAddr1)
	require.ErrorIs(t, err, collections.ErrNotFound) // require no rewards

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees)
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), disttypes.ModuleName, disttypes.ProtocolPoolModuleName, gomock.Any()) // something is sent to community pool
	poolKeeper.EXPECT().SetToDistribute(ctx, gomock.Any(), gomock.Any())

	votes := []comet.VoteInfo{
		{
			Validator: abciValA,
		},
		{
			Validator: abciValB,
		},
		{
			Validator: abciValC,
		},
	}

	require.NoError(t, distrKeeper.AllocateTokens(ctx, 31, votes))

	val0OutstandingRewards, err := distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsValid())

	val1OutstandingRewards, err := distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsValid())

	val2OutstandingRewards, err := distrKeeper.ValidatorOutstandingRewards.Get(ctx, valAddr2)
	require.NoError(t, err)
	require.True(t, val2OutstandingRewards.Rewards.IsValid())
}
