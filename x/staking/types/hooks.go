package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple staking hooks, all hook functions are run in array sequence
type MultiStakingHooks []StakingHooks

func NewMultiStakingHooks(hooks ...StakingHooks) MultiStakingHooks {
	return hooks
}

func (h MultiStakingHooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorCreated(ctx, valAddr)
	}
}
func (h MultiStakingHooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].BeforeValidatorModified(ctx, valAddr)
	}
}
func (h MultiStakingHooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorRemoved(ctx, consAddr, valAddr)
	}
}
func (h MultiStakingHooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorBonded(ctx, consAddr, valAddr)
	}
}
func (h MultiStakingHooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
	}
}
func (h MultiStakingHooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].BeforeDelegationCreated(ctx, delAddr, valAddr)
	}
}
func (h MultiStakingHooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].BeforeDelegationSharesModified(ctx, delAddr, valAddr)
	}
}
func (h MultiStakingHooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].BeforeDelegationRemoved(ctx, delAddr, valAddr)
	}
}
func (h MultiStakingHooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterDelegationModified(ctx, delAddr, valAddr)
	}
}
func (h MultiStakingHooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	for i := range h {
		h[i].BeforeValidatorSlashed(ctx, valAddr, fraction)
	}
}
