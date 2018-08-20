## Transaction & State Machine Interaction Overview

### Transactions

In this section we describe the processing of transactions for the `slashing` module.

#### TxUnjail

If a validator was automatically unbonded due to downtime and wishes to come back online &
possibly rejoin the bonded set, it must send `TxUnjail`:

```golang
type TxUnjail struct {
    ValidatorAddr sdk.AccAddress
}

handleMsgUnjail(tx TxUnjail)

    validator := getValidator(tx.ValidatorAddr)
    if validator == nil
      fail with "No validator found"

    if !validator.Jailed
      fail with "Validator not jailed, cannot unjail"

    info := getValidatorSigningInfo(operator)
    if BlockHeader.Time.Before(info.JailedUntil)
      fail with "Validator still jailed, cannot unjail until period has expired"

    // Update the start height so the validator won't be immediately unbonded again
    info.StartHeight = BlockHeight
    setValidatorSigningInfo(info)

    validator.Jailed = false
    setValidator(validator)

    return
```

If the validator has enough stake to be in the top hundred, they will be automatically rebonded,
and all delegators still delegated to the validator will be rebonded and begin to again collect
provisions and rewards.

### Interactions

In this section we describe the "hooks" - slashing module code that runs when other events happen.

#### Validator Bonded

Upon successful bonding of a validator (a given validator changing from "unbonded" state to "bonded" state,
which may happen on delegation, on unjailing, etc), we create a new `SlashingPeriod` structure for the
now-bonded validator, which `StartHeight` of the current block, `EndHeight` of `0` (sentinel value for not-yet-ended),
and `SlashedSoFar` of `0`:

```golang
onValidatorBonded(address sdk.ValAddress)

  slashingPeriod := SlashingPeriod{
      ValidatorAddr : address,
      StartHeight   : CurrentHeight,
      EndHeight     : 0,    
      SlashedSoFar  : 0,
  }
  setSlashingPeriod(slashingPeriod)
  
  return
```

#### Validator Unbonded

When a validator is unbonded, we update the in-progress `SlashingPeriod` with the current block as the `EndHeight`:

```golang
onValidatorUnbonded(address sdk.ValAddress)

  slashingPeriod = getSlashingPeriod(address, CurrentHeight)
  slashingPeriod.EndHeight = CurrentHeight
  setSlashingPeriod(slashingPeriod)

  return
```

#### Validator Slashed

When a validator is slashed, we look up the appropriate `SlashingPeriod` based on the validator
address and the time of infraction, cap the fraction slashed as `max(SlashFraction, SlashedSoFar)`
(which may be `0`), and update the `SlashingPeriod` with the increased `SlashedSoFar`:

```golang
beforeValidatorSlashed(address sdk.ValAddress, fraction sdk.Rat, infractionHeight int64)
  
  slashingPeriod = getSlashingPeriod(address, infractionHeight)
  totalToSlash = max(slashingPeriod.SlashedSoFar, fraction)
  slashingPeriod.SlashedSoFar = totalToSlash
  setSlashingPeriod(slashingPeriod)

  remainderToSlash = slashingPeriod.SlashedSoFar - totalToSlash
  fraction = remainderToSlash

  continue with slashing
```

##### Safety note

Slashing is capped fractionally per period, but the amount of total bonded stake associated with any given validator can change (by an unbounded amount) over that period.

For example, with MaxFractionSlashedPerPeriod = `0.5`, if a validator is initially slashed at `0.4` near the start of a period when they have 100 steak bonded,
then later slashed at `0.4` when they have `1000` steak bonded, the total amount slashed is just `40 + 100 = 140` (since the latter slash is capped at `0.1`) - 
whereas if they had `1000` steak bonded initially, the total amount slashed would have been `500`.

This means that any slashing events which utilize the slashing period (are capped-per-period) **must** *also* jail the validator when the infraction is discovered.
Otherwise it would be possible for a validator to slash themselves intentionally at a low bond, then increase their bond but no longer be at stake since they would have already hit the `SlashedSoFar` cap.

### State Cleanup

Once no evidence for a given slashing period can possibly be valid (the end time plus the unbonding period is less than the current time),
old slashing periods should be cleaned up. This will be implemented post-launch.
