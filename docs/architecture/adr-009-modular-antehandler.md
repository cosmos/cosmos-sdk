# ADR 009: Modular AnteHandler

## Changelog

- 2019 Aug 31: Initial draft

## Context

The current AnteHandler design allows users to either use the default AnteHandler provided in `x/auth` or to build their own AnteHandler from scratch. Ideally AnteHandler functionality is split into multiple, modular functions that can be chained together along with custom ante-functions so that users do not have to rewrite common antehandler logic when they want to implement custom behavior.

## Proposals

### Per-Module AnteHandler

One approach is to use the [ModuleManager](https://godoc.org/github.com/cosmos/cosmos-sdk/types/module) and have each module implement its own antehandler if it requires custom antehandler logic. The ModuleManager can then be passed in an AnteHandler order in the same way it has an order for BeginBlockers and EndBlockers. The ModuleManager returns a single AnteHandler function that will take in a tx and run each module's `AnteHandle` in the specified order. The module manager's AnteHandler is set as the baseapp's AnteHandler.

Pros:
1. Simple to implement
2. Utilizes the existing ModuleManager architecture

Cons:
1. Improves granularity but still cannot get more granular than a per-module basis. e.g. If auth's `AnteHandle` function is in charge of validating memo and signatures, users cannot swap the signature-checking functionality while keeping the rest of auth's `AnteHandle` functionality.
2. Module AnteHandler are run one after the other. There is no way for one AnteHandler to wrap or "decorate" another.

### Decorator Pattern

The [weave project](https://github.com/iov-one/weave) achieves AnteHandler modularity through the use of a decorator pattern. The interface is designed as follows:

```golang
// Decorator wraps a Handler to provide common functionality
// like authentication, or fee-handling, to many Handlers
type Decorator interface {
	Check(ctx Context, store KVStore, tx Tx, next Checker) (*CheckResult, error)
	Deliver(ctx Context, store KVStore, tx Tx, next Deliverer) (*DeliverResult, error)
}
```

Each decorator works like a modularized SDK antehandler function, but it can take in a `next` argument that may be another decorator or a Handler (which does not take in a next argument). These decorators can be chained together, one decorator being passed in as the `next` argument of the previous decorator in the chain. The chain ends in a Router which can take a tx and route to the appropriate msg handler.

A key benefit of this approach is that one Decorator can wrap its internal logic around the next Checker/Deliverer. A weave Decorator may do the following:

```golang
// Example Decorator's Deliver function
func (example Decorator) Deliver(ctx Context, store KVStore, tx Tx, next Deliverer) {
    // Do some pre-processing logic

    res, err := next.Deliver(ctx, store, tx)

    // Do some post-processing logic given the result and error
}
```

Pros:
1. Weave Decorators can wrap over the next decorator/handler in the chain. The ability to both pre-process and post-process may be useful in certain settings.
2. Provides a nested modular structure that isn't possible in the solution above, while also allowing for a linear one-after-the-other structure like the solution above.

Cons:
1. It is hard to understand at first glance the state updates that would occur after a Decorator runs given the `ctx`, `store`, and `tx`. A Decorator can have an arbitrary number of nested Decorators being called within its function body, each possibly doing some pre- and post-processing before calling the next decorator on the chain. Thus to understand what a Decorator is doing, one must also understand what every other decorator further along the chain is also doing. This can get quite complicated to understand. A linear, one-after-the-other approach while less powerful, may be much easier to reason about.

### Chained Micro-Functions

The benefit of Weave's approach is that the Decorators can be very concise, which when chained together allows for maximum customizability. However, the nested structure can get quite complex and thus hard to reason about.

Our recommended approach is to split the AnteHandler functionality into tightly scoped "micro-functions", while preserving the one-after-the-other ordering that would come from the ModuleManager approach.

We can then have a way to chain these micro-functions so that they run one after the other. Modules may define multiple ante micro-functions and then also provide a default per-module AnteHandler that implements a default, suggested order for these micro-functions. 

Users can order the AnteHandlers easily by simply using the ModuleManager. The ModuleManager will take in a list of AnteHandlers and return a single AnteHandler that runs each AnteHandler in the order of the list provided. If the user is comfortable with the default ordering of each module, this is as simple as providing a list with each module's antehandler (exactly the same as BeginBlocker and EndBlocker).

If however, users wish to change the order or add, modify, or delete ante micro-functions in anyway; they can always define their own ante micro-functions and add them explicitly to the list that gets passed into module manager.

#### Default Workflow:

##### SDK code:

```golang
func Chainer(order []AnteHandler) AnteHandler {
    return func(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
        for _, ante := range order {
            ctx, err := ante(ctx, tx, simulate)
            if err != nil {
                return ctx, err
            }
        }
        return ctx, err
    }
}
```

```golang
func VerifySignatures(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
    // verify signatures
    // Returns InvalidSignature Result and abort=true if sigs invalid
    // Return OK result and abort=false if sigs are valid
}

func ValidateMemo(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
    // validate memo
}

AuthModuleAnteHandler := Chainer([]AnteHandler{VerifySignatures, ValidateMemo})
```

```golang
func DeductFees(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
    // Deduct fees from tx
    // Abort if insufficient funds in account to pay for fees
}

func CheckMempoolFees(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
    // If CheckTx: Abort if the fees are less than the mempool's minFee parameter
}

DistrModuleAnteHandler := Chainer([]AnteHandler{CheckMempoolFees, DeductFees})
```

```golang
type ModuleManager struct {
    // other fields
    AnteHandlerOrder []AnteHandler
}

func (mm ModuleManager) GetAnteHandler() AnteHandler {
    retun Chainer(mm.AnteHandlerOrder)
}
```

##### User Code:

```golang
moduleManager.SetAnteHandlerOrder([]AnteHandler(AuthModuleAnteHandler, DistrModuleAnteHandler))

app.SetAnteHandler(mm.GetAnteHandler())
```

#### Custom Workflow

##### User Code

```golang
func CustomSigVerify(ctx Context, tx Tx, simulate bool) (newCtx Context, err error) {
    // do some custom signature verification logic
}
```

```golang
// Micro-functions allow users to change order of when they get executed, and swap out default ante-functionality with their own custom logic.
// Note that users can still chain the default distribution module handler, and auth micro-function along with their custom ante function
moduleManager.SetAnteHandlerOrder([]AnteHandler(ValidateMemo, CustomSigVerify, DistrModuleAnteHandler))
```

Pros:
1. Allows for ante functionality to be as modular as possible.
2. For users that do not need custom ante-functionality, there is little difference between how antehandlers work and how BeginBlock and EndBlock work in ModuleManager.
3. Still easy to understand

Cons:
1. Cannot wrap antehandlers with decorators like you can with Weave.

## Status

> Proposed

## Consequences

Since pros and cons are written for each approach, it is omitted from this section

## References

- [#4572](https://github.com/cosmos/cosmos-sdk/issues/4572):  Modular AnteHandler Issue
- [#4582](https://github.com/cosmos/cosmos-sdk/pull/4583): Initial Implementation of Per-Module AnteHandler Approach
- [Weave Decorator Code](https://github.com/iov-one/weave/blob/master/handler.go#L35)
- [Weave Design Videos](https://vimeo.com/showcase/6189877)
