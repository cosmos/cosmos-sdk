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

// TrackBeforeSend runs the TrackBeforeSend hooks in order for each BankHook in a MultiBankHooks struct
func (h MultiBankHooks) TrackBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	for i := range h {
		h[i].TrackBeforeSend(ctx, from, to, amount)
	}
}

// BlockBeforeSend runs the BlockBeforeSend hooks in order for each BankHook in a MultiBankHooks struct
func (h MultiBankHooks) BlockBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		err := h[i].BlockBeforeSend(ctx, from, to, amount)
		if err != nil {
			return err
		}
	}
	return nil
}
