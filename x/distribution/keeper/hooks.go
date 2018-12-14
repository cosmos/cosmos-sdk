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
func (h Hooks) OnValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	val := h.k.stakeKeeper.Validator(ctx, valAddr)
	h.k.initializeValidator(ctx, val)
}
func (h Hooks) OnValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakeKeeper.Validator(ctx, valAddr)

	// increment period
	h.k.incrementValidatorPeriod(ctx, val)

	// create new delegation period record
	h.k.initializeDelegation(ctx, valAddr, delAddr)
}
func (h Hooks) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakeKeeper.Validator(ctx, valAddr)
	del := h.k.stakeKeeper.Delegation(ctx, delAddr, valAddr)

	// withdraw delegation rewards (which also increments period)
	if err := h.k.withdrawDelegationRewards(ctx, val, del); err != nil {
		panic(err)
	}

	// create new delegation period record
	h.k.initializeDelegation(ctx, valAddr, delAddr)
}
func (h Hooks) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakeKeeper.Validator(ctx, valAddr)
	del := h.k.stakeKeeper.Delegation(ctx, delAddr, valAddr)

	// withdraw old delegation record (which also increments period)
	if err := h.k.withdrawDelegationRewards(ctx, val, del); err != nil {
		panic(err)
	}
}
func (h Hooks) OnValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorPowerDidChange(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
