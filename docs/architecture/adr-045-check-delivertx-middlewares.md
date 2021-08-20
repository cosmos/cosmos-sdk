# ADR 045: BaseApp `{Check,Deliver}Tx` as Middlewares

## Changelog

- 20.08.2021: Initial draft.

## Status

PROPOSED

## Abstract

This ADR replaces the current BaseApp `runTx` and antehandlers design (partially described in [ADR-10](./adr-010-modular-antehandler.md)) with a middleware-based design.

## Context

BaseApp's implementation of ABCI `{Check,Deliver}Tx()` and its own `Simulate()` method call the `runTx` method under the hood, which first runs antehandlers, then executes `Msg`s. However, the [transaction Tips](https://github.com/cosmos/cosmos-sdk/issues/9406) feature requires custom logic to be run after the `Msg`s execution. There is currently no way to achieve this.

An naive solution would be to add post-`Msg` hooks to BaseApp. However, the SDK team thinks in parallel about the bigger picture of making app wiring simpler ([#9181](https://github.com/cosmos/cosmos-sdk/discussions/9182)), which includes making BaseApp more lightweight and modular.

## Decision

We decide to transform Baseapp's implementation of ABCI `{Check,Deliver}Tx()` and its own `Simulate()` method to use a middleware-based design.

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

- [#9934](https://github.com/cosmos/cosmos-sdk/discussions/9934) Decomposing BaseApp's other ABCI methods into middlewares.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

- {reference link}
