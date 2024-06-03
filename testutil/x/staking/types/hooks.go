package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple staking hooks, all hook functions are run in array sequence
var _ StakingHooks = &MultiStakingHooks{}

type MultiStakingHooks []StakingHooks

func NewMultiStakingHooks(hooks ...StakingHooks) MultiStakingHooks {
	return hooks
}

func (h MultiStakingHooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].AfterValidatorCreated(ctx, valAddr); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiStakingHooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].BeforeDelegationSharesModified(ctx, delAddr, valAddr); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	for i := range h {
		if err := h[i].AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			return err
		}
	}
	return nil
}
