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
2. **[Keepers](keepers.md)**
3. **[Transactions](transactions.md)**
