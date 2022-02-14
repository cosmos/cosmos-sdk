package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple staking hooks, all hook functions are run in array sequence
type MultiBankHooks []BankHooks

func NewMultiStakingHooks(hooks ...BankHooks) MultiBankHooks {
	return hooks
}

func (h MultiBankHooks) BeforeSend(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		err := h[i].BeforeSend(ctx, from, to, amount)
		if err != nil {
			return err
		}
	}
	return nil
}
