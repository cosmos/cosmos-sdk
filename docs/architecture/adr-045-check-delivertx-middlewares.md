# ADR 045: BaseApp `{Check,Deliver}Tx` as Middlewares

## Changelog

- 20.08.2021: Initial draft.

## Status

PROPOSED

## Abstract

This ADR replaces the current BaseApp `runTx` and antehandlers design with a middleware-based design.

## Context

BaseApp's implementation of ABCI `{Check,Deliver}Tx()` and its own `Simulate()` method call the `runTx` method under the hood, which first runs antehandlers, then executes `Msg`s. However, the [transaction Tips](https://github.com/cosmos/cosmos-sdk/issues/9406) feature requires custom logic to be run after the `Msg`s execution. There is currently no way to achieve this.

An naive solution would be to add post-`Msg` hooks to BaseApp. However, the SDK team thinks in parallel about the bigger picture of making app wiring simpler ([#9181](https://github.com/cosmos/cosmos-sdk/discussions/9182)), which includes making BaseApp more lightweight and modular.

## Decision

We decide to transform Baseapp's implementation of ABCI `{Check,Deliver}Tx()` and its own `Simulate()` method to use a middleware-based design.

The two following interfaces are the base of the middleware design, and are defined in `types/tx`:

```go
type Handler interface {
	CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
	SimulateTx(ctx context.Context, tx sdk.Tx, req RequestSimulateTx) (ResponseSimulateTx, error)
}

type Middleware func(Handler) Handler
```

BaseApp holds a reference to a `tx.Handler`, and the ABCI `{Check,Deliver}Tx()` and `Simulate()` methods simply call `app.txHandler.{Check,Deliver,Simulate}Tx()` with the relevant arguments. For example, for `DeliverTx`:

```go
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
    tx, err := app.txDecoder(req.Tx)
    if err != nil {
        return sdkerrors.ResponseDeliverTx(err, 0, 0, app.trace)
    }

    ctx := app.getContextForTx(runTxModeDeliver, req.Tx)
    res, err := app.txHandler.DeliverTx(ctx, tx, req)
    if err != nil {
        return sdkerrors.ResponseDeliverTx(err, uint64(res.GasUsed), uint64(res.GasWanted), app.trace)
    }

    return res
}
```

The implementations are similar for `BaseApp.CheckTx` and `BaseApp.Simulate`.

### Composing Middlewares

While BaseApp simply holds a reference to a `tx.Handler`, this `tx.Handler` itself is defined using a middleware stack. We define a base `tx.Handler` called `RunMsgsTxHandler`, which executes messages.

Then, the app developer can compose multiple middlewares on top on the base `tx.Handler`. Each middleware can run pre-and-post-processing logic around its next middleware. Conceptually, as an example, given the middlewares `A`, `B`, and `C` and the base `tx.Handler` `H` the stack looks like:

```
A.pre
    B.pre
        C.pre
            H // The base handler, usually `RunMsgsTxHandler`
        C.post
    B.post
A.post
```

We define a `ComposeMiddlewares` function for composing middlewares. It takes the base handler as first argument, and middlewares in the "inner to outer" order. For the above stack, the final `tx.Handler` is:

```go
txHandler := middleware.ComposeMiddlewares(H, C, B, A)
```

The middleware is set in BaseApp via its `SetTxHandler` setter:

```go
// simapp/app.go

txHandler := middleware.ComposeMiddlewares(...)
app.SetTxHandler(txHandler)
```

Naturally, the app developer can define their own middlewares.

### Middlewares Maintained by the SDK

While the app developer can define and compose the middlewares of their choice, the SDK provides a set of middlewares that caters for the ecosystem's most common use cases. These middlewares are:

| Middleware              | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| RunMsgsTxHandler        | This is the base `tx.Handler`. It replaces the old baseapp's `runMsgs`, and executes a transaction's `Msg`s.                                                                                                                                                                                                                                                                                                                                                                             |
| {Antehandlers}          | Each antehandler is converted to its own middleware.                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| IndexEventsTxMiddleware | This is a simple middleware that chooses which events to index in Tendermint. Replaces `baseapp.indexEvents` (which unfortunately still exists in baseapp too, because it's used to index Begin/EndBlock events)                                                                                                                                                                                                                                                                         |
| RecoveryTxMiddleware    | This index recovers from panics. It replaces baseapp.runTx's panic recovery.                                                                                                                                                                                                                                                                                                                                                                                                             |
| GasTxMiddleware         | This replaces the [`Setup`](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/setup.go) Antehandler. It sets a GasMeter on sdk.Context. Note that before, GasMeter was set on sdk.Context inside the antehandlers, and there was some mess around the fact that antehandlers had their own panic recovery system so that the GasMeter could be read by baseapp's recovery system. Now, this mess is all removed: one middleware sets GasMeter, another one handles recovery. |

### Similarities and Differences between Antehandlers and Middlewares

The middleware-based design builds upon the existing antehandlers design described in [ADR-010](./adr-010-modular-antehandler.md).

#### Similarities

- Both design are based on chaining/composing small modular pieces.
-

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

```

```
