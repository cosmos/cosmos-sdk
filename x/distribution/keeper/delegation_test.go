package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
)

func TestWithdrawDelegationReward(t *testing.T) {
	ctx, accMapper, keeper, sk := CreateTestInput(t, false, 100)
	stakeHandler := stake.NewHandler(sk)
	denom := sk.GetParams(ctx).BondDenom

	//first make a validator
	msgCreateValidator := stake.NewTestMsgCreateValidator(valAddr1, valPk1, 10)
	got := stakeHandler(ctx, msgCreateValidator)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// delegate
	msgDelegate := stake.NewTestMsgDelegate(delAddr1, valAddr1, 10)
	got = stakeHandler(ctx, msgDelegate)
	require.True(t, got.IsOK())

	amt := accMapper.GetAccount(ctx, delAddr1).GetCoins().AmountOf(denom)
	require.Equal(t, int64(90), amt.Int64())

	keeper.WithdrawDelegationReward(ctx, delAddr1, valAddr1)
}

func TestWithdrawDelegationRewardsAll(t *testing.T) {

}

func TestGetDelegatorRewardsAll(t *testing.T) {

}
