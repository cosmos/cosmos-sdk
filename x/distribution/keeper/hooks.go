package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Create a new validator distribution record
func (k Keeper) onValidatorCreated(ctx sdk.Context, addr sdk.ValAddress) {

	height := ctx.BlockHeight()
	vdi := types.ValidatorDistInfo{
		OperatorAddr:            addr,
		FeePoolWithdrawalHeight: height,
		Pool:           types.DecCoins{},
		PoolCommission: types.DecCoins{},
		DelAccum:       types.NewTotalAccum(height),
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

	ddi := types.DelegationDistInfo{
		DelegatorAddr:    delAddr,
		ValOperatorAddr:  valAddr,
		WithdrawalHeight: ctx.BlockHeight(),
	}
	k.SetDelegationDistInfo(ctx, ddi)
	ctx.Logger().With("module", "x/distribution").Error(fmt.Sprintf("ddi created: %v", ddi))
}

// Withdrawal all validator rewards
func (k Keeper) onDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {

	k.WithdrawDelegationReward(ctx, delAddr, valAddr)
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
func (h Hooks) OnValidatorCreated(ctx sdk.Context, addr sdk.ValAddress) {
	h.k.onValidatorCreated(ctx, addr)
}
func (h Hooks) OnValidatorCommissionChange(ctx sdk.Context, addr sdk.ValAddress) {
	h.k.onValidatorCommissionChange(ctx, addr)
}
func (h Hooks) OnValidatorRemoved(ctx sdk.Context, addr sdk.ValAddress) {
	h.k.onValidatorRemoved(ctx, addr)
}
func (h Hooks) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationCreated(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationSharesModified(ctx, delAddr, valAddr)
}
func (h Hooks) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.onDelegationRemoved(ctx, delAddr, valAddr)
}

// nolint - unused hooks for interface
func (h Hooks) OnValidatorBonded(ctx sdk.Context, addr sdk.ConsAddress)         {}
func (h Hooks) OnValidatorBeginUnbonding(ctx sdk.Context, addr sdk.ConsAddress) {}
