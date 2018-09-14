package stake

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
func (k Keeper) onValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
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
func (v ValidatorHooks) OnValidatorBonded(ctx sdk.Context, address sdk.ConsAddress) {
	v.k.onValidatorBonded(ctx, address)
}
