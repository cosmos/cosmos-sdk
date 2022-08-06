package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Implements StakingHooks interface
var _ types.BankHooks = BaseSendKeeper{}

// BeforeSend executes the BeforeSend hook if registered.
func (k BaseSendKeeper) BeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	if k.hooks != nil {
		return k.hooks.BeforeSend(ctx, from, to, amount)
	}
	return nil
}
