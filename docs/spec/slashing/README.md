# Slashing module specification

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

1. **[Overview](overview.md)**
1. **[State](state.md)**
    1. [SigningInfo](state.md#signing-info)
2. **[Transactions](transactions.md)**
    1. [Unjail](transactions.md#unjail)
3. **[Hooks](hooks.md)**
    1. [Validator Bonded](hooks.md#validator-bonded)
4. **[Begin Block](begin-block.md)**
    1. [Evidence handling](begin-block.md#evidence-handling)
    2. [Uptime tracking](begin-block.md#uptime-tracking)
5. **[Future Improvements](future-improvements.md)**
    1. [State cleanup](future-improvements.md#state-cleanup)
