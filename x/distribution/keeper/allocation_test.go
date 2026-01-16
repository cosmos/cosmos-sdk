package keeper_test

import (
	"errors"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	sdkaddress "cosmossdk.io/core/address"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
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

type suite struct {
	ctx             sdk.Context
	distrKeeper     keeper.Keeper
	stakingKeeper   *distrtestutil.MockStakingKeeper
	accountKeeper   *distrtestutil.MockAccountKeeper
	bankKeeper      *distrtestutil.MockBankKeeper
	feeCollectorAcc *authtypes.ModuleAccount
	valCodec        sdkaddress.Codec
}

func setupTestKeeper(t *testing.T, nakamotoBonusCoefficient math.LegacyDec, height uint64) *suite {
	t.Helper()

	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now(), Height: int64(height)})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress()).AnyTimes()
	valCodec := address.NewBech32Codec("cosmosvaloper")
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(valCodec).AnyTimes()
	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc).AnyTimes()

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

	params, err := distrKeeper.Params.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		params = disttypes.DefaultParams()
	} else {
		require.NoError(t, err)
	}
	require.NoError(t, distrKeeper.Params.Set(ctx, params))

	err = distrKeeper.SetNakamotoBonusCoefficient(ctx, nakamotoBonusCoefficient)
	require.NoError(t, err)

	return &suite{
		ctx:             ctx,
		distrKeeper:     distrKeeper,
		stakingKeeper:   stakingKeeper,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		feeCollectorAcc: feeCollectorAcc,
		valCodec:        valCodec,
	}
}

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyZeroDec(), 100)

	// create a validator with 50% commission
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val, nil).AnyTimes()

	// allocate tokens
	tokens := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(10)},
	}
	require.NoError(t, s.distrKeeper.AllocateTokensToValidator(s.ctx, val, tokens))

	// check commission
	expected := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(5)},
	}

	valBz, err := s.valCodec.StringToBytes(val.GetOperator())
	require.NoError(t, err)

	valCommission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valBz)
	require.NoError(t, err)
	require.Equal(t, expected, valCommission.Commission)

	// check current rewards
	currentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valBz)
	require.NoError(t, err)
	require.Equal(t, expected, currentRewards.Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyZeroDec(), 100)

	// reset fee pool & set params
	require.NoError(t, s.distrKeeper.Params.Set(s.ctx, disttypes.DefaultParams()))
	require.NoError(t, s.distrKeeper.FeePool.Set(s.ctx, disttypes.InitialFeePool()))

	// create validator with 50% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 0% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// assert the initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	val0OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsZero())

	val1OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsZero())

	feePool, err := s.distrKeeper.FeePool.Get(s.ctx)
	require.NoError(t, err)
	require.True(t, feePool.CommunityPool.IsZero())

	val0Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0Commission.Commission.IsZero())

	val1Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	val0CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0CurrentRewards.Rewards.IsZero())

	val1CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1CurrentRewards.Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{Validator: abci.Validator{Address: valConsPk0.Address(), Power: 100}},
		{Validator: abci.Validator{Address: valConsPk1.Address(), Power: 100}},
	}
	require.NoError(t, s.distrKeeper.AllocateTokens(s.ctx, 200, votes))

	// 98 outstanding rewards (100 less 2 to community pool)
	val0OutstandingRewards, err = s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val0OutstandingRewards.Rewards)

	val1OutstandingRewards, err = s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1OutstandingRewards.Rewards)

	// 2 community pool coins
	feePool, err = s.distrKeeper.FeePool.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(2)}}, feePool.CommunityPool)

	// 50% commission for first proposer, (0.5 * 98%) * 100 / 2 = 23.25
	val0Commission, err = s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0Commission.Commission)

	// zero commission for second proposer
	val1Commission, err = s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	// just staking.proportional for first proposer less commission = (0.5 * 98%) * 100 / 2 = 24.50
	val0CurrentRewards, err = s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0CurrentRewards.Rewards)

	// proposer reward + staking.proportional for second proposer = (0.5 * (98%)) * 100 = 49
	val1CurrentRewards, err = s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1CurrentRewards.Rewards)
}

