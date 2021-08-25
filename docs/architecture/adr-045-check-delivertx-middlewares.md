# ADR 045: BaseApp `{Check,Deliver}Tx` as Middlewares

## Changelog

- 20.08.2021: Initial draft.

## Status

ACCEPTED

## Abstract

This ADR replaces the current BaseApp `runTx` and antehandlers design with a middleware-based design.

## Context

BaseApp's implementation of ABCI `{Check,Deliver}Tx()` and its own `Simulate()` method call the `runTx` method under the hood, which first runs antehandlers, then executes `Msg`s. However, the [transaction Tips](https://github.com/cosmos/cosmos-sdk/issues/9406) and [refunding unused gas](https://github.com/cosmos/cosmos-sdk/issues/2150) use cases require custom logic to be run after the `Msg`s execution. There is currently no way to achieve this.

An naive solution would be to add post-`Msg` hooks to BaseApp. However, the SDK team thinks in parallel about the bigger picture of making app wiring simpler ([#9181](https://github.com/cosmos/cosmos-sdk/discussions/9182)), which includes making BaseApp more lightweight and modular.

## Decision

We decide to transform Baseapp's implementation of ABCI `{Check,Deliver}Tx` and its own `Simulate` methods to use a middleware-based design.

The two following interfaces are the base of the middleware design, and are defined in `types/tx`:

```go
type Handler interface {
    CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
    DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
    SimulateTx(ctx context.Context, tx sdk.Tx, req RequestSimulateTx) (ResponseSimulateTx, error)
}

type Middleware func(Handler) Handler
```

BaseApp holds a reference to a `tx.Handler`:

```go
type BaseApp  struct {
    // other fields
    txHandler tx.Handler
}
```

Baseapp's ABCI `{Check,Deliver}Tx()` and `Simulate()` methods simply call `app.txHandler.{Check,Deliver,Simulate}Tx()` with the relevant arguments. For example, for `DeliverTx`:

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

`baseapp.txHandler`'s three methods' implementations can obviously be monolithic functions, but for modularity we propose a middleware composition design, where a middleware is simply a function that takes a `tx.Handler`, and returns another `tx.Handler` wrapped around the previous one.

### Implementing a Middleware

In practice, middlewares are created by Go function that takes as arguments some parameters needed for the middleware, and returns a `tx.Middleware`.

For example, for creating an arbitrary `MyMiddleware`, we can implement:

```go
// myTxHandler is the tx.Handler of this middleware. Note that it holds a
// reference to the next tx.Handler in the stack.
type myTxHandler struct {
    // next is the next tx.Handler in the middleware stack.
    next tx.Handler
    // some other fields that are relevant to the middleware can be added here
}

// NewMyMiddleware returns a middleware that does this and that.
func NewMyMiddleware(arg1, arg2) tx.Middleware {
    return func (txh tx.Handler) tx.Handler {
        return myTxHandler{
            next: txh,
            // optionally, set arg1, arg2... if they are needed in the middleware
        }
    }
}

// Assert myTxHandler is a tx.Handler.
var _ tx.Handler = myTxHandler{}

func (txh myTxHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
    // CheckTx specific pre-processing logic

    // run the next middleware
    res, err := txh.next.CheckTx(ctx, tx, req)

    // CheckTx specific post-processing logic

    return res, err
}

func (txh myTxHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
    // DeliverTx specific pre-processing logic

    // run the next middleware
    res, err := txh.next.DeliverTx(ctx, tx, req)

    // DeliverTx specific post-processing logic

    return res, err
}

func (txh myTxHandler) SimulateTx(ctx context.Context, tx sdk.Tx, req tx.RequestSimulateTx) (abci.ResponseSimulateTx, error) {
    // SimulateTx specific pre-processing logic

    // run the next middleware
    res, err := txh.next.SimulateTx(ctx, tx, req)

    // SimulateTx specific post-processing logic

    return res, err
}
```

### Composing Middlewares

While BaseApp simply holds a reference to a `tx.Handler`, this `tx.Handler` itself is defined using a middleware stack. The SDK exposes a base (i.e. innermost) `tx.Handler` called `RunMsgsTxHandler`, which executes messages.

Then, the app developer can compose multiple middlewares on top on the base `tx.Handler`. Each middleware can run pre-and-post-processing logic around its next middleware, as described in the section above. Conceptually, as an example, given the middlewares `A`, `B`, and `C` and the base `tx.Handler` `H` the stack looks like:

```
A.pre
    B.pre
        C.pre
            H # The base tx.handler, for example `RunMsgsTxHandler`
        C.post
    B.post
A.post
```

We define a `ComposeMiddlewares` function for composing middlewares. It takes the base handler as first argument, and middlewares in the "outer to inner" order. For the above stack, the final `tx.Handler` is:

```go
txHandler := middleware.ComposeMiddlewares(H, A, B, C)
```

The middleware is set in BaseApp via its `SetTxHandler` setter:

```go
// simapp/app.go

txHandler := middleware.ComposeMiddlewares(...)
app.SetTxHandler(txHandler)
```

The app developer can define their own middlewares, or use the SDK's pre-defined middlewares.

### Middlewares Maintained by the SDK

While the app developer can define and compose the middlewares of their choice, the SDK provides a set of middlewares that caters for the ecosystem's most common use cases. These middlewares are:

| Middleware              | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| RunMsgsTxHandler        | This is the base `tx.Handler`. It replaces the old baseapp's `runMsgs`, and executes a transaction's `Msg`s.                                                                                                                                                                                                                                                                                                                                                                             |
| {Antehandlers}          | Each antehandler is converted to its own middleware. These middlewares perform signature verification, fee deductions and other validations on the incoming transaction.                                                                                                                                                                                                                                                                                                                 |
| IndexEventsTxMiddleware | This is a simple middleware that chooses which events to index in Tendermint. Replaces `baseapp.indexEvents` (which unfortunately still exists in baseapp too, because it's used to index Begin/EndBlock events)                                                                                                                                                                                                                                                                         |
| RecoveryTxMiddleware    | This index recovers from panics. It replaces baseapp.runTx's panic recovery described in [ADR-022](./adr-022-custom-panic-handling.md).                                                                                                                                                                                                                                                                                                                                                  |
| GasTxMiddleware         | This replaces the [`Setup`](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/setup.go) Antehandler. It sets a GasMeter on sdk.Context. Note that before, GasMeter was set on sdk.Context inside the antehandlers, and there was some mess around the fact that antehandlers had their own panic recovery system so that the GasMeter could be read by baseapp's recovery system. Now, this mess is all removed: one middleware sets GasMeter, another one handles recovery. |

### Similarities and Differences between Antehandlers and Middlewares

The middleware-based design builds upon the existing antehandlers design described in [ADR-010](./adr-010-modular-antehandler.md). Even though the final decision of ADR-010 was to go with the "Simple Decorators" approach, the middleware design is actually very similar to the other [Decorator Pattern](./adr-010-modular-antehandler.md#decorator-pattern) proposal, also used in [weave](https://github.com/iov-one/weave).

#### Similarities with Antehandlers

- Designed as chaining/composing small modular pieces.
- Allow code reuse for `{Check,Deliver}Tx` and for `Simulate`.
- Set up in `app.go`, and easily customizable by app developers.

#### Differences with Antehandlers

- The Antehandlers are run before `Msg` execution, whereas middlewares can run before and after.
- The middleware approach uses separate methods for `{Check,Deliver,Simulate}Tx`, whereas the antehandlers pass a `simulate bool` flag and uses the `sdkCtx.Is{Check,Recheck}Tx()` flags to determine in which transaction mode we are.
- The middleware design lets each middleware hold a reference to the next middleware, whereas the antehandlers pass a `next` argument in the `AnteHandle` method.
- The middleware design use Go's standard `context.Context`, whereas the antehandlers use `sdk.Context`.

## Consequences

### Backwards Compatibility

Since this refactor removes some logic away from BaseApp and into middlewares, it introduces API-breaking changes for app developers. Most notably, instead of creating an antehandler chain in `app.go`, app developers need to create a middleware stack:

```diff
- anteHandler, err := ante.NewAnteHandler(
-    ante.HandlerOptions{
-        AccountKeeper:   app.AccountKeeper,
-        BankKeeper:      app.BankKeeper,
-        SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
-        FeegrantKeeper:  app.FeeGrantKeeper,
-        SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
-    },
-)
+txHandler, err := authmiddleware.NewDefaultTxHandler(authmiddleware.TxHandlerOptions{
+    Debug:             app.Trace(),
+    IndexEvents:       indexEvents,
+    LegacyRouter:      app.legacyRouter,
+    MsgServiceRouter:  app.msgSvcRouter,
+    LegacyAnteHandler: anteHandler,
+})
if err != nil {
    panic(err)
}
- app.SetAnteHandler(anteHandler)
+ app.SetTxHandler()
```

As usual, the SDK will provide a migration document for app developers.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

- [#9934](https://github.com/cosmos/cosmos-sdk/discussions/9934) Decomposing BaseApp's other ABCI methods into middlewares.

## Test Cases

We update the existing baseapp and antehandlers tests to the new middleware API, but keep the same test cases, to avoid introducing regressions. Existing CLI tests will also not be touched.

For new middlewares, we introduce unit tests. Since middlewares are purposefully small, unit tests suit well.

## References

- Initial discussion: https://github.com/cosmos/cosmos-sdk/issues/9585
- Implementation: https://github.com/cosmos/cosmos-sdk/pull/9920
