<!--
order: 5
-->

# Hooks

## Create or modify delegation distribution
 
 - triggered-by: `staking.MsgDelegate`, `staking.MsgBeginRedelegate`, `staking.MsgUndelegate`

The pool of a new delegator bond will be 0 for the height at which the bond was
added, or the withdrawal has taken place. This is achieved by setting
`DelegationDistInfo.WithdrawalHeight` to the height of the triggering transaction. 

## Commission rate change
 
 - triggered-by: `staking.MsgEditValidator`

If a validator changes its commission rate, all commission on fees must be
simultaneously withdrawn using the transaction `TxWithdrawValidator`.
Additionally the change and associated height must be recorded in a
`ValidatorUpdate` state record.

## Change in Validator State
 
 - triggered-by: `staking.Slash`, `staking.UpdateValidator`

Whenever a validator is slashed or enters/leaves the validator group all of the
validator entitled reward tokens must be simultaneously withdrawn from
`Global.Pool` and added to `ValidatorDistInfo.Pool`. 

Thoughts: MsgDelegate and MsgBeginDelegate, MsgUndelegate, Commission rate, validator state change  take action on epoching process
There are two ways
1) All hooks happen at epoching process, but it could be too heavy to do all of these at a single endblocker
2) Staging hooks happen and distribution module take Staging changes without actual distribution change, and make this to take effect on epoching. It involves `DelegationDistInfo` data management.