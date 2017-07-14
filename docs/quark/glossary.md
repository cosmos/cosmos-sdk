# Glossary

This defines many of the terms that are used in the other documents.  If there
is every a concept that seems unclear, check here. This is mainly to provide
a background and general understanding of the different words and concepts
that are used.  Other documents will explain in more detail how to combine
these concepts to build a particular application.

## Transaction

A transaction is a packet of binary data that contains all information to
validate and perform an action on the blockchain. The only other data that
it interacts with is the current state of the chain (kv store), and it must
have a deterministic action. The transaction is the main piece of one request.

We currently make heavy use of go-wire and go-data to provide automatic binary
and json encodings (and decodings) for objects, even when they embed many
interfaces inside. There is one public `TxMapper` in the basecoin root package,
and all modules can register their own transaction types there. This allows us
to deserialize the entire tx in one location (even with types defined in other
repos), to easily embed an arbitrary Tx inside another without specifying
the specific type, and provide an automatic json representation to provide to
users (or apps) to inspect the chain.

Note how we can wrap any other transaction, add a fee level, and not worry
about the encoding in our module any more?

```Go
type Fee struct {
  Fee   coin.Coin      `json:"fee"`
  Payer basecoin.Actor `json:"payer"` // the address who pays the fee
  Tx    basecoin.Tx    `json:"tx"`
}
```

## Context

As the request passes through the system, it can pick up information, that must
be carried along with it.  Like the authorized it has received from another
middleware, or the block height it runs at.  This is all deterministic
information from the context in which the request runs (based on the tx and
the block it was included in) and can be used to validate the tx.

## Data Store

To be able to provide proofs to tendermint, we keep all data in one key-value
store, indexed with a merkle tree.  This allows us to easily provide a root
hash and proofs for queries without requiring complex logic inside each
module. Standarizing this also allows powerful light-client tooling as it knows
how to verify all data in the store.

The downside is there is one quite simple interface that the application has
to `Get` and `Set` data. There is not even a range query.  Although there are
some data structures like queues and range queries that are also in the `state`
package to provide higher-level functionality in a standard format.

## Isolation

One of the main arguments for blockchain is security.  So while we encourage
the use of third-party modules, we must be vigilant against security holes.
If you use the `stack` package, it will provide two different types of
sandboxing for you.

The first step, is that when `DeliverTx` is called on a module, it is never
given the entire data store, but rather only its own prefixed section. This
is achieved by prefixing all keys transparently with `<module name> + 0x0`,
using the null byte as a separator.  Since module name must be a string, no
clever naming scheme can lead to a collision. Inside the module, we can write
anywhere we want, without worry that we have to touch some data that is not ours.

The second step involves the permissions in the context.  The context can say
that this tx was signed by eg. Rigel.  But if any module can add that permission,
it would be too easy to forge accounts.  Thus, each permission is associated
with the module that granted it (in this case `auth`), and if a module tries
to add a permission for another module, it will panic. There is also
protection if a module creates a brand new fake context to trick the downstream
modules.

This means that modules can confidently write to their local section of the
database and trust the permissions associated with the context, without concern
of interferance from other modules.  (Okay, if you see a bunch of C-code in
the module traversing through all the memory space of the application, then
get worried....)

## Handler

The ABCI interface is handled by `app`, which translates these data structures
into an internal format that is more convenient, but unable to travel over the
wire.  The basic interface for any code that modifies state is the `Handler`
interface, which provides four methods:

```Go
  Name() string
  CheckTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
  DeliverTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
  SetOption(l log.Logger, store state.KVStore, module, key, value string) (string, error)
```

Note the `Context`, `Store`, and `Tx` as principal carriers of information. And
that Result is always success, and we have a second error return for errors
(which is much more standard go that `res.IsErr()`)

The `Handler` interface is designed to be the basis for all modules that
execute transaction, and this can provide a large degree of code
interoperability, much like `http.Handler` does in golang web development.

## Middleware

Middleware is a series of processing steps that any request must travel through
before (and after) executing the registered `Handler`.  Some examples are a
logger (that records the time before executing the tx, then outputs info -
including duration - after the execution), of a signature checker (which
unwraps the tx by one layer, verifies signatutes, and adds the permissions to
the Context before passing the request along).

