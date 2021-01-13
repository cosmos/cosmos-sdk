<!--
order: 5
-->

# Hooks

In this section we describe the "hooks" - slashing module code that runs when other events happen.
The module provides the following hooks, which can be useful

## Staking hooks

The slashing module implements the x/staking `StakingHooks` used to keep information about validators. During the app initialization, these hooks should be registered in the staking module object.

The following hooks impact the slashing state:

+ `AfterValidatorBonded` - creates a `ValidatorSigningInfo` as described in the following section
+ `AfterValidatorCreated` - registers validator's consensus key
+ `AfterValidatorRemoved` - removes validator's consensus key


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
