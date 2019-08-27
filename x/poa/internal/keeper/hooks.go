package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

// Implements StakingHooks interface
var _ types.POAHooks = Keeper{}

// AfterValidatorCreated - call hook if registered
func (k Keeper) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.AfterValidatorCreated(ctx, valAddr)
	}
}

// BeforeValidatorModified - call hook if registered
func (k Keeper) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.BeforeValidatorModified(ctx, valAddr)
	}
}

// AfterValidatorRemoved - call hook if registered
func (k Keeper) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.AfterValidatorRemoved(ctx, consAddr, valAddr)
	}
}

// AfterValidatorBonded - call hook if registered
func (k Keeper) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.AfterValidatorBonded(ctx, consAddr, valAddr)
	}
}

// AfterValidatorBeginUnbonding - call hook if registered
func (k Keeper) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	if k.hooks != nil {
		k.hooks.AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
	}
}

// BeforeValidatorSlashed - call hook if registered
func (k Keeper) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	if k.hooks != nil {
		k.hooks.BeforeValidatorSlashed(ctx, valAddr, fraction)
	}
}
