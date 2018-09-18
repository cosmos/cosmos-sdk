package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

/*
## Create or modify delegation distribution

 - triggered-by: `stake.TxDelegate`, `stake.TxBeginRedelegate`, `stake.TxBeginUnbonding`

The pool of a new delegator bond will be 0 for the height at which the bond was
added, or the withdrawal has taken place. This is achieved by setting
`DelegatorDistInfo.WithdrawalHeight` to the height of the triggering transaction.

## Commission rate change

 - triggered-by: `stake.TxEditValidator`

If a validator changes its commission rate, all commission on fees must be
simultaneously withdrawn using the transaction `TxWithdrawValidator`.
Additionally the change and associated height must be recorded in a
`ValidatorUpdate` state record.

## Change in Validator State

 - triggered-by: `stake.Slash`, `stake.UpdateValidator`

Whenever a validator is slashed or enters/leaves the validator group all of the
validator entitled reward tokens must be simultaneously withdrawn from
`Global.Pool` and added to `ValidatorDistInfo.Pool`.
*/

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

// nolint
func (k Keeper) ValidatorHooks() sdk.ValidatorHooks { return ValidatorHooks{k} }
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
