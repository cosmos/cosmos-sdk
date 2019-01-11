package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Create a new validator distribution record
func (k Keeper) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {

	// defensive check for existence
	if k.HasValidatorDistInfo(ctx, valAddr) {
		panic("validator dist info already exists (not cleaned up properly)")
	}

	height := ctx.BlockHeight()
	vdi := types.ValidatorDistInfo{
		OperatorAddr:            valAddr,
		FeePoolWithdrawalHeight: height,
		DelAccum:                types.NewTotalAccum(height),
		DelPool:                 types.DecCoins{},
		ValCommission:           types.DecCoins{},
	}
	k.SetValidatorDistInfo(ctx, vdi)
}

// Withdraw all validator rewards
func (k Keeper) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	// Move the validator's rewards from the global pool to the validator's pools
	// (dist info), but without actually withdrawing the rewards. This does not
	// need to happen during the genesis block.
	if ctx.BlockHeight() > 0 {
		if err := k.takeValidatorFeePoolRewards(ctx, valAddr); err != nil {
			panic(err)
		}
	}
}

// Withdraw all validator rewards
func (k Keeper) AfterValidatorBonded(ctx sdk.Context, valAddr sdk.ValAddress) {
	lastPower := k.stakingKeeper.GetLastValidatorPower(ctx, valAddr)
	if !lastPower.Equal(sdk.ZeroInt()) {
		panic("expected last power to be 0 for validator entering bonded state")
	}
	k.BeforeValidatorModified(ctx, valAddr)
}

// Sanity check, very useful!
func (k Keeper) AfterValidatorPowerDidChange(ctx sdk.Context, valAddr sdk.ValAddress) {
	vi := k.GetValidatorDistInfo(ctx, valAddr)
	if vi.FeePoolWithdrawalHeight != ctx.BlockHeight() {
		panic(fmt.Sprintf("expected validator (%v) dist info FeePoolWithdrawalHeight to be updated to %v, but was %v.",
			valAddr.String(), ctx.BlockHeight(), vi.FeePoolWithdrawalHeight))
	}
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) AfterValidatorRemoved(ctx sdk.Context, valAddr sdk.ValAddress) {
	k.RemoveValidatorDistInfo(ctx, valAddr)
}

//_________________________________________________________________________________________

// Create a new delegator distribution record
func (k Keeper) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	ddi := types.DelegationDistInfo{
		DelegatorAddr:           delAddr,
		ValOperatorAddr:         valAddr,
		DelPoolWithdrawalHeight: ctx.BlockHeight(),
	}
	k.SetDelegationDistInfo(ctx, ddi)
}

// Withdrawal all validator rewards
func (k Keeper) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	if err := k.WithdrawDelegationReward(ctx, delAddr, valAddr); err != nil {
		panic(err)
	}
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {
	// Withdraw validator commission when validator self-bond is removed.
	// Because we maintain the invariant that all delegations must be removed
	// before a validator is deleted, this ensures that commission will be withdrawn
	// before the validator is deleted (and the corresponding ValidatorDistInfo removed).
	// If we change other parts of the code such that a self-delegation might remain after
	// a validator is deleted, this logic will no longer be safe.
	// TODO: Consider instead implementing this in a "BeforeValidatorRemoved" hook.
	if valAddr.Equals(sdk.ValAddress(delAddr)) {
		feePool, commission := k.withdrawValidatorCommission(ctx, valAddr)
		k.WithdrawToDelegator(ctx, feePool, delAddr, commission)
	}

	k.RemoveDelegationDistInfo(ctx, delAddr, valAddr)
}

//_________________________________________________________________________________________

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ sdk.StakingHooks = Hooks{}

// New Validator Hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// nolint
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.AfterValidatorCreated(ctx, valAddr)
}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.BeforeValidatorModified(ctx, valAddr)
}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.AfterValidatorRemoved(ctx, valAddr)
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.BeforeValidatorModified(ctx, valAddr)
	h.k.BeforeDelegationCreated(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.BeforeValidatorModified(ctx, valAddr)
	h.k.BeforeDelegationSharesModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.BeforeDelegationRemoved(ctx, delAddr, valAddr)
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.BeforeValidatorModified(ctx, valAddr)
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.AfterValidatorBonded(ctx, valAddr)
}
func (h Hooks) AfterValidatorPowerDidChange(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.AfterValidatorPowerDidChange(ctx, valAddr)
}
