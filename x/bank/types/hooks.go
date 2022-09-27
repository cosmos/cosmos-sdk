package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MultiBankHooks combine multiple bank hooks, all hook functions are run in array sequence
type MultiBankHooks []BankHooks

// NewMultiBankHooks takes a list of BankHooks and returns a MultiBankHooks
func NewMultiBankHooks(hooks ...BankHooks) MultiBankHooks {
	return hooks
}

// BeforeSend runs the BeforeSend hooks in order for each BankHook in a MultiBankHooks struct
func (h MultiBankHooks) BeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		err := h[i].BeforeSend(ctx, from, to, amount)
		if err != nil {
			return err
		}
	}
	return nil
}
