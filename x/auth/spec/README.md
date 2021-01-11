<!--
order: 0
title: "Auth Overview"
parent:
  title: "auth"
-->

# `auth`

## Abstract

This document specifies the auth module of the Cosmos SDK.

The auth module is responsible for specifying the base transaction and account types
for an application, since the SDK itself is agnostic to these particulars. It contains
the ante handler, where all basic transaction validity checks (signatures, nonces, auxiliary fields)
are performed, and exposes the account keeper, which allows other modules to read, write, and modify accounts.

This module will be used in the Cosmos Hub.

## Contents

1. **[Concepts](01_concepts.md)**
   - [Gas & Fees](01_concepts.md#gas-&-fees)
2. **[State](02_state.md)**
   - [Accounts](02_state.md#accounts)
3. **[AnteHandlers](03_antehandlers.md)**
   - [Handlers](03_antehandlers.md#handlers)
4. **[Keepers](04_keepers.md)**
   - [Account Keeper](04_keepers.md#account-keeper)
5. **[Vesting](05_vesting.md)**
   - [Intro and Requirements](05_vesting.md#intro-and-requirements)
   - [Vesting Account Types](05_vesting.md#vesting-account-types)
   - [Vesting Account Specification](05_vesting.md#vesting-account-specification)
   - [Keepers & Handlers](05_vesting.md#keepers-&-handlers)
   - [Genesis Initialization](05_vesting.md#genesis-initialization)
   - [Examples](05_vesting.md#examples)
   - [Glossary](05_vesting.md#glossary)
6. **[Parameters](07_params.md)**
