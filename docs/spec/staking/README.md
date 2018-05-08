# Staking module specification

## Abstract

This paper specifies the Staking module of the Cosmos-SDK, which was first described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper) in June 2016. 

The module enables Cosmos-SDK based blockchain to support an advanced Proof-of-Stake system. In this system, holders of the native staking token of the chain can become candidate validators and can delegate tokens to candidate validators, ultimately determining the effective validator set for the system.

The module currently supports the following features:

- TODO
- **Declare Candidacy:** 
- **Edit Candidacy:** 
- **Delegate:** 
- **Unbond:** 

This module will be used in the Cosmos Hub, the first Hub in the Cosmos network.

## Contents

The following specification uses *Atom* as the native staking token. The module can be adapted to any Proof-Of-Stake blockchain by replacing *Atom* with the native staking token of the chain.

1.  **[Design overview](overview.md)**
2.  **Implementation**
    1. **[State](state.md)**
        1.  Global State
        2.  Validator Candidates
        3.  Delegator Bonds
        4.  Unbond and Rebond Queue
    2. **[Transactions](transactions.md)**
        1.  Declare Candidacy
        2.  Edit Candidacy
        3.  Delegate
        4.  Unbond
        5.  Redelegate
        6.  ProveLive
3.  **[Future improvements](future_improvements.md)**
