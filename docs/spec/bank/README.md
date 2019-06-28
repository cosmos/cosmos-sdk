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

1. **[State](01_state.md)**
2. **[Keepers](02_keepers.md)**
    - [Common Types](02_keepers.md#common-types)
    - [BaseKeeper](02_keepers.md#basekeeper)
    - [SendKeeper](02_keepers.md#sendkeeper)
    - [ViewKeeper](02_keepers.md#viewkeeper)
3. **[Messages](03_messages.md)**
    - [MsgSend](03_messages.md#msgsend)
4. **[Events](04_events.md)**
    - [Handlers](04_events.md#handlers)
5. **[Parameters](05_params.md)**
