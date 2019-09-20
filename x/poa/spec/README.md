# POA module specification

## Overview

The POA module described here is meant to be used as an alternative and less complex module then the original staking located [here](../staking). In this system full nodes can become validators with out having a token at stake. The security model of a public blockchain using the POA should be based on the governance curated registry that defines validator set and assigns corresponding voting power to each individual validator.

## Validator Set Curation

The validator set currently has two options on how its curated, currently this module is built with reliance on the [governance module](../governance/README.md). By default `AcceptAllValidators` is set to false, meaning a validator needs to make a governance proposal to request to be added to the validator set. If `AcceptAllValidators` is set to true then a validator request will not go through governance and the validator will become one instantly. All validators are defaulted to a weight of 10 on creation.

If a validator would like to increase his weight then `IncreaseWeight` must be set to true and the proposal to increase the weight will always have to go through governance, but if a validator would like to decrease their weight then they can do this without a governance proposal.

## Contents

1. **[State](01_state.md)**
   - [Pool](01_state.md#pool)
   - [LastTotalPower](01_state.md#lasttotalpower)
   - [Params](01_state.md#params)
   - [Validator](01_state.md#validator)
   - [Queues](01_state.md#queues)
2. **[State Transitions](02_state_transitions.md)**
   - [Validators](02_state_transitions.md#validators)
   - [Slashing](02_state_transitions.md#slashing)
3. **[Messages](03_messages.md)**
   - [MsgCreateValidator](03_messages.md#msgcreatevalidator)
   - [MsgEditValidator](03_messages.md#msgeditvalidator)
   - [MsgBeginUnbonding](03_messages.md#msgbeginunbonding)
4. **[End-Block ](04_end_block.md)**
   - [Validator Set Changes](04_end_block.md#validator-set-changes)
   - [Queues ](04_end_block.md#queues-)
5. **[Hooks](05_hooks.md)**
6. **[Events](06_events.md)**
   - [EndBlocker](06_events.md#endblocker)
   - [Handlers](06_events.md#handlers)
7. **[Parameters](07_params.md)**
