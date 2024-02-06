package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Implements StakingHooks interface
var _ types.StakingHooks = Keeper{}

// AfterValidatorCreated - call hook if registered
func (k Keeper) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterValidatorCreated(ctx, valAddr)
	}
	return nil
}

// BeforeValidatorModified - call hook if registered
func (k Keeper) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.BeforeValidatorModified(ctx, valAddr)
	}
	return nil
}

// AfterValidatorRemoved - call hook if registered
func (k Keeper) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterValidatorRemoved(ctx, consAddr, valAddr)
	}
	return nil
}

// AfterValidatorBonded - call hook if registered
func (k Keeper) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterValidatorBonded(ctx, consAddr, valAddr)
	}
	return nil
}

// AfterValidatorBeginUnbonding - call hook if registered
func (k Keeper) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
	}
	return nil
}

// BeforeDelegationCreated - call hook if registered
func (k Keeper) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.BeforeDelegationCreated(ctx, delAddr, valAddr)
	}
	return nil
}

// BeforeDelegationSharesModified - call hook if registered
func (k Keeper) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.BeforeDelegationSharesModified(ctx, delAddr, valAddr)
	}
	return nil
}

// BeforeDelegationRemoved - call hook if registered
func (k Keeper) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		k.hooks.BeforeDelegationRemoved(ctx, delAddr, valAddr)
	}
	return nil
}

// AfterDelegationModified - call hook if registered
func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if k.hooks != nil {
		return k.hooks.AfterDelegationModified(ctx, delAddr, valAddr)
	}
	return nil
}

// BeforeValidatorSlashed - call hook if registered
func (k Keeper) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	if k.hooks != nil {
		return k.hooks.BeforeValidatorSlashed(ctx, valAddr, fraction)
	}
	return nil
}
