package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingHooksTemplate struct{}

var _ types.StakingHooks = StakingHooksTemplate{}

func (h StakingHooksTemplate) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	return nil
}
func (h StakingHooksTemplate) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationRemoved(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) UnbondingDelegationEntryCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress, _ int64, _ time.Time, _ sdk.Int, id uint64) {
}
func (h StakingHooksTemplate) BeforeUnbondingDelegationEntryComplete(_ sdk.Context, _ uint64) bool {
	return false
}
