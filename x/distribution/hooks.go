package distribution

import "github.com/cosmos/cosmos-sdk/x/distribution/types"

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

// Withdrawal all distubution rewards // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX XXX
func (k Keeper) onValidatorBondModified(ctx sdk.Context, addr sdk.ValAddress) {
	slashingPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: address,
		StartHeight:   ctx.BlockHeight(),
		EndHeight:     0,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	k.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onValidatorRemoved(ctx sdk.Context, addr sdk.ValAddress) {
	k.RemoveValidatorDistInfo(ctx, addr)
}

//_________________________________________________________________________________________

// Create a new validator distribution record
func (k Keeper) onDelegationCreated(ctx sdk.Context, address sdk.ConsAddress) {
	slashingPeriod := ValidatorSlashingPeriod{
		ValidatorAddr: address,
		StartHeight:   ctx.BlockHeight(),
		EndHeight:     0,
		SlashedSoFar:  sdk.ZeroDec(),
	}
	k.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)
}

//_________________________________________________________________________________________

// Wrapper struct for sdk.ValidatorHooks
type ValidatorHooks struct {
	k Keeper
}

var _ sdk.ValidatorHooks = ValidatorHooks{}

// nolint
func (k Keeper) ValidatorHooks() sdk.ValidatorHooks { return ValidatorHooks{k} }
func (v ValidatorHooks) OnValidatorCreated(ctx sdk.Context, addr sdk.VlAddress) {
	v.k.OnValidatorCreated(ctx, address)
}
func (v ValidatorHooks) OnValidatorBondModified(ctx sdk.Context, addr sdk.ValAddress) {
	v.k.OnValidatorBondModified(ctx, address)
}
func (v ValidatorHooks) OnValidatorRemoved(ctx sdk.Context, addr sdk.ValAddress) {
	v.k.OnValidatorRemoved(ctx, address)
}
func (v ValidatorHooks) OnValidatorBonded(_ sdk.Context, _ sdk.ConsAddress)      {}
func (v ValidatorHooks) OnValidatorBeginBonded(_ sdk.Context, _ sdk.ConsAddress) {}

//_________________________________________________________________________________________

// Wrapper struct for sdk.DelegationHooks
type DelegationHooks struct {
	k Keeper
}

var _ sdk.DelegationHooks = DelegationHooks{}

// nolint
func (k Keeper) DelegationHooks() sdk.DelegationHooks { return DelegationHooks{k} }
func (d DelegationHooks) OnDelegatoinCreated(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	d.k.OnDelegatoinCreated(ctx, address)
}
func (d DelegationHooks) OnDelegationSharesModified(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	d.k.OnDelegationSharesModified(ctx, address)
}
func (d DelegationHooks) OnDelegationRemoved(ctx sdk.Context,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	d.k.OnDelegationRemoved(ctx, address)
}
