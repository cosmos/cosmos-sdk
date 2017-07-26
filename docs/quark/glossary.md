# Glossary

This glossary defines many terms used throughout documentation of Quark.  If
there is every a concept that seems unclear, check here. This is mainly to
provide a background and general understanding of the different words and
concepts that are used.  Other documents will explain in more detail how to
combine these concepts to build a particular application.

## Transaction (tx)

A transaction is a packet of binary data that contains all information to
validate and perform an action on the blockchain. The only other data that it
interacts with is the current state of the chain (key-value store), and
it must have a deterministic action. The tx is the main piece of one request.

We currently make heavy use of [go-wire](https://github.com/tendermint/go-wire)
and [data](https://github.com/tendermint/go-wire/tree/master/data) to provide
binary and json encodings and decodings for `struct` or  interface` objects.
Here, encoding and decoding operations are designed to operate with interfaces
nested any amount times (like an onion!). There is one public `TxMapper`
in the basecoin root package, and all modules can register their own transaction types there. This allows us to deserialize the entire tx in
one location (even with types defined in other repos), to easily embed
an arbitrary tx inside another without specifying the type, and provide
an automatic json representation to provide to users (or apps) to
inspect the chain.

Note how we can wrap any other transaction, add a fee level, and not worry
about the encoding in our module any more?

```golang
type Fee struct {
  Fee   coin.Coin      `json:"fee"`
  Payer basecoin.Actor `json:"payer"` // the address who pays the fee
  Tx    basecoin.Tx    `json:"tx"`
}
```

## Context (ctx)

As a request passes through the system, it may pick up information such as the
authorization it has received from another middleware, or the block height the
request runs at.  In order to carry this information between modules it is
saved to the context.  further, it all information must be deterministic from
the context in which the request runs (based on the tx and the block it was
included in) and can be used to validate the tx.

## Data Store

To be able to provide proofs to Tendermint, we keep all data in one key-value
(kv) store which is indexed with a merkle tree.  This allows for the easy
generation of a root hash and proofs for queries without requiring complex
logic inside each module. Standardization of this process also allows powerful
light-client tooling as any store data may be verified on the fly.

The largest limitation of the current implemenation of the kv-store is that
interface that the application must use can only `Get` and `Set` single data
points.  This said, there are some data structures like queues and range
queries that are available in `state` package. These provide higher-level
functionality in a standard format, but have not yet been integrated into the
kv-store interface.

## Isolation

One of the main arguments for blockchain is security.  So while we encourage
the use of third-party modules, all developers must be vigilant against
security holes.  If you use the
[stack](https://github.com/tendermint/basecoin/tree/unstable/stack)
package, it will provide two different types of compartmentalization security.

The first is to limit the working kv-store space of each module. When
`DeliverTx` is called for a module, it is never given the entire data store,
but rather only its own prefixed subset of the store. This is achieved by
prefixing all keys transparently with `<module name> + 0x0`, using the null
byte as a separator.  Since the module name must be a string, no malicious
naming scheme can ever lead to a collision. Inside a module, we can
write using any key value we desire without the possibility that we
have modified data belonging to separate module.

The second is to add permissions to the transaction context.  The tx context
can specify that the tx has been signed by one or multiple specific
[actors](https://github.com/tendermint/basecoin/blob/unstable/context.go#L18).
A tx will only be executed if the permission requirements have been fulfilled.
For example the sender of funds must have signed, or 2 out of 3
multi-signature actors must have signed a joint account.  To prevent the
forgery of account signatures from unintended modules each permission
is associated with the module that granted it (in this case
[auth](https://github.com/tendermint/basecoin/tree/unstable/modules/auth)),
and if a module tries to add a permission for another module, it will
panic.  There is also protection if a module creates a brand new fake
context to trick the downstream modules. Each context enforces
the rules on how to make child contexts, and the stack middleware builder
enforces that the context passed from one level to the next is a valid
child of the original one.

These security measures ensure that modules can confidently write to their
local section of the database and trust the permissions associated with the
context, without concern of interference from other modules.  (Okay,
if you see a bunch of C-code in the module traversing through all the
memory space of the application, then get worried....)

## Handler

The ABCI interface is handled by `app`, which translates these data structures
into an internal format that is more convenient, but unable to travel over the
wire.  The basic interface for any code that modifies state is the `Handler`
interface, which provides four methods:

```golang
  Name() string
  CheckTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
  DeliverTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
  SetOption(l log.Logger, store state.KVStore, module, key, value string) (string, error)
```

Note the `Context`, `KVStore`, and `Tx` as principal carriers of information.
And that Result is always success, and we have a second error return
for errors (which is much more standard golang that `res.IsErr()`)

The `Handler` interface is designed to be the basis for all modules that
execute transactions, and this can provide a large degree of code
interoperability, much like `http.Handler` does in golang web development.

## Middleware

Middleware is a series of processing steps that any request must travel through
before (and after) executing the registered `Handler`.  Some examples are a
logger (that records the time before executing the tx, then outputs info -
including duration - after the execution), of a signature checker (which
unwraps the tx by one layer, verifies signatures, and adds the permissions to
the Context before passing the request along).

In keeping with the standardization of `http.Handler` and inspired by the
super minimal [negroni](https://github.com/urfave/negroni/blob/master/README.md)
package, we just provide one more `Middleware` interface, which has an extra
`next` parameter, and a `Stack` that can wire all the levels together (which
also gives us a place to perform isolation of each step).

```golang
  Name() string
  CheckTx(ctx Context, store state.KVStore, tx Tx, next Checker) (Result, error)
  DeliverTx(ctx Context, store state.KVStore, tx Tx, next Deliver) (Result, error)
  SetOption(l log.Logger, store state.KVStore, module, key, value string, next Optioner) (string, error)
```

## Modules

A module is a set of functionality which should be typically designed as
self-sufficient. Common elements of a module are:

* transaction types (either end transactions, or transaction wrappers)
* custom error codes
* data models (to persist in the kv-store)
* handler (to handle any end transactions)
* middleware (to handler any wrapper transactions)

To enable a module, you must add the appropriate middleware (if any) to the
stack in `main.go` for the client application (Quark default:
`basecli/main.go`), as well as adding the handler (if any) to the dispatcher
(Quark default: `app/app.go`).  Once the stack is compiled into a `Handler`,
then each tx is handled by the appropriate module.

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

```golang
const (
  NameCoin = "coin"
  TypeSend = NameCoin + "/send"
)
```

## Inter-Plugin Communication (IPC)

But wait, there's more... since we have isolated all the modules from each
other, we need to allow some way for them to interact in a controlled fashion.
One example is the `fee` middleware, which wants to deduct coins from the
calling account and can accomplished most easilty with the `coin` module.

If we want to make a call from the middleware, this is relatively simple.  The
middleware already has a handle to the `next` Handler, which will execute the
rest of the stack. It can simple create a new SendTx and pass it down the
stack. If it returns success, then do the rest of the processing (and send the
original tx down the stack), otherwise abort.

However, if one `Handler` inside the `Dispatcher` wants to do this, it becomes
more complex.  The solution is that the `Dispatcher` accepts not a `Handler`,
but a `Dispatchable`, which looks like a middleware, except that the `next`
argument is a callback to the dispatcher to execute a sub-transaction.  If a
module doesn't want to use this functionality, it can just implement `Handler`
and call `stack.WrapHandler(h)` to convert it to a `Dispatchable` that never
uses the callback.

One example of this is the counter app, which can optionally accept a payment.
If the tx contains a payment, it must create a SendTx and pass this to the
dispatcher to deduct the amount from the proper account. Take a look at
[counter
plugin](https://github.com/tendermint/basecoin/blob/unstable/docs/guide/counter/plugins/counter/counter.go)
for a better idea.

## Permissions

IPC requires a more complex permissioning system to allow the modules to have
limited access to each other.  Also to allow more types of permissions than
simple public key signatures. So, rather than just use an address to identify
who is performing an action, we can use a more complex structure:

```golang
type Actor struct {
  ChainID string     `json:"chain"` // this is empty unless it comes from a different chain
  App     string     `json:"app"`   // the app that the actor belongs to
  Address data.Bytes `json:"addr"`  // arbitrary app-specific unique id
}
```

Here, the `Actor` abstracts any address that can authorize actions, hold funds,
or initiate any sort of transaction. It doesn't just have to be a pubkey on
this chain, it could stem from another app (such as multi-sig account), or even
another chain (via IBC)

`ChainID` is to be used for IBC, which is discussed below, but right now focus
on `App` and `Address`.  For a signature, the App is `auth`, and any modules
can check to see if a specific public key address signed like this
`ctx.HasPermission(auth.SigPerm(addr))`.  However, we can also authorize a tx
with `roles`, which handles multi-sig accounts, it checks if there were enough
signatures by checking as above, then it can add the role permission like `ctx
= ctx.WithPermissions(NewPerm(assume.Role))`

In addition to permissioning, the Actors are addresses just like public key
addresses. So one can create a mulit-sig role, then send coin there, which can
only be moved upon meeting the authorization requirements from that module.
`coin` doesn't even know the existence of `roles` and one could build any other
sort of module to provide permissions (like bind the outcome of an election to
move coins or to modify the accounts on a role).

One idea (not implemented) is to provide scopes on the permissions.  Right now,
if I sign a tx to one module, it can pass it on to any other module over IPC
with the same permissions.  It could move coins, vote in an election, or
anything else. Ideally, when signing, one could also specify the scope(s) that
this signature authorizes. The [oauth
protocol](https://api.slack.com/docs/oauth-scopes) also has to deal with a
similar problem, and maybe could provide some inspiration.


## Replay Protection

In order to prevent [replay
attacks](https://en.wikipedia.org/wiki/Replay_attack) a multi account nonce system
has been constructed as a module, which can be found in
`modules/nonce`.  By adding the nonce module to the stack, each
transaction is verified for authenticity against replay attacks. This is
achieved by requiring that a new signed copy of the sequence number which must
be exactly 1 greater than the sequence number of the previous transaction. A
distinct sequence number is assigned per chain-id, application, and group of
signers. Each sequence number is tracked as a nonce-store entry where the key
is the marshaled list of actors after having been sorted by chain, app, and
address.

```golang
// Tx - Nonce transaction structure, contains list of signers and current sequence number
type Tx struct {
	Sequence uint32           `json:"sequence"`
	Signers  []basecoin.Actor `json:"signers"`
	Tx       basecoin.Tx      `json:"tx"`
}
```

By distinguishing sequence numbers across groups of Signers, multi-signature
Actors need not lock up use of their Address while waiting for all the members
of a multi-sig transaction to occur. Instead only the multi-sig account will
be locked, while other accounts belonging to that signer can be used and signed
with other sequence numbers.

By abstracting out the nonce module in the stack, entire series of transactions
can occur without needing to verify the nonce for each member of the series. An
common example is a stack which will send coins and charge a fee. Within Quark
this can be achieved using separate modules in a stack, one to send the coins
and the other to charge the fee, however both modules do not need to check the
nonce. This can occur as a separate module earlier in the stack.

## IBC (Inter-Blockchain Communication)

Wow, this is a big topic.  Also a WIP.  Add more here...
