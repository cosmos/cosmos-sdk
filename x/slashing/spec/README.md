<!--
order: 20
title: Slashing Overview
parent:
  title: "slashing"
-->

# `x/slashing`

## Abstract

This section specifies the slashing module of the Cosmos SDK, which implements functionality
first outlined in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper) in June 2016.

The slashing module enables Cosmos SDK-based blockchains to disincentivize any attributable action
by a protocol-recognized actor with value at stake by penalizing them ("slashing").

Penalties may include, but are not limited to:

- Burning some amount of their stake
- Removing their ability to vote on future blocks for a period of time.

This module will be used by the Cosmos Hub, the first hub in the Cosmos ecosystem.

## Contents

1. **[Concepts](01_concepts.md)**
   - [States](01_concepts.md#states)
   - [Tombstone Caps](01_concepts.md#tombstone-caps)
   - [ASCII timelines](01_concepts.md#ascii-timelines)
2. **[State](02_state.md)**
   - [Signing Info](02_state.md#signing-info)
3. **[Messages](03_messages.md)**
   - [Unjail](03_messages.md#unjail)
4. **[Begin-Block](04_begin_block.md)**
   - [Evidence handling](04_begin_block.md#evidence-handling)
   - [Uptime tracking](04_begin_block.md#uptime-tracking)
5. **[05_hooks.md](05_hooks.md)**
   - [Hooks](05_hooks.md#hooks)
6. **[Events](06_events.md)**
   - [BeginBlocker](06_events.md#beginblocker)
   - [Handlers](06_events.md#handlers)
7. **[Staking Tombstone](07_tombstone.md)**
   - [Abstract](07_tombstone.md#abstract)
8. **[Parameters](08_params.md)**
