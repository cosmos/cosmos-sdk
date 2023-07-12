package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StakingHooksTemplate struct{}

var _ StakingHooks = StakingHooksTemplate{}

func (h StakingHooksTemplate) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h StakingHooksTemplate) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
}
func (h StakingHooksTemplate) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	return nil
}
func (h StakingHooksTemplate) BeforeTokenizeShareRecordRemoved(ctx sdk.Context, recordId uint64) error {
	return nil
}
