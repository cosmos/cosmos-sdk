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

1. **[State](01_state.md)**
    - [Pool](01_state.md#pool)
    - [LastTotalPower](01_state.md#lasttotalpower)
    - [Params](01_state.md#params)
    - [Validator](01_state.md#validator)
    - [Delegation](01_state.md#delegation)
    - [UnbondingDelegation](01_state.md#unbondingdelegation)
    - [Redelegation](01_state.md#redelegation)
    - [Queues](01_state.md#queues)
2. **[State Transitions](02_state_transitions.md)**
    - [Validators](02_state_transitions.md#validators)
    - [Delegations](02_state_transitions.md#delegations)
    - [Slashing](02_state_transitions.md#slashing)
3. **[Messages](03_messages.md)**
    - [MsgCreateValidator](03_messages.md#msgcreatevalidator)
    - [MsgEditValidator](03_messages.md#msgeditvalidator)
    - [MsgDelegate](03_messages.md#msgdelegate)
    - [MsgBeginUnbonding](03_messages.md#msgbeginunbonding)
    - [MsgBeginRedelegate](03_messages.md#msgbeginredelegate)
4. **[End-Block ](04_end_block.md)**
    - [Validator Set Changes](04_end_block.md#validator-set-changes)
    - [Queues ](04_end_block.md#queues-)
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
    - [EndBlocker](06_events.md#endblocker)
    - [Handlers](06_events.md#handlers)
7. **[Parameters](07_params.md)**
