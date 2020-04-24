# ADR 022: Custom baseapp panic handling

## Changelog

- 2020 Apr 24: Initial Draft

## Context

Current implementation of BaseApp does not allow developers to write custom error handlers in
[runTx()](https://github.com/cosmos/cosmos-sdk/blob/bad4ca75f58b182f600396ca350ad844c18fc80b/baseapp/baseapp.go#L538)
method. We think that this method can be more flexible and can give SDK users more options for customizations without
the need to rewrite whole BaseApp. Also there's one special case for `sdk.ErrorOutOfGas` error which feels like dirty
hack in non-flexible environment.

We propose middleware-solution, which could help developers implement following cases:
* add external logging (let's say sending reports to external services like Sentry);
* call panic for specific error cases;

It will also make `OutOfGas` case and `default` case one of the middlewares.

## Decision

### Design

Instead of hardcoding custom error handling into BaseApp we suggest using set of middlewares which can be customized
externally and will allow developers use as many custom error handlers as they want.

Implementation is already proposed [here](https://github.com/cosmos/cosmos-sdk/pull/6053).

## Status

Proposed

## Consequences

### Positive

- Developers of Cosmos SDK based projects can add custom panic handlers to:
    * add error context for custom panic sources (panic inside of custom keepers);
    * emit `panic()`: passthrough recovery object to the Tendermint core;
    * other necessary handling;
- Developers can use standard Cosmos SDK `baseapp` implementation, rather that rewriting it in their projects;
- Proposed solution doesn't break the current "standard" `runTx()` flow;

### Negative

- Introduces changes to the execution model design.

### Neutral

- `OutOfGas` error handler becomes one of the middlewares;
- Default panic handler becomes one of the middlewares;

## References

- [PR-6053 with proposed solution](https://github.com/cosmos/cosmos-sdk/pull/6053)
