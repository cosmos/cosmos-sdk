# End-Block 

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

Complete the unbonding and transfer the coins to the delegate. Perform any
slashing that occurred during the unbonding period.

```golang
unbondingQueue(currTime time.Time):
    for all unbondings whose CompleteTime < currTime:
        validator = GetValidator(unbonding.ValidatorAddr)
        returnTokens = ExpectedTokens * unbonding.startSlashRatio/validator.SlashRatio
        AddCoins(unbonding.DelegatorAddr, returnTokens)
        removeUnbondingDelegation(unbonding)
    return
```



## CompleteRedelegation

Note that unlike CompleteUnbonding slashing of redelegating shares does not
take place during completion. Slashing on redelegated shares takes place
actively as a slashing occurs.

```golang
redelegationQueue(currTime time.Time):
    for all redelegations whose CompleteTime < currTime:
        removeRedelegation(redelegation)
    return
```