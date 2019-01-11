# End-Block 

Each block the following operations occur:

## Unbonding Validator Maturity

Each block the validator queue is to be checked for mature unbonding
validators.  For all unbonding validators that have finished their unbonding
period, the validator.Status is switched from sdk.Unbonding to sdk.Unbonded if
they still have any delegation left.  Otherwise, the validator object is
deleted from state.

## Validator Set Changes

The Tendermint validator set may be updated by state transitions that run at
the end of every block. The Tendermint validator set may be changed by
validators either being jailed due to inactivity/unexpected behaviour (covered
in slashing) or changed in validator power. Determining which validator set
changes must be made occurs during staking transactions (and slashing
transactions) - during end-block the already accounted changes are applied and
the changes cleared

## CompleteUnbonding

Complete the unbonding of all mature `UnbondingDelegations.Entries` by
transferring the balance coins to the delegator's wallet address and removing
the mature `UnbondingDelegation.Entries`. If there are no remaining entries also
remove the `UnbondingDelegation` object from the store. 

## CompleteRedelegation

Complete the unbonding of all mature `Redelegation.Entries` by removing the
mature `Redelegation.Entries`. If there are no remaining entries also remove
the `Redelegation` object from the store. 