In keeping with the standardazation of `http.Handler` and inspired by the
super minimal [negroni](https://github.com/urfave/negroni/blob/master/README.md)
package, we just provide one more `Middleware` interface, which has an extra
`next` parameter, and a `Stack` that can wire all the levels together (which
also gives us a place to perform isolation of each step).

```Go
  Name() string
  CheckTx(ctx Context, store state.KVStore, tx Tx, next Checker) (Result, error)
  DeliverTx(ctx Context, store state.KVStore, tx Tx, next Deliver) (Result, error)
  SetOption(l log.Logger, store state.KVStore, module, key, value string, next Optioner) (string, error)
```

## Modules

A module is a set of functionality that is more or less self-sufficient. It
usually contains the following pieces:

* transaction types (either end transactions, or transaction wrappers)
* custom error codes
* data models (to persist in the kv store)
* handler (to handle any end transactions)
* middleware (to handler any wrapper transactions)

To enable a module, you must add the appropriate middleware (if any) to the
stack in main.go, as well as adding the handler (if any) to the dispatcher.
One the stack is compiled into a `Handler`, then all tx are handled by the
proper module.

## Dispatcher

We usually will want to have multiple modules working together, and need to
make sure the correct transactions get to the correct module. So we have have
`coin` sending money, `roles` creating multi-sig accounts, and `ibc` following
other chains all working together without interference.

After the chain of middleware, we can register a `Dispatcher`, which also
implements the `Handler` interface.  We then register a list of modules with
the dispatcher. Every module has a unique `Name()`, which is used for
isolating its state space.  We use this same name for routing tx.  Each tx
implementation must be registed with go-wire via `TxMapper`, so we just look
at the registered name of this tx, which should be of the form
`<module name>/xxx`. The dispatcher grabs the appropriate module name from
 the tx name and routes it if the module is present.

This all seems a bit of magic, but really just making use of the other magic
(go-wire) that we are already using, rather than add another layer. The only
thing you need to remember is to use the following pattern, then all the tx
will be properly routed:

```Go
const (
  NameCoin = "coin"
  TypeSend = NameCoin + "/send"
)
```

## IPC (Inter-Plugin Communication)

But wait, there's more... since we have isolated all the modules from each
other, we need to allow some way for them to interact in a controlled fashion.
Some examples are the `fee` middleware, which wants to deduct coins from
the calling account (in the `coin` module), or a vote that requires a payment.

If we want to make a call from the middleware, this is relatively simple.
The middleware already has a handle to the `next` Handler, which will
execute the rest of the stack. It can simple create a new SendTx and pass
it down the stack. If it returns success, then do the rest of the processing
(and send the original tx down the stack), otherwise abort.

However, if one `Handler` inside the `Dispatcher` wants to do this, it
becomes more complex.  The solution is that the `Dispatcher` accepts not
a `Handler`, but a `Dispatchable`, which looks like a middleware, except
that the `next` argument is a callback to the dispatcher to execute a
sub-transaction.  If a module doesn't want to use this functionality,
it can just implement `Handler` and call `stack.WrapHandler(h)` to convert
it to a `Dispatchable` that never uses the callback.

One example of this is the counter app, which can optionally accept a payment.
If the tx contains a payment, it must create a SendTx and pass this to the
dispatcher to deduct the amount from the proper account. Take a look at
[counter plugin](https://github.com/tendermint/basecoin/blob/unstable/docs/guide/counter/plugins/counter/counter.go) for a better idea.

## Permissions

This system requires a more complex permissioning system to allow the modules
to have limited access to each other.  Also to allow more types of permissions
than simple public key signatures. So, rather than just use an address to
identify who is performing an action, we can use a more complex structure:

```Go
type Actor struct {
  ChainID string     `json:"chain"` // this is empty unless it comes from a different chain
  App     string     `json:"app"`   // the app that the actor belongs to
  Address data.Bytes `json:"addr"`  // arbitrary app-specific unique id
}
```

`ChainID` is to be used for IBC, which is discussed below, but right now focus
on `App` and `Address`.  For a signature, the App is `auth`, and any modules can
check to see if a specific public key address signed like this
`ctx.HasPermission(auth.SigPerm(addr))`.  However, we can also authorize a
tx with `roles`, which handles multi-sig accounts, it checks if there were
enough signatures by checking as above, then it can add the role permission like
`ctx = ctx.WithPermissions(NewPerm(assume.Role))`

In addition to permissioning, the Actors are addresses just like public key
addresses. So one can create a mulit-sig role, then send coin there, which
can only be moved upon meeting the authorization requirements from that module.
`coin` doesn't even know the existence of `roles` and one could build any
other sort of module to provide permissions (like bind the outcome of an
election to move coins or to modify the accounts on a role).

One idea (not implemented) is to provide scopes on the permissions.  Right now,
if I sign a tx to one module, it can pass it on to any other module over IPC
with the same permissions.  It could move coins, vote in an election, or
anything else. Ideally, when signing, one could also specify the scope(s) that
this signature authorizes. The [oauth protocol](https://api.slack.com/docs/oauth-scopes)
also has to deal with a similar problem, and maybe could provide some inspiration.

## Replay Protection

Is implemented as middleware.  Rigel can add more info here.  Or look
at [the github issue](https://github.com/tendermint/basecoin/issues/160)

## IBC (Inter-Blockchain Communication)

Wow, this is a big topic.  Also a WIP.  Add more here...
