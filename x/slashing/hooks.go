package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create a new slashing period when a validator is bonded
func (k Keeper) onValidatorBonded(ctx sdk.Context, address sdk.ValAddress) {
	slashingPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: address,
		StartHeight:   ctx.BlockHeight(),
		EndHeight:     0,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	k.setValidatorSlashingPeriod(ctx, slashingPeriod)
}

// Mark the slashing period as having ended when a validator is unbonded
func (k Keeper) onValidatorUnbonded(ctx sdk.Context, address sdk.ValAddress) {
	slashingPeriod := k.getValidatorSlashingPeriodForHeight(ctx, address, ctx.BlockHeight())
	slashingPeriod.EndHeight = ctx.BlockHeight()
	k.setValidatorSlashingPeriod(ctx, slashingPeriod)
}

// Wrapper struct for sdk.ValidatorHooks
type ValidatorHooks struct {
	k Keeper
}

// Return a sdk.ValidatorHooks interface over the wrapper struct
func (k Keeper) ValidatorHooks() sdk.ValidatorHooks {
	return ValidatorHooks{k}
}

// Implements sdk.ValidatorHooks
func (v ValidatorHooks) OnValidatorBonded(ctx sdk.Context, address sdk.ValAddress) {
	v.k.onValidatorBonded(ctx, address)
}

// Implements sdk.ValidatorHooks
func (v ValidatorHooks) OnValidatorUnbonded(ctx sdk.Context, address sdk.ValAddress) {
	v.k.onValidatorUnbonded(ctx, address)
}

// Assert implementation
var _ sdk.ValidatorHooks = ValidatorHooks{}