func TestAllocateTokens_NakamotoBonusDisabled(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyNewDecWithPrec(5, 2), 100) // η = 0.05 (should not matter since disabled)

	// Set nakamoto_bonus_enabled parameter to false
	params, err := s.distrKeeper.Params.Get(s.ctx)
	require.NoError(t, err)
	params.NakamotoBonus.Enabled = false
	require.NoError(t, s.distrKeeper.Params.Set(s.ctx, params))

	// η can be any value, should have no effect
	err = s.distrKeeper.SetNakamotoBonusCoefficient(s.ctx, math.LegacyNewDecWithPrec(5, 2))
	require.NoError(t, err)

	// Setup validators: two validators, equal power, 0% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(
		math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(
		math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	abciValA := abci.Validator{
		Address: valConsPk0.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{Validator: abciValA},
		{Validator: abciValB},
	}

	require.NoError(t, s.distrKeeper.AllocateTokens(s.ctx, 200, votes))

	// With nakamoto_bonus_enabled = false, all rewards should be proportional (no bonus)
	// Community tax is 2%, so 98 left, each validator gets 49
	expectedReward := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}
	var expectedCommission sdk.DecCoins

	val0OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, expectedReward, val0OutstandingRewards.Rewards)

	val1OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, expectedReward, val1OutstandingRewards.Rewards)

	feePool, err := s.distrKeeper.FeePool.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDecCoinsFromCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2))), feePool.CommunityPool)

	val0Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, expectedCommission, val0Commission.Commission)

	val1Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, expectedCommission, val1Commission.Commission)

	val0CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, expectedReward, val0CurrentRewards.Rewards)

	val1CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, expectedReward, val1CurrentRewards.Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyZeroDec(), 100)

	// reset fee pool
	require.NoError(t, s.distrKeeper.FeePool.Set(s.ctx, disttypes.InitialFeePool()))
	require.NoError(t, s.distrKeeper.Params.Set(s.ctx, disttypes.DefaultParams()))

	// create a validator with 10% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 10% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// create third validator with 10% commission
	valAddr2 := sdk.ValAddress(valConsAddr2)
	val2, err := stakingtypes.NewValidator(sdk.ValAddress(valConsAddr2).String(), valConsPk1, stakingtypes.Description{})
	require.NoError(t, err)
	val2.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).Return(val2, nil).AnyTimes()

	// assert the initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	val0OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsZero())

	val1OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsZero())

	feePool, err := s.distrKeeper.FeePool.Get(s.ctx)
	require.NoError(t, err)
	require.True(t, feePool.CommunityPool.IsZero())

	val0Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0Commission.Commission.IsZero())

	val1Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	val0CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0CurrentRewards.Rewards.IsZero())

	val1CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1CurrentRewards.Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{Validator: abci.Validator{Address: valConsPk0.Address(), Power: 11}},
		{Validator: abci.Validator{Address: valConsPk1.Address(), Power: 10}},
		{Validator: abci.Validator{Address: valConsPk2.Address(), Power: 10}},
	}
	require.NoError(t, s.distrKeeper.AllocateTokens(s.ctx, 31, votes))

	val0OutstandingRewards, err = s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsValid())

	val1OutstandingRewards, err = s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsValid())

	val2OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr2)
	require.NoError(t, err)
	require.True(t, val2OutstandingRewards.Rewards.IsValid())
}

func TestAllocateTokensWithNakamotoBonus(t *testing.T) {
	s := setupTestKeeper(t, math.LegacyNewDecWithPrec(2, 1), 100) // η = 0.20

	require.NoError(t, s.distrKeeper.FeePool.Set(s.ctx, disttypes.InitialFeePool()))

	// Create validators with imbalanced power distribution
	// High power validator (80% of voting power)
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(800))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(
		math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// Low power validator (20% of voting power)
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(200))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(
		math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// 1000 uatom collected
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{Validator: abci.Validator{Address: valConsPk0.Address(), Power: 800}},
		{Validator: abci.Validator{Address: valConsPk1.Address(), Power: 200}},
	}

	require.NoError(t, s.distrKeeper.AllocateTokens(s.ctx, 1000, votes))

	// Expectation with η = 0.20:
	// - 2% community tax → 20
	// - 980 left to distribute
	// - Nakamoto bonus pool: 0.20 * 980 = 196
	// - Proportional pool: 980 - 196 = 784
	//
	// Nakamoto bonus (equal per validator):
	// - val0: 196 / 2 = 98
	// - val1: 196 / 2 = 98
	//
	// Proportional distribution (by voting power):
	// - val0: 784 * (800 / 1000) = 627.2
	// - val1: 784 * (200 / 1000) = 156.8
	//
	// Total rewards:
	// - val0: 98 + 627.2 = 725.2
	// - val1: 98 + 156.8 = 254.8

	expectedVal0Reward := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(7252, 1)}, // 725.2
	}
	expectedVal1Reward := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2548, 1)}, // 254.8
	}

	val0OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, expectedVal0Reward, val0OutstandingRewards.Rewards)

	val1OutstandingRewards, err := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, expectedVal1Reward, val1OutstandingRewards.Rewards)

	feePool, err := s.distrKeeper.FeePool.Get(s.ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDecCoinsFromCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20))), feePool.CommunityPool)

	// Zero commission for both validators
	val0Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0Commission.Commission.IsZero())

	val1Commission, err := s.distrKeeper.GetValidatorAccumulatedCommission(s.ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	val0CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, expectedVal0Reward, val0CurrentRewards.Rewards)

	val1CurrentRewards, err := s.distrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, expectedVal1Reward, val1CurrentRewards.Rewards)

	// Verify RPS (rewards per stake) - smaller validator should have higher RPS
	// val0 RPS: 725.2 / 800 = 0.9065
	// val1 RPS: 254.8 / 200 = 1.274
	val0RPS := val0OutstandingRewards.Rewards.AmountOf(sdk.DefaultBondDenom).Quo(math.LegacyNewDec(800))
	val1RPS := val1OutstandingRewards.Rewards.AmountOf(sdk.DefaultBondDenom).Quo(math.LegacyNewDec(200))

	require.True(t, val1RPS.GT(val0RPS),
		"Small validator RPS should be higher due to Nakamoto Bonus: val1_rps=%s, val0_rps=%s", val1RPS, val0RPS)
}
