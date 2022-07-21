# ADR 056: AUTO-COMPOUNDING

## Changelog

* 2022 July 01: An [implementation](https://github.com/cosmos/cosmos-sdk/pull/12414) is proposed
* 2022 July 21: Initial draft of the ADR

## Status

DRAFT

## Abstract

This ADR proposes to add a feature to the `x/distribution` module to enable users to automatically withdraw their staking rewards and delegate them again to validator.

## Context

Currently if a user want's to delegate their staking rewards, they need to do it manually by taking the following steps:

1. Check the staking rewards.
2. Withdraw the rewards if they are above a threshold.
3. Select a validator
4. Delegate the rewards to it and keep some amount in the account
5. Wait for a while then jump to #1 

This ADR proposes to automate this process.


### Proposed Implementation
An implementation which uses `BeginBlock` is proposed here: https://github.com/cosmos/cosmos-sdk/pull/12414

It basically has a linked list and processes `1` request per block in order to reduce overhead.
There are a number of design choices that have to be discussed.

### Design Choices (questions to discuss)

1. How many delegations should be processed per block? (_currently one_)
2. What is the best interval for the entire process execution? Right now it is based on `n` number of blocks.
3. How to select a validator to delegate to? Currently, the validator with the minimum delegation of the current account is selected. But there are a number of choices here:
    1.  A user defined ratio for each validator in % to distribute the delegations
    2. Select them pseudo randomly in a deterministic way
    3.  Select the validators based on a specific criteria
    4. Delegate based on the ratio that user has delegated manually
    5. Another way
4. Currently the account address is stored in string. It could be more efficient to store it in sdk.Address.

## Decision

> {decision body}

## Consequences

> {consequences}

### Backwards Compatibility


### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}

## Further Discussions

> {further discussions}

## References

* https://github.com/cosmos/cosmos-sdk/pull/12414
