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
