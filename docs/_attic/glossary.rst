Glossary
========

This glossary defines many terms used throughout documentation of Quark.
If there is every a concept that seems unclear, check here. This is
mainly to provide a background and general understanding of the
different words and concepts that are used. Other documents will explain
in more detail how to combine these concepts to build a particular
application.

Transaction
-----------

A transaction is a packet of binary data that contains all information
to validate and perform an action on the blockchain. The only other data
that it interacts with is the current state of the chain (key-value
store), and it must have a deterministic action. The transaction is the
main piece of one request.

We currently make heavy use of
`go-amino <https://github.com/tendermint/go-amino>`__ to
provide binary and json encodings and decodings for ``struct`` or
interface\ ``objects. Here, encoding and decoding operations are designed to operate with interfaces nested any amount times (like an onion!). There is one public``\ TxMapper\`
in the basecoin root package, and all modules can register their own
transaction types there. This allows us to deserialize the entire
transaction in one location (even with types defined in other repos), to
easily embed an arbitrary transaction inside another without specifying
the type, and provide an automatic json representation allowing for
users (or apps) to inspect the chain.

Note how we can wrap any other transaction, add a fee level, and not
worry about the encoding in our module any more?

::

    type Fee struct {
      Fee   coin.Coin      `json:"fee"`
      Payer basecoin.Actor `json:"payer"` // the address who pays the fee
      Tx    basecoin.Tx    `json:"tx"`
    }

Context (ctx)
-------------

As a request passes through the system, it may pick up information such
as the block height the request runs at. In order to carry this information
between modules it is saved to the context. Further, all information
must be deterministic from the context in which the request runs (based
on the transaction and the block it was included in) and can be used to
validate the transaction.

Data Store
----------

In order to provide proofs to Tendermint, we keep all data in one
key-value (kv) store which is indexed with a merkle tree. This allows
for the easy generation of a root hash and proofs for queries without
requiring complex logic inside each module. Standardization of this
process also allows powerful light-client tooling as any store data may
be verified on the fly.

The largest limitation of the current implemenation of the kv-store is
that interface that the application must use can only ``Get`` and
``Set`` single data points. That said, there are some data structures
like queues and range queries that are available in ``state`` package.
These provide higher-level functionality in a standard format, but have
not yet been integrated into the kv-store interface.

Isolation
---------

One of the main arguments for blockchain is security. So while we
encourage the use of third-party modules, all developers must be
vigilant against security holes. If you use the
`stack <https://github.com/cosmos/cosmos-sdk/tree/master/stack>`__
package, it will provide two different types of compartmentalization
security.

The first is to limit the working kv-store space of each module. When
``DeliverTx`` is called for a module, it is never given the entire data
store, but rather only its own prefixed subset of the store. This is
achieved by prefixing all keys transparently with
``<module name> + 0x0``, using the null byte as a separator. Since the
module name must be a string, no malicious naming scheme can ever lead
to a collision. Inside a module, we can write using any key value we
desire without the possibility that we have modified data belonging to
separate module.

The second is to add permissions to the transaction context. The
transaction context can specify that the tx has been signed by one or
multiple specific actors.

A transactions will only be executed if the permission requirements have
been fulfilled. For example the sender of funds must have signed, or 2
out of 3 multi-signature actors must have signed a joint account. To
prevent the forgery of account signatures from unintended modules each
permission is associated with the module that granted it (in this case
`auth <https://github.com/cosmos/cosmos-sdk/tree/master/x/auth>`__),
and if a module tries to add a permission for another module, it will
panic. There is also protection if a module creates a brand new fake
context to trick the downstream modules. Each context enforces the rules
on how to make child contexts, and the stack builder enforces
that the context passed from one level to the next is a valid child of
the original one.

These security measures ensure that modules can confidently write to
their local section of the database and trust the permissions associated
with the context, without concern of interference from other modules.
(Okay, if you see a bunch of C-code in the module traversing through all
the memory space of the application, then get worried....)

Handler
-------

