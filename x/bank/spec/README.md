<!--
order: 0
title: Bank Overview
parent:
  title: "bank"
-->

# `x/bank`

## Abstract

This document specifies the bank module of the Cosmos SDK.

The bank module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts (notably delegating/undelegating for vesting
accounts). It exposes several interfaces with varying capabilities for secure
interaction with other modules which must alter user balances.

In addition, the bank module tracks and provides query support for the total
supply of all assets used in the application.

This module will be used in the Cosmos Hub.

## Supply

The `supply` functionality:

- passively tracks the total supply of coins within a chain,
- provides a pattern for modules to hold/interact with `Coins`, and
- introduces the invariant check to verify a chain's total supply.

### Total Supply

The total `Supply` of the network is equal to the sum of all coins from the
account. The total supply is updated every time a `Coin` is minted (eg: as part
of the inflation mechanism) or burned (eg: due to slashing or if a governance
proposal is vetoed).

## Module Accounts

The supply functionality introduces a new type of `auth.Account` which can be used by
modules to allocate tokens and in special cases mint or burn tokens. At a base
level these module accounts are capable of sending/receiving tokens to and from
`auth.Account`s and other module accounts. This design replaces previous
alternative designs where, to hold tokens, modules would burn the incoming
tokens from the sender account, and then track those tokens internally. Later,
in order to send tokens, the module would need to effectively mint tokens
within a destination account. The new design removes duplicate logic between
modules to perform this accounting.

The `ModuleAccount` interface is defined as follows:

```go
type ModuleAccount interface {
  auth.Account               // same methods as the Account interface

  GetName() string           // name of the module; used to obtain the address
  GetPermissions() []string  // permissions of module account
  HasPermission(string) bool
}
```

> **WARNING!**
> Any module or message handler that allows either direct or indirect sending of funds must explicitly guarantee those funds cannot be sent to module accounts (unless allowed).

The supply `Keeper` also introduces new wrapper functions for the auth `Keeper`
and the bank `Keeper` that are related to `ModuleAccount`s in order to be able
to:

- Get and set `ModuleAccount`s by providing the `Name`.
- Send coins from and to other `ModuleAccount`s or standard `Account`s
  (`BaseAccount` or `VestingAccount`) by passing only the `Name`.
- `Mint` or `Burn` coins for a `ModuleAccount` (restricted to its permissions).

### Permissions

Each `ModuleAccount` has a different set of permissions that provide different
object capabilities to perform certain actions. Permissions need to be
registered upon the creation of the supply `Keeper` so that every time a
`ModuleAccount` calls the allowed functions, the `Keeper` can lookup the
permissions to that specific account and perform or not the action.

The available permissions are:

- `Minter`: allows for a module to mint a specific amount of coins.
- `Burner`: allows for a module to burn a specific amount of coins.
- `Staking`: allows for a module to delegate and undelegate a specific amount of coins.

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
