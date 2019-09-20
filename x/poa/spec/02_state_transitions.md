# State Transitions

This document describes the state transition operations pertaining to:

1. [Validators](./02_state_transitions.md#validators)
2. [Slashing](./02_state_transitions.md#slashing)

## Validators

State transitions in validators are performed on every [`EndBlock`](./04_end_block.md#validator-set-changes) in order to check for changes in the active `ValidatorSet`.

### Non-Bonded to Bonded

When a validator is bonded from any other state the following operations occur:

- set `validator.Status` to `Bonded`
- delete the existing record from `ValidatorByPowerIndex`
- add a new updated record to the `ValidatorByPowerIndex`
- update the `Validator` object for this validator
- if it exists, delete any `ValidatorQueue` record for this validator

### Bonded to Unbonding

When a validator begins the unbonding process the following operations occur:

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

### Begin Unbonding

When a validator wants to remove himself from the validator.
The following operations occur:

- subtract the total weight from the validator
- remove the validator if it is unbonded and there is no more weight associated with it.

## Slashing

### Slash Validator

### Slash Unbonding Delegation

### Slash Redelegation
