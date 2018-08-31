package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create a new slashing period when a validator is bonded
func (k Keeper) onValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
	slashingPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: address,
		StartHeight:   ctx.BlockHeight(),
		EndHeight:     0,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	k.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
}

// Mark the slashing period as having ended when a validator begins unbonding
func (k Keeper) onValidatorBeginUnbonding(ctx sdk.Context, address sdk.ConsAddress) {
	slashingPeriod := k.getValidatorSlashingPeriodForHeight(ctx, address, ctx.BlockHeight())
	slashingPeriod.EndHeight = ctx.BlockHeight()
	k.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
}

// Wrapper struct for sdk.ValidatorHooks
type ValidatorHooks struct {
	k Keeper
}

// Assert implementation
var _ sdk.ValidatorHooks = ValidatorHooks{}

// Return a sdk.ValidatorHooks interface over the wrapper struct
func (k Keeper) ValidatorHooks() sdk.ValidatorHooks {
	return ValidatorHooks{k}
}

// Implements sdk.ValidatorHooks
func (v ValidatorHooks) OnValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
	v.k.onValidatorBonded(ctx, address)
}

// Implements sdk.ValidatorHooks
func (v ValidatorHooks) OnValidatorBeginUnbonding(ctx sdk.Context, address sdk.ConsAddress) {
	v.k.onValidatorBeginUnbonding(ctx, address)
}
