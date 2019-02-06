# State Transitions

This document describes the state transition operations pertaining to:
 - Validators
 - Delegations
 - Slashing


## Validators

### non-bonded to bonded
When a validator is bonded from any other state the following operations occur:  
 - set `Validator.BondHeight` to current height
 - set `validator.Status` to `Bonded`
 - update the `Pool` object with tokens moved from `NotBondedTokens` to `BondedTokens`
 - delete record the existing record from `ValidatorByPowerIndex`
 - add an new updated record to the `ValidatorByPowerIndex`
 - update the `Validator` object for this validator
 - if it exists, delete any `ValidatorQueue` record for this validator 
 - call the `AfterValidatorBonded` hook

### bonded to unbonding
When a validator begins the unbonding process the following operations occur: 
 - update the `Pool` object with tokens moved from `BondedTokens` to `NotBondedTokens`
 - set `validator.Status` to `Unbonding`
 - delete record the existing record from `ValidatorByPowerIndex`
 - add an new updated record to the `ValidatorByPowerIndex`
 - update the `Validator` object for this validator
 - insert a new record into the `ValidatorQueue` for this validator 
 - call the `AfterValidatorBeginUnbonding` hook

### unbonding to unbonded
A validator moves from unbonding to unbonded when the `ValidatorQueue` object
moves from bonded to unbonded
 - update the `Validator` object for this validator
 - set `validator.Status` to `Unbonded`


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

