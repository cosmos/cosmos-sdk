# State Transitions

This document describes the state transition operations pertaining to:

1. [Validators](./02_state_transitions.md#validators)
2. [Delegations](./02_state_transitions.md#delegations)
3. [Slashing](./02_state_transitions.md#slashing)

## Validators

State transitions in validators are performed on every [`EndBlock`](./04_end_block.md#validator-set-changes) in order to check for changes in the active `ValidatorSet`.

### Non-Bonded to Bonded

When a validator is bonded from any other state the following operations occur:  

- set `validator.Status` to `Bonded`
- send the `validator.Tokens` from the `NotBondedTokens` to the `BondedPool` `ModuleAccount`
- delete the existing record from `ValidatorByPowerIndex`
- add a new updated record to the `ValidatorByPowerIndex`
- update the `Validator` object for this validator
- if it exists, delete any `ValidatorQueue` record for this validator

### Bonded to Unbonding

When a validator begins the unbonding process the following operations occur:

- send the `validator.Tokens` from the `BondedPool` to the `NotBondedTokens` `ModuleAccount`
- set `validator.Status` to `Unbonding`
- delete the existing record from `ValidatorByPowerIndex`
- add a new updated record to the `ValidatorByPowerIndex`
- update the `Validator` object for this validator
- insert a new record into the `ValidatorQueue` for this validator

### Unbonding to Unbonded

A validator moves from unbonding to unbonded when the `ValidatorQueue` object
moves from bonded to unbonded

- update the `Validator` object for this validator
- set `validator.Status` to `Unbonded`

### Jail/Unjail

when a validator is jailed it is effectively removed from the Tendermint set.
this process may be also be reversed. the following operations occur:

- set `Validator.Jailed` and update object
- if jailed delete record from `ValidatorByPowerIndex`
- if unjailed add record to `ValidatorByPowerIndex`

## Delegations

### Delegate

When a delegation occurs both the validator and the delegation objects are affected  

- determine the delegators shares based on tokens delegated and the validator's exchange rate
- remove tokens from the sending account
- add shares the delegation object or add them to a created validator object
- add new delegator shares and update the `Validator` object
- transfer the `delegation.Amount`  from the delegator's account to the `BondedPool` or the `NotBondedPool` `ModuleAccount` depending if the `validator.Status` is `Bonded` or not
- delete the existing record from `ValidatorByPowerIndex`
- add an new updated record to the `ValidatorByPowerIndex`

### Begin Unbonding

As a part of the Undelegate and Complete Unbonding state transitions Unbond
Delegation may be called.

- subtract the unbonded shares from delegator
- if the validator is `Unbonding` or `Bonded` add the tokens to an `UnbondingDelegation` Entry
- if the validator is `Unbonded` send the tokens directly to the withdraw
  account
- update the delegation or remove the delegation if there are no more shares
- if the delegation is the operator of the validator and no more shares exist then trigger a jail validator
- update the validator with removed the delegator shares and associated coins
- if the validator state is `Bonded`, transfer the `Coins` worth of the unbonded
shares from the `BondedPool` to the `NotBondedPool` `ModuleAccount`
- remove the validator if it is unbonded and there are no more delegation shares.

### Complete Unbonding

For undelegations which do not complete immediately, the following operations
occur when the unbonding delegation queue element matures:

- remove the entry from the `UnbondingDelegation` object
- transfer the tokens from the `NotBondedPool` `ModuleAccount` to the delegator `Account`

### Begin Redelegation

Redelegations affect the delegation, source and destination validators.

- perform an `unbond` delegation from the source validator to retrieve the tokens worth of the unbonded shares
- using the unbonded tokens, `Delegate` them to the destination validator
- if the `sourceValidator.Status` is `Bonded`, and the `destinationValidator` is not, transfer the newly delegated tokens from the `BondedPool` to the `NotBondedPool` `ModuleAccount`
- otherwise, if the `sourceValidator.Status` is not `Bonded`, and the `destinationValidator` is `Bonded`, transfer the newly delegated tokens from the `NotBondedPool` to the `BondedPool` `ModuleAccount`
- record the token amount in an new entry in the relevant `Redelegation`

### Complete Redelegation

When a redelegations complete the following occurs:

- remove the entry from the `Redelegation` object

## Slashing

### Slash Validator

### Slash Unbonding Delegation

### Slash Redelegation
