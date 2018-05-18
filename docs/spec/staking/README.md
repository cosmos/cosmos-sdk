# Staking module specification

## Abstract

This paper specifies the Staking module of the Cosmos-SDK, which was first
described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper)
in June 2016. 

The module enables Cosmos-SDK based blockchain to support an advanced
Proof-of-Stake system. In this system, holders of the native staking token of
the chain can become validators and can delegate tokens to validator
validators, ultimately determining the effective validator set for the system.

This module will be used in the Cosmos Hub, the first Hub in the Cosmos
network.

## Contents

The following specification uses *Atom* as the native staking token. The module
can be adapted to any Proof-Of-Stake blockchain by replacing *Atom* with the
native staking token of the chain.

1.  **[Design overview](overview.md)**
2.  **Implementation**
    1. **[State](state.md)**
        1.  Params
        1.  Pool
        2.  Validators
        3.  Delegations
    2. **[Transactions](transactions.md)**
        1.  Create-Validator
        2.  Edit-Validator
        3.  Repeal-Revocation
        4.  Delegate
        5.  Unbond
        6.  Redelegate
    3. **[Validator Set Changes](valset-changes.md)**
        1.  Validator set updates
        2.  Slashing
        3.  Automatic Unbonding
3.  **[Future improvements](future_improvements.md)**
