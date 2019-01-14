# Governance module specification

## Abstract

This paper specifies the Governance module of the Cosmos-SDK, which was first described in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper) in June 2016.

The module enables Cosmos-SDK based blockchain to support an on-chain governance system. In this system, holders of the native staking token of the chain can vote on proposals on a 1 token 1 vote basis. Next is a list of features the module currently supports:

- **Proposal submission:** Users can submit proposals with a deposit. Once the minimum deposit is reached, proposal enters voting period
- **Vote:** Participants can vote on proposals that reached MinDeposit
- **Inheritance and penalties:** Delegators inherit their validator's vote if they don't vote themselves. If validators do not vote, they get partially slashed.
- **Signal and switch:** If a proposal of type `SoftwareUpgradeProposal` is accepted, validators can signal it and switch once enough validators have signalled.
- **Claiming deposit:** Users that deposited on proposals can recover their deposits if the proposal was accepted OR if the proposal never entered voting period.

Features that may be added in the future are described in [Future improvements](future_improvements.md)

This module will be used in the Cosmos Hub, the first Hub in the Cosmos network.


## Contents

The following specification uses *Atom* as the native staking token. The module can be adapted to any Proof-Of-Stake blockchain by replacing *Atom* with the native staking token of the chain.

1.  **[Design overview](overview.md)**
2.  **Implementation**
    1. **[State](state.md)**
        1.  Parameters
        2.  Proposals
        3.  Proposal Processing Queue
    2. **[Transactions](transactions.md)**
        1.  Proposal Submission
        2.  Deposit
        3.  Claim Deposit
        4.  Vote
3.  **[Future improvements](future_improvements.md)**
