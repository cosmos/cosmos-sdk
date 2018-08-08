# End-Block 

## Validator Set Changes

The Tendermint validator set may be updated by state transitions that run at
the end of every block. The Tendermint validator set may be changed by
validators either being revoked due to inactivity/unexpected behaviour (covered
in slashing) or changed in validator power. Determining which validator set
changes must be made occurs during staking transactions (and slashing
transactions) - during end-block the already accounted changes are applied and
the changes cleared

```golang
EndBlock() ValidatorSetChanges
    vsc = GetTendermintUpdates()
    ClearTendermintUpdates()
    return vsc
```

