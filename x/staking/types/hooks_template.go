package types

import (
	"time"

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
func (h StakingHooksTemplate) UnbondingDelegationEntryCreated(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress, creationHeight int64, completionTime time.Time, balance sdk.Int, id uint64) {
}
func (h StakingHooksTemplate) BeforeUnbondingDelegationEntryComplete(ctx sdk.Context, id uint64) bool {
	return false
}
