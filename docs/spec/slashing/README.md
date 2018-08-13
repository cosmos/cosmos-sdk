# Slashing module specification

## Abstract

This section specifies the slashing module of the Cosmos SDK, which implements functionality
first outlined in the [Cosmos Whitepaper](https://cosmos.network/about/whitepaper) in June 2016.

The slashing module enables Cosmos SDK-based blockchains to disincentivize any attributable action
by a protocol-recognized actor with value at stake by "slashing" them: burning some amount of their
stake - and possibly also removing their ability to vote on future blocks for a period of time.

This module will be used by the Cosmos Hub, the first hub in the Cosmos ecosystem.

## Contents

1. **[State](state.md)**
    1. [SigningInfo](state.md#signing-info)
    1. [SlashingPeriod](state.md#slashing-period)
1. **[State Machine](state-machine.md)**
    1. [Transactions](state-machine.md#transactions)
          1. Unjail
    1. [Interactions](state-machine.md#interactions)
          1. Validator Bonded
          1. Validator Unbonding
          1. Validator Slashed
    1. [State Cleanup](state-machine.md#state-cleanup)
1. **[Begin Block](begin-block.md)**
    1. [Evidence handling](begin-block.md#evidence-handling)
    1. [Uptime tracking](begin-block.md#uptime-tracking)
