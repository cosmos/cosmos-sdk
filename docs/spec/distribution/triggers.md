# Triggers

## Create validator distribution

 - triggered-by: validator entering bonded validator group (`stake.bondValidator()`)

Whenever a new validator is added to the Tendermint validator set they are
entitled to begin earning rewards of atom provisions and fees. At this point
`ValidatorDistribution.Pool()` must be zero (as the validator has not yet
earned any rewards) meaning that the initial value for `validator.Adjustment`
must be set to the value of `validator.SimplePool()` for the height which the
validator is added on the validator set. 

## Create or modify delegation distribution
 
 - triggered-by: `stake.TxDelegate`, `stake.TxBeginRedelegate`, `stake.TxBeginUnbonding`

The pool of a new delegator bond will be 0 for the height at which the bond was
added. This is achieved by setting `DelegationDistribution.WithdrawalHeight` to
the height which the bond was added. Additionally the `AdjustmentPool` and
`AdjustmentProposerPool` must be set to the equivalent values of
`DelegationDistribution.SimplePool()` and
`DelegationDistribution.SimpleProposerPool()` for the height of delegation. 

## Commission rate change
 
 - triggered-by: `stake.TxEditValidator`

If a validator changes its commission rate, all commission on fees must be
simultaneously withdrawn using the transaction `TxWithdrawValidator`

