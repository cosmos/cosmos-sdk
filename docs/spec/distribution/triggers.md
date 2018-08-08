# Triggers

## Create or modify delegation distribution
 
 - triggered-by: `stake.TxDelegate`, `stake.TxBeginRedelegate`, `stake.TxBeginUnbonding`

The pool of a new delegator bond will be 0 for the height at which the bond was
added, or the withdrawal has taken place. This is achieved by setting
`DelegatorDist.WithdrawalHeight` to the relevant height, withdrawing any
remaining fees, and setting `DelegatorDist.Accum` and
`DelegatorDist.ProposerAccum` to 0.

## Commission rate change
 
 - triggered-by: `stake.TxEditValidator`

If a validator changes its commission rate, all commission on fees must be
simultaneously withdrawn using the transaction `TxWithdrawValidator`

## Change in Validator State
 
 - triggered-by: `stake.Slash`, `stake.UpdateValidator`

Whenever a validator is slashed or enters/leaves the validator group
`ValidatorUpdate` information must be recorded in order to properly calculate
the accum factors. 
