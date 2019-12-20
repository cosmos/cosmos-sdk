<!--
order: 0
title: Upgrade Overview
parent:
  title: "upgrade"
-->

# `upgrade`

## Abstract

`x/upgrade` is an implementation of a Cosmos SDK module that facilitates smoothly
upgrading a live Cosmos chain to a new (breaking) software version. It accomplishes this by
providing a `BeginBlocker` hook that prevents the blockchain state machine from
proceeding once a pre-defined upgrade block time or height has been reached.

The module does not prescribe anything regarding how governance decides to do an
upgrade, but just the mechanism for coordinating the upgrade safely. Without software
support for upgrades, upgrading a live chain is risky because all of the validators
need to pause their state machines at exactly the same point in the process. If
this is not done correctly, there can be state inconsistencies which are hard to
recover from.

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Events](03_events.md)**
