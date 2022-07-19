<!--
order: 2
-->

# State Transitions

This document describes the state transition operations pertaining to:

1. [Validators](./02_state_transitions.md#validators)
2. [Delegations](./02_state_transitions.md#delegations)
3. [Slashing](./02_state_transitions.md#slashing)

## Validators

State transitions in validators are performed on every [`EndBlock`](./05_end_block.md#validator-set-changes)
in order to check for changes in the active `ValidatorSet`.

A validator can be `Unbonded`, `Unbonding` or `Bonded`. `Unbonded`
and `Unbonding` are collectively called `Not Bonded`. A validator can move
directly between all the states, except for from `Bonded` to `Unbonded`.

### Not bonded to Bonded

The following transition occurs when a validator's ranking in the `ValidatorPowerIndex` surpasses
that of the `LastValidator`.

* set `validator.Status` to `Bonded`
* send the `validator.Tokens` from the `NotBondedTokens` to the `BondedPool` `ModuleAccount`
* delete the existing record from `ValidatorByPowerIndex`
* add a new updated record to the `ValidatorByPowerIndex`
* update the `Validator` object for this validator
* if it exists, delete any `ValidatorQueue` record for this validator

### Bonded to Unbonding

When a validator begins the unbonding process the following operations occur:

* send the `validator.Tokens` from the `BondedPool` to the `NotBondedTokens` `ModuleAccount`
* set `validator.Status` to `Unbonding`
* delete the existing record from `ValidatorByPowerIndex`
* add a new updated record to the `ValidatorByPowerIndex`
* update the `Validator` object for this validator
* insert a new record into the `ValidatorQueue` for this validator

### Unbonding to Unbonded

A validator moves from unbonding to unbonded when the `ValidatorQueue` object
moves from bonded to unbonded

* update the `Validator` object for this validator
* set `validator.Status` to `Unbonded`

### Jail/Unjail

when a validator is jailed it is effectively removed from the Tendermint set.
this process may be also be reversed. the following operations occur:

* set `Validator.Jailed` and update object
* if jailed delete record from `ValidatorByPowerIndex`
* if unjailed add record to `ValidatorByPowerIndex`

Jailed validators are not present in any of the following stores:

* the power store (from consensus power to address)

## Delegations

### Delegate

When a delegation occurs both the validator and the delegation objects are affected

* determine the delegators shares based on tokens delegated and the validator's exchange rate
* remove tokens from the sending account
* add shares the delegation object or add them to a created validator object
* add new delegator shares and update the `Validator` object
* transfer the `delegation.Amount` from the delegator's account to the `BondedPool` or the `NotBondedPool` `ModuleAccount` depending if the `validator.Status` is `Bonded` or not
* delete the existing record from `ValidatorByPowerIndex`
* add an new updated record to the `ValidatorByPowerIndex`

### Begin Unbonding

As a part of the Undelegate and Complete Unbonding state transitions Unbond
Delegation may be called.

* subtract the unbonded shares from delegator
* add the unbonded tokens to an `UnbondingDelegation` Entry
* update the delegation or remove the delegation if there are no more shares
* if the delegation is the operator of the validator and no more shares exist then trigger a jail validator
* update the validator with removed the delegator shares and associated coins
* if the validator state is `Bonded`, transfer the `Coins` worth of the unbonded
  shares from the `BondedPool` to the `NotBondedPool` `ModuleAccount`
* remove the validator if it is unbonded and there are no more delegation shares.

### Cancel an `UnbondingDelegation` Entry 
When a `cancel unbond delegation` occurs both the `validator`, the `delegation` and an `UnbondingDelegationQueue` state will be updated.
* if cancel unbonding delegation amount equals to the `UnbondingDelegation` entry `balance`, then the `UnbondingDelegation` entry deleted from `UnbondingDelegationQueue`.
* if the `cancel unbonding delegation amount is less than the `UnbondingDelegation` entry balance, then the `UnbondingDelegation` entry will be updated with new balance in the `UnbondingDelegationQueue`. 
* cancel `amount` is [Delegated](02_state_transitions.md#delegations) back to  the original `validator`.

### Complete Unbonding

For undelegations which do not complete immediately, the following operations
occur when the unbonding delegation queue element matures:

* remove the entry from the `UnbondingDelegation` object
* transfer the tokens from the `NotBondedPool` `ModuleAccount` to the delegator `Account`

### Begin Redelegation

Redelegations affect the delegation, source and destination validators.

* perform an `unbond` delegation from the source validator to retrieve the tokens worth of the unbonded shares
* using the unbonded tokens, `Delegate` them to the destination validator
* if the `sourceValidator.Status` is `Bonded`, and the `destinationValidator` is not,
  transfer the newly delegated tokens from the `BondedPool` to the `NotBondedPool` `ModuleAccount`
* otherwise, if the `sourceValidator.Status` is not `Bonded`, and the `destinationValidator`
  is `Bonded`, transfer the newly delegated tokens from the `NotBondedPool` to the `BondedPool` `ModuleAccount`
* record the token amount in an new entry in the relevant `Redelegation`

From when a redelegation begins until it completes, the delegator is in a state of "pseudo-unbonding", and can still be
slashed for infractions that occured before the redelegation began.

### Complete Redelegation

When a redelegations complete the following occurs:

* remove the entry from the `Redelegation` object

## Slashing

### Slash Validator

When a Validator is slashed, the following occurs:

* The total `slashAmount` is calculated as the `slashFactor` (a chain parameter) \* `TokensFromConsensusPower`,
  the total number of tokens bonded to the validator at the time of the infraction.
* Every unbonding delegation and pseudo-unbonding redelegation such that the infraction occured before the unbonding or
  redelegation began from the validator are slashed by the `slashFactor` percentage of the initialBalance.
* Each amount slashed from redelegations and unbonding delegations is subtracted from the
  total slash amount.
* The `remaingSlashAmount` is then slashed from the validator's tokens in the `BondedPool` or
  `NonBondedPool` depending on the validator's status. This reduces the total supply of tokens.

In the case of a slash due to any infraction that requires evidence to submitted (for example double-sign), the slash
occurs at the block where the evidence is included, not at the block where the infraction occured.
Put otherwise, validators are not slashed retroactively, only when they are caught.

### Slash Unbonding Delegation

When a validator is slashed, so are those unbonding delegations from the validator that began unbonding
after the time of the infraction. Every entry in every unbonding delegation from the validator
is slashed by `slashFactor`. The amount slashed is calculated from the `InitialBalance` of the
delegation and is capped to prevent a resulting negative balance. Completed (or mature) unbondings are not slashed.

### Slash Redelegation

When a validator is slashed, so are all redelegations from the validator that began after the
infraction. Redelegations are slashed by `slashFactor`.
Redelegations that began before the infraction are not slashed.
The amount slashed is calculated from the `InitialBalance` of the delegation and is capped to
prevent a resulting negative balance.
Mature redelegations (that have completed pseudo-unbonding) are not slashed.

## How Shares are calculated

At any given point in time, each validator has a number of tokens, `T`, and has a number of shares issued, `S`.
Each delegator, `i`, holds a number of shares, `S_i`.
The number of tokens is the sum of all tokens delegated to the validator, plus the rewards, minus the slashes.

The delegator is entitled to a portion of the underlying tokens proportional to their proportion of shares.
So delegator `i` is entitled to `T * S_i / S` of the validator's tokens.

When a delegator delegates new tokens to the validator, they receive a number of shares proportional to their contribution.
So when delegator `j` delegates `T_j` tokens, they receive `S_j = S * T_j / T` shares.
The total number of tokens is now `T + T_j`, and the total number of shares is `S + S_j`.
`j`s proportion of the shares is the same as their proportion of the total tokens contributed: `(S + S_j) / S = (T + T_j) / T`.

A special case is the initial delegation, when `T = 0` and `S = 0`, so `T_j / T` is undefined.
For the initial delegation, delegator `j` who delegates `T_j` tokens receive `S_j = T_j` shares.
So a validator that hasn't received any rewards and has not been slashed will have `T = S`.
