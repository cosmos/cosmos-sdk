package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Create a new validator distribution record
func (k Keeper) onValidatorCreated(ctx sdk.Context, addr sdk.ValAddress) {

	height := ctx.BlockHeight()
	vdi := types.ValidatorDistInfo{
		OperatorAddr:           addr,
		GlobalWithdrawalHeight: height,
		Pool:           DecCoins{},
		PoolCommission: DecCoins{},
		DelAccum:       NewTotalAccum(height),
	}
	k.SetValidatorDistInfo(ctx, vdi)
}

// Withdrawal all validator rewards
func (k Keeper) onValidatorCommissionChange(ctx sdk.Context, addr sdk.ValAddress) {
	k.WithdrawValidatorRewardsAll(ctx, addr)
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onValidatorRemoved(ctx sdk.Context, addr sdk.ValAddress) {
	k.RemoveValidatorDistInfo(ctx, addr)
}

//_________________________________________________________________________________________

// Create a new delegator distribution record
func (k Keeper) onDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	ddi := types.DelegatorDistInfo{
		DelegatorAddr:    delAddr,
		ValOperatorAddr:  valAddr,
		WithdrawalHeight: ctx.BlockHeight(),
	}
	k.SetDelegatorDistInfo(ctx, ddi)
}

// Withdrawal all validator rewards
func (k Keeper) onDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	k.WithdrawDelegationReward(ctx, delAddr, valAddr)
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	k.RemoveDelegatorDistInfo(ctx, delAddr, valAddr)
}

//_________________________________________________________________________________________

// Wrapper struct
type Hooks struct {
	k Keeper
}

// New Validator Hooks
func (k Keeper) ValidatorHooks() Hooks { return Hooks{k} }

// nolint
func (h Hooks) OnValidatorCreated(ctx sdk.Context, addr sdk.VlAddress) {
	v.k.onValidatorCreated(ctx, address)
}
func (h Hooks) OnValidatorCommissionChange(ctx sdk.Context, addr sdk.ValAddress) {
	v.k.onValidatorCommissionChange(ctx, address)
}
func (h Hooks) OnValidatorRemoved(ctx sdk.Context, addr sdk.ValAddress) {
	v.k.onValidatorRemoved(ctx, address)
}
func (h Hooks) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationCreated(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	d.k.onDelegationSharesModified(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationRemoved(ctx, delAddr, valAddr)
}
