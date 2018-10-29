package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Create a new validator distribution record
func (k Keeper) onValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {

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
func (k Keeper) onValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	// This doesn't need to be run at genesis
	if ctx.BlockHeight() > 0 {
		if err := k.WithdrawValidatorRewardsAll(ctx, valAddr); err != nil {
			panic(err)
		}
	}
}

// Withdraw all validator rewards
func (k Keeper) onValidatorBonded(ctx sdk.Context, valAddr sdk.ValAddress) {
	lastPower := k.stakeKeeper.GetLastValidatorPower(ctx, valAddr)
	if !lastPower.Equal(sdk.ZeroInt()) {
		panic("expected last power to be 0 for validator entering bonded state")
	}
	k.onValidatorModified(ctx, valAddr)
}

// XXX Consider removing this after debugging.
func (k Keeper) onValidatorPowerDidChange(ctx sdk.Context, valAddr sdk.ValAddress) {
	vi := k.GetValidatorDistInfo(ctx, valAddr)
	if vi.FeePoolWithdrawalHeight != ctx.BlockHeight() {
		panic("expected validator dist info FeePoolWithdrawalHeight to be updated, but was not.")
	}
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onValidatorRemoved(ctx sdk.Context, valAddr sdk.ValAddress) {
	k.RemoveValidatorDistInfo(ctx, valAddr)
}

//_________________________________________________________________________________________

// Create a new delegator distribution record
func (k Keeper) onDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	ddi := types.DelegationDistInfo{
		DelegatorAddr:           delAddr,
		ValOperatorAddr:         valAddr,
		DelPoolWithdrawalHeight: ctx.BlockHeight(),
	}
	k.SetDelegationDistInfo(ctx, ddi)
}

// Withdrawal all validator rewards
func (k Keeper) onDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	if err := k.WithdrawDelegationReward(ctx, delAddr, valAddr); err != nil {
		panic(err)
	}
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

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
func (h Hooks) OnValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.onValidatorCreated(ctx, valAddr)
}
func (h Hooks) OnValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.onValidatorModified(ctx, valAddr)
}
func (h Hooks) OnValidatorRemoved(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.onValidatorRemoved(ctx, valAddr)
}
func (h Hooks) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onValidatorModified(ctx, valAddr)
	h.k.onDelegationCreated(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onValidatorModified(ctx, valAddr)
	h.k.onDelegationSharesModified(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationRemoved(ctx, delAddr, valAddr)
}
func (h Hooks) OnValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.onValidatorModified(ctx, valAddr)
}
func (h Hooks) OnValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.onValidatorBonded(ctx, valAddr)
}
func (h Hooks) OnValidatorPowerDidChange(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.k.onValidatorPowerDidChange(ctx, valAddr)
}
