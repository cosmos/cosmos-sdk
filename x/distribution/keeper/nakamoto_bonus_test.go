package keeper_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func createValidators(ctx context.Context, stakingKeeper *distrtestutil.MockStakingKeeper, powers ...int64) {
	vals := make([]stakingtypes.Validator, len(powers))
	for i, p := range powers {
		vals[i] = stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress([]byte{byte(i)}).String(),
			Tokens:          math.NewInt(p),
			Status:          stakingtypes.Bonded,
			Commission:      stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
		}
	}
	stakingKeeper.EXPECT().GetBondedValidatorsByPower(ctx).Return(vals, nil).AnyTimes()

	// Mock ValidatorByConsAddr for AllocateTokens
	for i := range vals {
		stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(vals[i], nil).AnyTimes()
	}
}

func TestAdjustEta_NakamotoDisabled(t *testing.T) {
	initialEta := math.LegacyNewDecWithPrec(5, 2) // 0.05
	s := setupTestKeeper(t, initialEta, types.DefaultNakamotoBonusPeriod)

	// Disable the feature
	p, err := s.distrKeeper.Params.Get(s.ctx)
	require.NoError(t, err)
	p.NakamotoBonus.Enabled = false
	require.NoError(t, s.distrKeeper.Params.Set(s.ctx, p))

	// Even with validators that would trigger an increase, should not adjust
	createValidators(s.ctx, s.stakingKeeper, 100, 100, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	// η should remain unchanged (still 0.05)
	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	require.Equal(t, initialEta, nakamotoBonusCoefficient,
		"η should not change when feature is disabled")
}

func TestAdjustEta_NoInterval(t *testing.T) {
	initialEta := types.DefaultNakamotoBonusStep
	s := setupTestKeeper(t, initialEta, 119_999) // period configured but height doesn't match

	// Block height is the default from setupTestKeeper (DefaultNakamotoBonusPeriod)
	// Since height % period != 0, should skip adjustment
	createValidators(s.ctx, s.stakingKeeper, 10, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	require.Equal(t, initialEta, nakamotoBonusCoefficient,
		"η should not change when not on adjustment block")
}

func TestAdjustEta_NotEnoughValidators(t *testing.T) {
	initialEta := types.DefaultNakamotoBonusStep
	s := setupTestKeeper(t, initialEta, types.DefaultNakamotoBonusPeriod)

	// Only 2 validators (need at least 3 for grouping)
	createValidators(s.ctx, s.stakingKeeper, 10, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	require.Equal(t, initialEta, nakamotoBonusCoefficient,
		"η should not change with fewer than 3 validators")
}

func TestAdjustEta_Increase(t *testing.T) {
	initialEta := math.LegacyNewDecWithPrec(5, 2) // Start at 0.05 (above minimum)
	s := setupTestKeeper(t, initialEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 100, lowAvg = 10, ratio = 10 >= 3, should increase
	createValidators(s.ctx, s.stakingKeeper, 100, 100, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	expectedEta := initialEta.Add(types.DefaultNakamotoBonusStep) // 0.05 + 0.01 = 0.06
	require.Equal(t, expectedEta, nakamotoBonusCoefficient,
		"η should increase by step when ratio >= 3")
}

func TestAdjustEta_Decrease(t *testing.T) {
	initialEta := math.LegacyNewDecWithPrec(5, 2) // Start at 0.05 (above minimum)
	s := setupTestKeeper(t, initialEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 20, lowAvg = 10, ratio = 2 < 3, should decrease
	createValidators(s.ctx, s.stakingKeeper, 20, 20, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	// 0.05 - 0.01 = 0.04, which is above minimum, so no clamping needed
	expectedEta := math.LegacyNewDecWithPrec(4, 2)
	require.Equal(t, expectedEta, nakamotoBonusCoefficient,
		"η should decrease by step when ratio < 3")
}

func TestAdjustEta_ClampZero(t *testing.T) {
	initEta := types.DefaultNakamotoBonusMinimumCoefficient // Start at minimum (0.03)
	s := setupTestKeeper(t, initEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 20, lowAvg = 10, ratio = 2 < 3, would decrease but already at min
	createValidators(s.ctx, s.stakingKeeper, 20, 20, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	require.True(t, nakamotoBonusCoefficient.GTE(types.DefaultNakamotoBonusMinimumCoefficient),
		"η should never go below minimum (%s), got: %s", types.DefaultNakamotoBonusMinimumCoefficient, nakamotoBonusCoefficient)
	require.Equal(t, types.DefaultNakamotoBonusMinimumCoefficient, nakamotoBonusCoefficient)
}

func TestAdjustEta_ClampOne(t *testing.T) {
	initEta := types.DefaultNakamotoBonusMaximumCoefficient // Start at maximum (1.0)
	s := setupTestKeeper(t, initEta, types.DefaultNakamotoBonusPeriod)

	// highAvg = 100, lowAvg = 10, ratio = 10 >= 3, would increase but already at max
	createValidators(s.ctx, s.stakingKeeper, 100, 100, 10)

	require.NoError(t, s.distrKeeper.AdjustNakamotoBonusCoefficient(s.ctx))

	nakamotoBonusCoefficient, err := s.distrKeeper.GetNakamotoBonusCoefficient(s.ctx)
	require.NoError(t, err)
	require.True(t, nakamotoBonusCoefficient.LTE(types.DefaultNakamotoBonusMaximumCoefficient),
		"η should never exceed maximum (%s), got: %s", types.DefaultNakamotoBonusMaximumCoefficient, nakamotoBonusCoefficient)
	require.Equal(t, types.DefaultNakamotoBonusMaximumCoefficient, nakamotoBonusCoefficient)
}

func TestAllocateTokensWithNakamotoBonusImbalancedValidators(t *testing.T) {
	// Realistic scenario: top 5 validators have 70% of stake
	s := setupTestKeeper(t, math.LegacyNewDecWithPrec(5, 2), 100) // η = 0.05

	require.NoError(t, s.distrKeeper.FeePool.Set(s.ctx, types.InitialFeePool()))

	// Create validators with realistic power distribution
	// High power validators
	val0, _ := distrtestutil.CreateValidator(valConsPk0, math.NewInt(1000))
	val0.Commission = stakingtypes.NewCommission(
		math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	val1, _ := distrtestutil.CreateValidator(valConsPk1, math.NewInt(800))
	val1.Commission = stakingtypes.NewCommission(
		math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// Low power validator (benefits from Nakamoto bonus)
	val2, _ := distrtestutil.CreateValidator(valConsPk2, math.NewInt(100))
	val2.Commission = stakingtypes.NewCommission(
		math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyZeroDec(),
	)
	s.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).Return(val2, nil).AnyTimes()

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1000)))
	s.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), s.feeCollectorAcc.GetAddress()).Return(fees)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", types.ModuleName, fees)

	votes := []abci.VoteInfo{
		{Validator: abci.Validator{Address: valConsPk0.Address(), Power: 1000}},
		{Validator: abci.Validator{Address: valConsPk1.Address(), Power: 800}},
		{Validator: abci.Validator{Address: valConsPk2.Address(), Power: 100}},
	}

	require.NoError(t, s.distrKeeper.AllocateTokens(s.ctx, 1900, votes))

	// Verify: Small validator should have higher RPS due to fixed bonus
	val0Rewards, _ := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, sdk.ValAddress(valConsAddr0))
	val2Rewards, _ := s.distrKeeper.GetValidatorOutstandingRewards(s.ctx, sdk.ValAddress(valConsAddr2))

	// Calculate RPS (rewards per stake)
	val0RPS := val0Rewards.Rewards.AmountOf(sdk.DefaultBondDenom).Quo(math.LegacyNewDec(1000))
	val2RPS := val2Rewards.Rewards.AmountOf(sdk.DefaultBondDenom).Quo(math.LegacyNewDec(100))

	// val2 should have higher RPS
	require.True(t, val2RPS.GT(val0RPS),
		"Small validator RPS should be higher: val2_rps=%s, val0_rps=%s", val2RPS, val0RPS)
}
