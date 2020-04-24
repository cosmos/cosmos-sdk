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

#### Overview

Instead of hardcoding custom error handling into BaseApp we suggest using set of middlewares which can be customized
externally and will allow developers use as many custom error handlers as they want. Implementation with tests
can be found [here](https://github.com/cosmos/cosmos-sdk/pull/6053).

#### Implementation details

##### RecoveryHandler

We add a `recover()` object handler type:

```go
type RecoveryHandler func(recoveryObj interface{}) error
```

This handler receives an object, processes it and returns contextually enriched `error` or `nil`
if object is not a handle's target type.

An example:

```go
func(recoveryObj interface{}) error {
    err, ok := recoveryObj.(error)
    if !ok { return nil }
    
    if someSpecificError.Is(err) {
        panic(customPanicMsg)
    } else {
        return nil
    }
}
```

This example breaks the application execution, but it also might enrich the error's context like the `OutOfGas` handler.

##### RecoveryMiddleware

We also add a middleware type:

```go
type recoveryMiddleware func(recoveryObj interface{}) (recoveryMiddleware, error)
```

Function receives a `recover()` object and returns (`next middleware`, `nil`) if object wasn't handled (not a target type)
or (`nil`, `error`) if input object was handled and other middlewares in the chain should not be executed.

`OutOfGas` middleware example:
```go
func newRecoveryMiddleware(handler RecoveryHandler, next recoveryMiddleware) recoveryMiddleware {
    return func(recoveryObj interface{}) (recoveryMiddleware, error) {
        if err := handler(recoveryObj); err != nil {
            return nil, err
        }
        return next, nil
    }
}

func newOutOfGasRecoveryMiddleware(gasWanted uint64, ctx sdk.Context, next recoveryMiddleware) recoveryMiddleware {
    handler := func(recoveryObj interface{}) error {
        err, ok := recoveryObj.(sdk.ErrorOutOfGas)
        if !ok { return nil }

        return sdkerrors.Wrap(
            sdkerrors.ErrOutOfGas, fmt.Sprintf(
                "out of gas in location: %v; gasWanted: %d, gasUsed: %d", err.Descriptor, gasWanted, ctx.GasMeter().GasConsumed(),
            ),
        )
    }
    
    return newRecoveryMiddleware(handler, next)
}
```

##### Recovery processing

Basic chain of middlewares processing would look like:

```go
func processRecovery(recoveryObj interface{}, middleware recoveryMiddleware) error {
	if middleware == nil { return nil }

	next, err := middleware(recoveryObj)
	if err != nil { return err }

	if next != nil { return processRecovery(recoveryObj, next) }

	return nil
}
```

That way we can create a middleware chain which is executed from left to right, the rightmost middleware is a
`default` handler which must return an `error`.

##### Baseapp changes

The default middleware chain must exist in a `baseapp` object and developers can add their custom `RecoveryHandler`s:

```go
func (app *BaseApp) AddRunTxRecoveryHandler(handlers ...RecoveryHandler)
```

This method would prepend handlers to an existing chain.

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
