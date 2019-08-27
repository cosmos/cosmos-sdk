package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple poa hooks, all hook functions are run in array sequence
type MultiPOAHooks []POAHooks

func NewMultiPOAHooks(hooks ...POAHooks) MultiPOAHooks {
	return hooks
}

func (h MultiPOAHooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorCreated(ctx, valAddr)
	}
}
func (h MultiPOAHooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].BeforeValidatorModified(ctx, valAddr)
	}
}
func (h MultiPOAHooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorRemoved(ctx, consAddr, valAddr)
	}
}
func (h MultiPOAHooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorBonded(ctx, consAddr, valAddr)
	}
}
func (h MultiPOAHooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	for i := range h {
		h[i].AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
	}
}
func (h MultiPOAHooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	for i := range h {
		h[i].BeforeValidatorSlashed(ctx, valAddr, fraction)
	}
}
