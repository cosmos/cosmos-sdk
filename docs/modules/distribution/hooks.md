# Hooks

## Create or modify delegation distribution
 
 - triggered-by: `stake.TxDelegate`, `stake.TxBeginRedelegate`, `stake.TxBeginUnbonding`

The pool of a new delegator bond will be 0 for the height at which the bond was
added, or the withdrawal has taken place. This is achieved by setting
`DelegationDistInfo.WithdrawalHeight` to the height of the triggering transaction. 

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
