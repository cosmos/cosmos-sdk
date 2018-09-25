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

//_________________________________________________________________________________________

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ sdk.ValidatorHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) ValidatorHooks() Hooks {
	return Hooks{k}
}

// Implements sdk.ValidatorHooks
func (h Hooks) OnValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
	h.k.onValidatorBonded(ctx, address)
}

// Implements sdk.ValidatorHooks
func (h Hooks) OnValidatorBeginUnbonding(ctx sdk.Context, address sdk.ConsAddress) {
	h.k.onValidatorBeginUnbonding(ctx, address)
}
