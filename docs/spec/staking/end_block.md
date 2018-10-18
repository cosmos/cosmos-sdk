# End-Block 

## Unbonding Validator Queue

For all unbonding validators that have finished their unbonding period, this switches their validator.Status
from sdk.Unbonding to sdk.Unbonded if they still have any delegation left.  Otherwise, it deletes it from state.

```golang
validatorQueue(currTime time.Time):
    // unbonding validators are in ordered queue from oldest to newest
    for all unbondingValidators whose CompleteTime < currTime:
        validator = GetValidator(unbondingValidator.ValidatorAddr)
        if validator.DelegatorShares == 0 {
            RemoveValidator(unbondingValidator)
        } else {
            validator.Status = sdk.Unbonded
            SetValidator(unbondingValidator)
        }
    return
```

## Validator Set Changes

The Tendermint validator set may be updated by state transitions that run at
the end of every block. The Tendermint validator set may be changed by
validators either being jailed due to inactivity/unexpected behaviour (covered
in slashing) or changed in validator power. Determining which validator set
changes must be made occurs during staking transactions (and slashing
transactions) - during end-block the already accounted changes are applied and
the changes cleared

```golang
EndBlock() ValidatorSetChanges
    vsc = GetValidTendermintUpdates()
    ClearTendermintUpdates()
    return vsc
```

## CompleteUnbonding

Complete the unbonding and transfer the coins to the delegate. Realize any
slashing that occurred during the unbonding period.

```golang
unbondingQueue(currTime time.Time):
    // unbondings are in ordered queue from oldest to newest
    for all unbondings whose CompleteTime < currTime:
        validator = GetValidator(unbonding.ValidatorAddr)
        AddCoins(unbonding.DelegatorAddr, unbonding.Balance)
        removeUnbondingDelegation(unbonding)
    return
```

## CompleteRedelegation

Note that unlike CompleteUnbonding slashing of redelegating shares does not
take place during completion. Slashing on redelegated shares takes place
actively as a slashing occurs. The redelegation completion queue serves simply to
clean up state, as redelegations older than an unbonding period need not be kept,
as that is the max time that their old validator's evidence can be used to slash them.

```golang
redelegationQueue(currTime time.Time):
    // redelegations are in ordered queue from oldest to newest
    for all redelegations whose CompleteTime < currTime:
        removeRedelegation(redelegation)
    return
```
