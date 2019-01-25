# State Transitions

This document describes the state transition operations pertaining to:
 - Validators
 - Delegations
 - Slashing

## Validators

### non-bonded to bonded
 - delete record from `ValidatorByPowerIndex`
 - set `Validator.BondHeight` to current height

### unbonding to unbonded
 -> completeUnbondingValidator

### bonded to unbonding
 -> beginUnbondingValidator

### jail/unjail 
when a validator is jailed it is effectively removed from the Tendermint set.
this process may be also be reversed. the following operations occur:
 - set `Validator.Jailed` and update object 
 - if jailed delete record from `ValidatorByPowerIndex`
 - if unjailed add record to `ValidatorByPowerIndex`

## Delegations

### Delegate
   ### AddValidatorTokensAndShares
### unbond
### Undelegate
   ### RemoveValidatorTokensAndShares
### CompleteUnbonding
### BeginRedelegation
### CompleteRedelegation

## Slashing
### Slash
      ### RemoveValidatorTokens
### slashUnbondingDelegation
### slashRedelegation

