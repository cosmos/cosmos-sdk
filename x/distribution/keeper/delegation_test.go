package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestWithdrawDelegationRewardBasic(t *testing.T) {
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

	totalPower := int64(20)
	totalPowerDec := sdk.NewDec(totalPower)

	// allocate 100 denom of fees
	feeInputs := sdk.NewInt(100)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))
	keeper.SetProposerConsAddr(ctx, valConsAddr1)
	keeper.SetSumPrecommitPower(ctx, totalPowerDec)
	keeper.AllocateFees(ctx)

	// withdraw delegation
	ctx = ctx.WithBlockHeight(1)
	keeper.WithdrawDelegationReward(ctx, delAddr1, valOpAddr1)
	amt = accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)

	expRes := sdk.NewDec(90).Add(sdk.NewDec(100).Quo(sdk.NewDec(2))) // 90 + 100 tokens * 10/20
	require.True(sdk.DecEq(t, expRes, sdk.NewDecFromInt(amt)))
}

func TestWithdrawDelegationRewardWithCommission(t *testing.T) {

}

func TestWithdrawDelegationRewardsAll(t *testing.T) {

}

func TestGetDelegatorRewardsAll(t *testing.T) {

}
