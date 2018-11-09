//nolint
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Expose the hooks if present
func (k Keeper) OnValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorCreated(ctx, valAddr)
	}
}
func (k Keeper) OnValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorModified(ctx, valAddr)
	}
}

func (k Keeper) OnValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorRemoved(ctx, consAddr, valAddr)
	}
}

func (k Keeper) OnValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorBonded(ctx, consAddr, valAddr)
	}
}

func (k Keeper) OnValidatorPowerDidChange(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorPowerDidChange(ctx, consAddr, valAddr)
	}
}

func (k Keeper) OnValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.OnValidatorBeginUnbonding(ctx, consAddr, valAddr)
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
