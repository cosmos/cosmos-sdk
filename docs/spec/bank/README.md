# Bank module specification

## Abstract

This document specifies the bank module of the Cosmos SDK.

The bank module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts (notably delegating/undelegating for vesting
accounts). It exposes several interfaces with varying capabilities for secure
interaction with other modules which must alter user balances.

This module will be used in the Cosmos Hub.

## Contents

1. **[State](state.md)**
1. **[Keepers](keepers.md)**
    1. [Common Types](keepers.md#common-types)
        1. [Input](keepers.md#input)
        1. [Output](keepers.md#output)
    1. [BaseKeeper](keepers.md#basekeeper)
    1. [SendKeeper](keepers.md#sendkeeper)
    1. [ViewKeeper](keepers.md#viewkeeper) 
1. **[Transactions](transactions.md)**
    1. [MsgSend](transactions.md#msgsend)
