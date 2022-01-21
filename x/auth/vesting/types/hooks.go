package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type distributionHooks struct {
	accountKeeper AccountKeeper
	bankKeeper    BankKeeper
	stakingKeeper StakingKeeper
}

var _ DistributionHooks = distributionHooks{}

func NewDistributionHooks(ak AccountKeeper, bk BankKeeper, sk StakingKeeper) DistributionHooks {
	return distributionHooks{
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

func (dh distributionHooks) AllowWithdrawAddr(ctx sdk.Context, delAddr sdk.AccAddress) bool {
	acc := dh.accountKeeper.GetAccount(ctx, delAddr)
	_, isClawback := acc.(*TrueVestingAccount)
	return !isClawback
}

func (dh distributionHooks) AfterDelegationReward(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress, reward sdk.Coins) {
	acc := dh.accountKeeper.GetAccount(ctx, delAddr)
	cva, isClawback := acc.(*TrueVestingAccount)
	if isClawback {
		cva.PostReward(ctx, reward, dh.accountKeeper, dh.bankKeeper, dh.stakingKeeper)
	}
}
