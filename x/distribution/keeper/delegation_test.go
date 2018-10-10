package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestWithdrawDelegationReward(t *testing.T) {
	ctx, accMapper, keeper, sk, fck := CreateTestInputDefault(t, false, 100)
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

	feeInputs := sdk.NewInt(20)
	fck.SetCollectedFees(sdk.Coins{sdk.NewCoin(denom, feeInputs)})
	require.Equal(t, feeInputs, fck.GetCollectedFees(ctx).AmountOf(denom))

	keeper.WithdrawDelegationReward(ctx, delAddr1, valOpAddr1)

	amt = accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)
	require.Equal(t, int64(100), amt.Int64())
}

func TestWithdrawDelegationRewardsAll(t *testing.T) {

}

func TestGetDelegatorRewardsAll(t *testing.T) {

}
