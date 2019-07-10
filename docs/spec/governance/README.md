# Governance module specification

## Abstract

This paper specifies the Governance module of the Cosmos-SDK, which was first
described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper) in
June 2016.

The module enables Cosmos-SDK based blockchain to support an on-chain governance
system. In this system, holders of the native staking token of the chain can vote
on proposals on a 1 token 1 vote basis. Next is a list of features the module
currently supports:

- **Proposal submission:** Users can submit proposals with a deposit. Once the
minimum deposit is reached, proposal enters voting period
- **Vote:** Participants can vote on proposals that reached MinDeposit
- **Inheritance and penalties:** Delegators inherit their validator's vote if
they don't vote themselves.
- **Claiming deposit:** Users that deposited on proposals can recover their
deposits if the proposal was accepted OR if the proposal never entered voting period.

This module will be used in the Cosmos Hub, the first Hub in the Cosmos network.
Features that may be added in the future are described in [Future Improvements](05_future_improvements.md).

## Contents

The following specification uses *ATOM* as the native staking token. The module
can be adapted to any Proof-Of-Stake blockchain by replacing *ATOM* with the native
staking token of the chain.

1. **[Concepts](01_concepts.md)**
    - [Proposal submission](01_concepts.md#proposal-submission)
    - [Vote](01_concepts.md#vote)
    - [Software Upgrade](01_concepts.md#software-upgrade)
2. **[State](02_state.md)**
    - [Parameters and base types](02_state.md#parameters-and-base-types)
    - [Deposit](02_state.md#deposit)
    - [ValidatorGovInfo](02_state.md#validatorgovinfo)
    - [Proposals](02_state.md#proposals)
    - [Stores](02_state.md#stores)
    - [Proposal Processing Queue](02_state.md#proposal-processing-queue)
3. **[Messages](03_messages.md)**
    - [Proposal Submission](03_messages.md#proposal-submission)
    - [Deposit](03_messages.md#deposit)
    - [Vote](03_messages.md#vote)
4. **[Events](04_events.md)**
    - [EndBlocker](04_events.md#endblocker)
    - [Handlers](04_events.md#handlers)
5. **[Future Improvements](05_future_improvements.md)**
6. **[Parameters](06_params.md)**
