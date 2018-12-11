package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestWithdrawValidatorRewardsAllNoDelegator(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	// first make a validator
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, 10)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.AllocateTokens(ctx, sdk.OneDec(), valConsAddr1)

	// withdraw self-delegation reward
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawValidatorRewardsAll(ctx, valOpAddr1)
	amt := accMapper.GetAccount(ctx, valAccAddr1).GetCoins().AmountOf(denom)
	expRes := sdk.NewDec(90).Add(sdk.NewDec(100)).TruncateInt()
	require.True(sdk.IntEq(t, expRes, amt))
}

func TestWithdrawValidatorRewardsAllDelegatorNoCommission(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	msgCreateValidator := stake.NewTestMsgCreateValidator(valOpAddr1, valConsPk1, 10)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// delegate
	msgDelegate := stake.NewTestMsgDelegate(delAddr1, valOpAddr1, 10)
	got = stakeHandler(ctx, msgDelegate)
	require.True(t, got.IsOK())
	amt := accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)
	require.Equal(t, int64(90), amt.Int64())

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.AllocateTokens(ctx, sdk.OneDec(), valConsAddr1)

	// withdraw self-delegation reward
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawValidatorRewardsAll(ctx, valOpAddr1)
	amt = accMapper.GetAccount(ctx, valAccAddr1).GetCoins().AmountOf(denom)
	expRes := sdk.NewDec(90).Add(sdk.NewDec(100).Quo(sdk.NewDec(2))).TruncateInt() // 90 + 100 tokens * 10/20
	require.True(sdk.IntEq(t, expRes, amt))
}

func TestWithdrawValidatorRewardsAllDelegatorWithCommission(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	commissionRate := sdk.NewDecWithPrec(1, 1)
	msgCreateValidator := stake.NewTestMsgCreateValidatorWithCommission(
		valOpAddr1, valConsPk1, 10, commissionRate)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// delegate
	msgDelegate := stake.NewTestMsgDelegate(delAddr1, valOpAddr1, 10)
	got = stakeHandler(ctx, msgDelegate)
	require.True(t, got.IsOK())
	amt := accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)
	require.Equal(t, int64(90), amt.Int64())

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.AllocateTokens(ctx, sdk.OneDec(), valConsAddr1)

	// withdraw validator reward
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawValidatorRewardsAll(ctx, valOpAddr1)
	amt = accMapper.GetAccount(ctx, valAccAddr1).GetCoins().AmountOf(denom)
	commissionTaken := sdk.NewDec(100).Mul(commissionRate)
	afterCommission := sdk.NewDec(100).Sub(commissionTaken)
	selfDelegationReward := afterCommission.Quo(sdk.NewDec(2))
	expRes := sdk.NewDec(90).Add(commissionTaken).Add(selfDelegationReward).TruncateInt() // 90 + 100 tokens * 10/20
	require.True(sdk.IntEq(t, expRes, amt))
}

func TestWithdrawValidatorRewardsAllMultipleValidator(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	// Make some  validators with different commissions.
	// Bond 10 of 100 with 0.1 commission.
	msgCreateValidator := stake.NewTestMsgCreateValidatorWithCommission(
		valOpAddr1, valConsPk1, 10, sdk.NewDecWithPrec(1, 1))
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// Bond 50 of 100 with 0.2 commission.
	msgCreateValidator = stake.NewTestMsgCreateValidatorWithCommission(
		valOpAddr2, valConsPk2, 50, sdk.NewDecWithPrec(2, 1))
	got = stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// Bond 40 of 100 with 0.3 commission.
	msgCreateValidator = stake.NewTestMsgCreateValidatorWithCommission(
		valOpAddr3, valConsPk3, 40, sdk.NewDecWithPrec(3, 1))
	got = stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// Allocate 1000 denom of fees.
	feeInputs := sdk.NewInt(1000)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	// Collect proposer reward for 100% of votes.
	keeper.AllocateTokens(ctx, sdk.OneDec(), valConsAddr1)

	// Withdraw validator reward.
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawValidatorRewardsAll(ctx, valOpAddr1)
	amt := accMapper.GetAccount(ctx, valAccAddr1).GetCoins().AmountOf(denom)

	feesInNonProposer := sdk.NewDecFromInt(feeInputs).Mul(sdk.NewDecWithPrec(95, 2))
	feesInProposer := sdk.NewDecFromInt(feeInputs).Mul(sdk.NewDecWithPrec(5, 2))
	// NOTE: the non-proposer rewards (95) and proposer rewards (50) add up to
	// 145.  During computation, this is further split into 130.5 and 14.5,
	// which is the non-commission and commission respectively, but the
	// commission is for self so the result is just 145.
	expRes := sdk.NewDec(90). // orig tokens (100) - bonded (10)
					Add(feesInNonProposer.Quo(sdk.NewDec(10))). // validator 1 has 1/10 total power (non-proposer rewards = 95)
					Add(feesInProposer).                        // (proposer rewards = 50)
					TruncateInt()
	require.True(sdk.IntEq(t, expRes, amt))
}

func TestWithdrawValidatorRewardsAllMultipleDelegator(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputAdvanced(t, false, 100, sdk.ZeroDec())
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator with 10% commission
	commissionRate := sdk.NewDecWithPrec(1, 1)
	msgCreateValidator := stake.NewTestMsgCreateValidatorWithCommission(
		valOpAddr1, valConsPk1, 10, sdk.NewDecWithPrec(1, 1))
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	_ = sk.ApplyAndReturnValidatorSetUpdates(ctx)

	// delegate
	msgDelegate := stake.NewTestMsgDelegate(delAddr1, valOpAddr1, 10)
	got = stakeHandler(ctx, msgDelegate)
	require.True(t, got.IsOK())
	amt := accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)
	require.Equal(t, int64(90), amt.Int64())

	msgDelegate = stake.NewTestMsgDelegate(delAddr2, valOpAddr1, 20)
	got = stakeHandler(ctx, msgDelegate)
	require.True(t, got.IsOK())
	amt = accMapper.GetAccount(ctx, delAddr2).GetCoins().AmountOf(denom)
	require.Equal(t, int64(80), amt.Int64())

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.AllocateTokens(ctx, sdk.OneDec(), valConsAddr1)

	// withdraw validator reward
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawValidatorRewardsAll(ctx, valOpAddr1)
	amt = accMapper.GetAccount(ctx, valAccAddr1).GetCoins().AmountOf(denom)

	commissionTaken := sdk.NewDec(100).Mul(commissionRate)
	afterCommission := sdk.NewDec(100).Sub(commissionTaken)
	expRes := sdk.NewDec(90).
		Add(afterCommission.Quo(sdk.NewDec(4))).
		Add(commissionTaken).
		TruncateInt() // 90 + 100*90% tokens * 10/40
	require.True(sdk.IntEq(t, expRes, amt))
}
