//nolint
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Expose the hooks if present
func (k Keeper) OnValidatorCreated(ctx sdk.Context, address sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorCreated(ctx, address)
	}
}
func (k Keeper) OnValidatorCommissionChange(ctx sdk.Context, address sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorCommissionChange(ctx, address)
	}
}

func (k Keeper) OnValidatorRemoved(ctx sdk.Context, address sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorRemoved(ctx, address)
	}
}

func (k Keeper) OnValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorBonded(ctx, address)
	}
}

func (k Keeper) OnValidatorBeginUnbonding(ctx sdk.Context, address sdk.ConsAddress, operator sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorBeginUnbonding(ctx, address, operator)
	}
}

func (k Keeper) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnDelegationCreated(ctx, delAddr, valAddr)
	}
}

func (k Keeper) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnDelegationSharesModified(ctx, delAddr, valAddr)
	}
}

func (k Keeper) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnDelegationRemoved(ctx, delAddr, valAddr)
	}
}
