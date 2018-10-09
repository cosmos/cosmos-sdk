package keeper

import "testing"

func TestWithdrawDelegationReward(t *testing.T) {

	keeper.WithdrawDelegationReward(ctx, delAddr1, valAddr1)
}

func TestWithdrawDelegationRewardsAll(t *testing.T) {

}

func TestGetDelegatorRewardsAll(t *testing.T) {

}