The ABCI interface is handled by ``app``, which translates these data
structures into an internal format that is more convenient, but unable
to travel over the wire. The basic interface for any code that modifies
state is the ``Handler`` interface, which provides four methods:

::

      Name() string
      CheckTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
      DeliverTx(ctx Context, store state.KVStore, tx Tx) (Result, error)
      SetOption(l log.Logger, store state.KVStore, module, key, value string) (string, error)

Note the ``Context``, ``KVStore``, and ``Tx`` as principal carriers of
information. And that Result is always success, and we have a second
error return for errors (which is much more standard golang that
``res.IsErr()``)

The ``Handler`` interface is designed to be the basis for all modules
that execute transactions, and this can provide a large degree of code
interoperability, much like ``http.Handler`` does in golang web
development.

Modules
-------

TODO: update (s/Modules/handlers+mappers+stores/g) & add Msg + Tx (a signed message)

A module is a set of functionality which should be typically designed as
self-sufficient. Common elements of a module are:

-  transaction types (either end transactions, or transaction wrappers)
-  custom error codes
-  data models (to persist in the kv-store)
-  handler (to handle any end transactions)

Dispatcher
----------

We usually will want to have multiple modules working together, and need
to make sure the correct transactions get to the correct module. So we
have ``coin`` sending money, ``roles`` to create multi-sig accounts, and
``ibc`` for following other chains all working together without
interference.

We can then register a ``Dispatcher``, which
also implements the ``Handler`` interface. We then register a list of
modules with the dispatcher. Every module has a unique ``Name()``, which
is used for isolating its state space. We use this same name for routing
transactions. Each transaction implementation must be registed with
go-amino via ``TxMapper``, so we just look at the registered name of this
transaction, which should be of the form ``<module name>/xxx``. The
dispatcher grabs the appropriate module name from the tx name and routes
it if the module is present.

This all seems like a bit of magic, but really we're just making use of
go-amino magic that we are already using, rather than add another layer.
For all the transactions to be properly routed, the only thing you need
to remember is to use the following pattern:

::

    const (
      NameCoin = "coin"
      TypeSend = NameCoin + "/send"
    )

Permissions
-----------

TODO: replaces perms with object capabilities/object capability keys
- get rid of IPC

IPC requires a more complex permissioning system to allow the modules to
have limited access to each other and also to allow more types of
permissions than simple public key signatures. Rather than just use an
address to identify who is performing an action, we can use a more
complex structure:

::

    type Actor struct {
      ChainID string     `json:"chain"` // this is empty unless it comes from a different chain
      App     string     `json:"app"`   // the app that the actor belongs to
      Address data.Bytes `json:"addr"`  // arbitrary app-specific unique id
    }

Here, the ``Actor`` abstracts any address that can authorize actions,
hold funds, or initiate any sort of transaction. It doesn't just have to
be a pubkey on this chain, it could stem from another app (such as
multi-sig account), or even another chain (via IBC)

``ChainID`` is for IBC, discussed below. Let's focus on ``App`` and
``Address``. For a signature, the App is ``auth``, and any modules can
check to see if a specific public key address signed like this
``ctx.HasPermission(auth.SigPerm(addr))``. However, we can also
authorize a tx with ``roles``, which handles multi-sig accounts, it
checks if there were enough signatures by checking as above, then it can
add the role permission like
``ctx= ctx.WithPermissions(NewPerm(assume.Role))``

In addition to the permissions schema, the Actors are addresses just
like public key addresses. So one can create a mulit-sig role, then send
coin there, which can only be moved upon meeting the authorization
requirements from that module. ``coin`` doesn't even know the existence
of ``roles`` and one could build any other sort of module to provide
permissions (like bind the outcome of an election to move coins or to
modify the accounts on a role).

One idea - not yet implemented - is to provide scopes on the
permissions. Currently, if I sign a transaction to one module, it can
pass it on to any other module over IPC with the same permissions. It
could move coins, vote in an election, or anything else. Ideally, when
signing, one could also specify the scope(s) that this signature
authorizes. The `oauth
protocol <https://api.slack.com/docs/oauth-scopes>`__ also has to deal
with a similar problem, and maybe could provide some inspiration.
