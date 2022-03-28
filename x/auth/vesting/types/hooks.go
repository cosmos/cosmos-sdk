package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

type distributionHooks struct {
	accountKeeper AccountKeeper
	rewardAction  exported.RewardAction
}

var _ DistributionHooks = distributionHooks{}

func NewDistributionHooks(ak AccountKeeper, bk BankKeeper, sk StakingKeeper) DistributionHooks {
	return distributionHooks{
		accountKeeper: ak,
		rewardAction:  NewClawbackRewardAction(ak, bk, sk),
	}
}

func (dh distributionHooks) AllowWithdrawAddr(ctx sdk.Context, delAddr sdk.AccAddress) bool {
	acc := dh.accountKeeper.GetAccount(ctx, delAddr)
	_, isClawback := acc.(exported.ClawbackVestingAccountI)
	return !isClawback
}

func (dh distributionHooks) AfterDelegationReward(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress, reward sdk.Coins) {
	acc := dh.accountKeeper.GetAccount(ctx, delAddr)
	cva, isClawback := acc.(exported.ClawbackVestingAccountI)
	if isClawback {
		err := cva.PostReward(ctx, reward, dh.rewardAction)
		if err != nil {
			panic(err)
		}
	}
}
