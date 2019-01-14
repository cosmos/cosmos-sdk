package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ sdk.StakingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// nolint
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	h.k.initializeValidator(ctx, val)
}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	// increment period
	h.k.incrementValidatorPeriod(ctx, val)
}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)

	// increment period
	h.k.incrementValidatorPeriod(ctx, val)
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)

	// withdraw delegation rewards (which also increments period)
	if err := h.k.withdrawDelegationRewards(ctx, val, del); err != nil {
		panic(err)
	}
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// nothing needed here since OnDelegationSharesModified will always also be called
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// create new delegation period record
	h.k.initializeDelegation(ctx, valAddr, delAddr)
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorPowerDidChange(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// record the slash event
	h.k.updateValidatorSlashFraction(ctx, valAddr, fraction)
}
