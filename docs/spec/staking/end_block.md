# End-Block 

## Validator Set Changes

The Tendermint validator set is be updated by state transitions that run at the
end of every block. Operations are as following:

 - The new validator set is taken as the top `params.MaxValidators` number of
   validators retrieved from the ValidatorsByPower index
 - The previous validator set is compared with the new validator set 
   - missing validators begin unbonding
   - new validator are instantly bonded

in all cases any validators leaving or entering the bonded validator set or
changing balances and staying within the bonded validator set incure an update
message which is passed back to Tendermint.

## Queues 

### Unbonding Validators

Each block the validator queue is to be checked for mature unbonding
validators.  For all unbonding validators that have finished their unbonding
period, the validator.Status is switched from sdk.Unbonding to sdk.Unbonded if
they still have any delegation left.  Otherwise, the validator object is
deleted from state.

### Unbonding Delegations

Complete the unbonding of all mature `UnbondingDelegations.Entries` by
transferring the balance coins to the delegator's wallet address and removing
the mature `UnbondingDelegation.Entries`. If there are no remaining entries also
remove the `UnbondingDelegation` object from the store. 

### Redelegations

Complete the unbonding of all mature `Redelegation.Entries` by removing the
mature `Redelegation.Entries`. If there are no remaining entries also remove
the `Redelegation` object from the store. 

