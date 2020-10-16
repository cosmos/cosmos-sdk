<!--
order: 5
-->

# Hooks

In this section we describe the "hooks" - slashing module code that runs when other events happen.

## Validator Bonded

Upon successful first-time bonding of a new validator, we create a new `ValidatorSigningInfo` structure for the
now-bonded validator, which `StartHeight` of the current block.

```
onValidatorBonded(address sdk.ValAddress)

  signingInfo, found = GetValidatorSigningInfo(address)
  if !found {
    signingInfo = ValidatorSigningInfo {
      StartHeight         : CurrentHeight,
      IndexOffset         : 0,
      JailedUntil         : time.Unix(0, 0),
      Tombstone           : false,
      MissedBloskCounter  : 0
    }
    setValidatorSigningInfo(signingInfo)
  }

  return
```
