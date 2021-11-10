package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type MockStakingHooks struct {
	beforeUnbondingDelegationEntryComplete func() bool
	unbondingDelegationEntryCreated        func(uint64)
}

var _ types.StakingHooks = MockStakingHooks{}

func (h MockStakingHooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	return nil
}
func (h MockStakingHooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error { return nil }
func (h MockStakingHooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}
func (h MockStakingHooks) BeforeDelegationRemoved(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h MockStakingHooks) UnbondingDelegationEntryCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress, _ int64, _ time.Time, _ sdk.Int, id uint64) {
	h.unbondingDelegationEntryCreated(id)
}

func (h MockStakingHooks) BeforeUnbondingDelegationEntryComplete(_ sdk.Context, _ uint64) bool {
	return h.beforeUnbondingDelegationEntryComplete()
}
