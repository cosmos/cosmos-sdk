# Message Handling

## Context

The SDK uses a `Context` to propogate common information across functions. The
`Context` is modeled after the Golang `context.Context` object, which has
become ubiquitous in networking middleware and routing applications as a means
to easily propogate request context through handler functions.

The main information stored in the `Context` includes the application
MultiStore, the last block header, and the transaction bytes.
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

While the context holds the entire application state (ie. the 
MultiStore), handlers are restricted in what they can do based on the
capabilities they were given when the application was set up.

For instance, suppose we have a `newFooHandler`: 

```go
func newFooHandler(key sdk.StoreKey) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
        store := ctx.KVStore(key)
        // ...
    }
}
```

This handler can only access one store based on whichever key its given.
So when we register the handler for the `foo` message type, we make sure 
to give it the `fooKey`:

```
app.Router().AddRoute("foo", newFooHandler(fooKey))
```

Now it can only access the `foo` store, but not the `bar` or `cat` stores!
