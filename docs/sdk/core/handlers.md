
## Context

The SDK uses a `Context` to propogate common information across functions. The
`Context` is modeled after the Golang `context.Context` object, which has
become ubiquitous in networking middleware and routing applications as a means
to easily propogate request context through handler functions.

The main information stored in the `Context` includes the application
MultiStore (see below), the last block header, and the transaction bytes.
Effectively, the context contains all data that may be necessary for processing
a transaction.

Many methods on SDK objects receive a context as the first argument. 

## Handler

Message processing in the SDK is defined through `Handler` functions:

```go
type Handler func(ctx Context, msg Msg) Result
```

A handler takes a context and a message and returns a result.  All
information necessary for processing a message should be available in the
context.

While the context holds the entire application state (all referenced from the
root MultiStore), a particular handler only needs a particular kind of access
to a particular store (or two or more). Access to stores is managed using
capabilities keys and mappers.  When a handler is initialized, it is passed a
key or mapper that gives it access to the relevant stores.

```go
// File: cosmos-sdk/examples/basecoin/app/init_stores.go
app.BaseApp.MountStore(app.capKeyMainStore, sdk.StoreTypeIAVL)
app.accountMapper = auth.NewAccountMapper(
	app.capKeyMainStore, // target store
	&types.AppAccount{}, // prototype
)

// File: cosmos-sdk/examples/basecoin/app/init_handlers.go
app.router.AddRoute("bank", bank.NewHandler(app.accountMapper))

// File: cosmos-sdk/x/bank/handler.go
// NOTE: Technically, NewHandler only needs a CoinMapper
func NewHandler(am sdk.AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		cm := CoinMapper{am}
		...
	}
}
```

