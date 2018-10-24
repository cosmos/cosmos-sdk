## Hooks

In this section we describe the "hooks" - slashing module code that runs when other events happen.

### Validator Bonded

Upon successful bonding of a validator (a given validator entering the "bonded" state,
which may happen on delegation, on unjailing, etc), we create a new `SlashingPeriod` structure for the
now-bonded validator, which `StartHeight` of the current block, `EndHeight` of `0` (sentinel value for not-yet-ended),
and `SlashedSoFar` of `0`:

```
onValidatorBonded(address sdk.ValAddress)

  signingInfo, found = getValidatorSigningInfo(address)
  if !found {
    signingInfo = ValidatorSigningInfo {
      StartHeight         : CurrentHeight,
      IndexOffset         : 0,
      JailedUntil         : time.Unix(0, 0),
      MissedBloskCounter  : 0
    }
    setValidatorSigningInfo(signingInfo)
  }

  slashingPeriod = SlashingPeriod{
      ValidatorAddr : address,
      StartHeight   : CurrentHeight,
      EndHeight     : 0,    
      SlashedSoFar  : 0,
  }
  setSlashingPeriod(slashingPeriod)
  
  return
```

### Validator Unbonded

When a validator is unbonded, we update the in-progress `SlashingPeriod` with the current block as the `EndHeight`:

```
onValidatorUnbonded(address sdk.ValAddress)

  slashingPeriod = getSlashingPeriod(address, CurrentHeight)
  slashingPeriod.EndHeight = CurrentHeight
  setSlashingPeriod(slashingPeriod)

  return
```

### Validator Slashed

When a validator is slashed, we look up the appropriate `SlashingPeriod` based on the validator
address and the time of infraction, cap the fraction slashed as `max(SlashFraction, SlashedSoFar)`
(which may be `0`), and update the `SlashingPeriod` with the increased `SlashedSoFar`:

```
beforeValidatorSlashed(address sdk.ValAddress, fraction sdk.Rat, infractionHeight int64)
  
  slashingPeriod = getSlashingPeriod(address, infractionHeight)
  totalToSlash = max(slashingPeriod.SlashedSoFar, fraction)
  slashingPeriod.SlashedSoFar = totalToSlash
  setSlashingPeriod(slashingPeriod)

  remainderToSlash = slashingPeriod.SlashedSoFar - totalToSlash
  fraction = remainderToSlash

  continue with slashing
```
