Glossary
========

.. (s/Modules/handlers+mappers+stores/g) + add Tx (signed message) & Msg

This glossary defines many terms used throughout documentation of the Cosmos-SDK.
This is mainly to provide a background and general understanding of the
different words and concepts that are used. Other documents will explain
in more detail how to combine these concepts to build a particular
application.

If you are unfamiliar with other Cosmos and Tendermint terminology, please refer
to the Cosmos Academy docs.

Can't find the term that you're looking for? Please submit a PR or fill this form
and we'll add it to this glossary.

Composability
-------------

Anyone can create a module for the Cosmos-SDK and integrating the already-built
modules is as simple as importing them into your blockchain application.

Capabilities
------------



Keys
----
To sign a message, you'll need a public and private key pair.
These are cryptographic schemes that are generated from an elliptic curve.
Cosmos uses ``ed25519`` elliptic curve by default for key generation.

To recover a lost key you'll have to use the 24 word seed phrase you were given when you generated.

Signatures
----------
A Signature in the SDK is a bytes message that you attest with your private key.

We define a standard SDK signature as the following:

::

    type StdSignature struct {
    crypto.PubKey    `json:"pub_key"` // optional
    crypto.Signature `json:"signature"`
    Sequence         int64 `json:"sequence"`
    }

Accounts
--------


Messages
--------

Messages are packets containing arbitrary information.

::

  type Msg interface {

    // Return the message type.
    // Must be alphanumeric or empty.
    Type() string

    // Get the canonical byte representation of the Msg.
    GetSignBytes() []byte

    // ValidateBasic does a simple validation check that
    // doesn't require access to any other information.
    ValidateBasic() error

    // Signers returns the addrs of signers that must sign.
    // CONTRACT: All signatures must be present to be valid.
    // CONTRACT: Returns addrs in some deterministic order.
    GetSigners() []Address
  }

Messages must specify their type via the ``Type()`` method. The type should
correspond to the messages handler, so there can be many messages with the same
type.

Messages must also specify how they are to be authenticated. The ``GetSigners()``
method return a list of addresses that must sign the message, while the
``GetSignBytes()`` method returns the bytes that must be signed for a signature
to be valid.

Addresses in the SDK are arbitrary byte arrays that are hex-encoded when
displayed as a string or rendered in JSON.

Messages can specify basic self-consistency checks using the ``ValidateBasic()``
method to enforce that message contents are well formed before any actual logic
begins.

Transaction
-----------

A transaction (``Tx``) is a packet of binary data that contains all information
to validate and perform an action on the blockchain. In general, a ``Tx`` is a ``Msg``
with additional data for authentication and fees.


The only other data that it interacts with is the current state of the chain
(key-value store), and it must have a deterministic action. The transaction is the
main piece of one request.

All transactions must contain a ``Msg`` and a destination.

Note how we can wrap any other transaction, add a fee level, and not
worry about the encoding in our module any more?

::

  type Tx interface {

    GetMsg() Msg

    // Signatures returns the signature of signers who signed the Msg.
    // CONTRACT: Length returned is same as length of
    // pubkeys returned from MsgKeySigners, and the order
    // matches.
    // CONTRACT: If the signature is missing (ie the Msg is
    // invalid), then the corresponding signature is
    // .Empty().
    GetSignatures() []StdSignature
  }

The ``tx.GetSignatures()`` method returns a list of signatures, which must match
the list of addresses returned by ``tx.Msg.GetSigners()``. The signatures come in
a standard form:

::

  type StdSignature struct {
  	crypto.PubKey // optional
  	crypto.Signature
  	Sequence int64
  }

The standard way to create a transaction from a message is to use the `StdTx`:

::

  type StdTx struct {
  	Msg
  	Signatures []StdSignature
  }


Rational
--------

The SDK implementation of rational numbers (*a.k.a* ``Rat``) is based on the ``math/big``
Golang library with additional methods for increased security and functionalities.


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
key-value (KV) store which is indexed with a merkle tree. This allows
for the easy generation of a root hash and proofs for queries without
requiring complex logic inside each module. Standardization of this
process also allows powerful light-client tooling as any store data may
be verified on the fly.

::

  type Store interface {
  	GetStoreType() StoreType
  	CacheWrapper
  }

Multi Store
^^^^^^^^^^^

::

  type MultiStore interface { //nolint
  	Store

  	// Cache wrap MultiStore.
  	// NOTE: Caller should probably not call .Write() on each, but
  	// call CacheMultiStore.Write().
  	CacheMultiStore() CacheMultiStore

  	// Convenience for fetching substores.
  	GetStore(StoreKey) Store
  	GetKVStore(StoreKey) KVStore
  	GetKVStoreWithGas(GasMeter, StoreKey) KVStore
  }

Key-Value Store
^^^^^^^^^^^^^^^

``KVStore`` is a simple interface to get/set data.

The largest limitation of the current implemenation of the kv-store is
that interface that the application must use can only ``Get`` and
``Set`` single data points. That said, there are some data structures
like queues and range queries that are available in ``state`` package.
These provide higher-level functionality in a standard format, but have
not yet been integrated into the kv-store interface.

::

  type KVStore interface {
  	Store

  	// Get returns nil iff key doesn't exist. Panics on nil key.
  	Get(key []byte) []byte

  	// Has checks if a key exists. Panics on nil key.
  	Has(key []byte) bool

  	// Set sets the key. Panics on nil key.
  	Set(key, value []byte)

  	// Delete deletes the key. Panics on nil key.
  	Delete(key []byte)

  	// Iterator over a domain of keys in ascending order. End is exclusive.
  	// Start must be less than end, or the Iterator is invalid.
  	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
  	Iterator(start, end []byte) Iterator

  	// Iterator over a domain of keys in descending order. End is exclusive.
  	// Start must be greater than end, or the Iterator is invalid.
  	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
  	ReverseIterator(start, end []byte) Iterator

  	// Iterator over all the keys with a certain prefix in ascending order.
  	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
  	SubspaceIterator(prefix []byte) Iterator

  	// Iterator over all the keys with a certain prefix in descending order.
  	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
  	ReverseSubspaceIterator(prefix []byte) Iterator

  	// TODO Not yet implemented.
  	// CreateSubKVStore(key *storeKey) (KVStore, error)

  	// TODO Not yet implemented.
  	// GetSubKVStore(key *storeKey) KVStore

  }

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

Transaction processing in the SDK is defined through ``Handler`` functions:

A handler takes a context and a transaction and returns a result.  All
information necessary for processing a transaction should be available in the
context.

- **Handler**:
- **FeeHandler**: application runs to handle fees
- **AnteHandler**: handler that checks and increments sequence numbers, checks signatures and deducts fees from the first signer.

While the context holds the entire application state (all referenced from the
root MultiStore), a particular handler only needs a particular kind of access
to a particular store (or two or more). Access to stores is managed using
capabilities keys and mappers.  When a handler is initialized, it is passed a
key or mapper that gives it access to the relevant stores.

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

Keepers
-------

Mappers that can be passed to other modules to grant a pre-defined set of capabilities.
For example, if an instance of module A's ``keepers`` is passed to module B,
the latter will be able to call a restricted set of module A's functions.

A keeper is used for getting and setting values.

Codec
-----


Modules
-------

Each module is an extension of the ``BaseApp`` functionalities that defines transactions,
handles application state and the state transition logic.

Common elements of a module are:

-  Transaction types (either end transactions, or transaction wrappers)
-  Custom error codes
-  Data models (*i.e.* types) to persist in the ``KV-store``
-  Handlers: to handle the logic of messages and transactions

SDK modules are stored inside the ``x`` folder. The current prebuilt-modules for the SDK are:
*Auth*, *Bank*, *Staking* and *IBC*.


BaseApp
-------
``BaseApp`` is the basic application of the Cosmos-SDK. When you create a new SDK app, ypu must define its name, logger and database
``BaseApp`` provides data structures that provide basic data storage
functionality and act as a bridge between the ABCI interface and the SDK
abstractions.

``BaseApp`` has no state except the CommitMultiStore you provide upon init.

::

  type BaseApp struct {
    // initialized on creation
    Logger     log.Logger
    name       string               // application name from abci.Info
    cdc        *wire.Codec          // Amino codec
    db         dbm.DB               // common DB backend
    cms        sdk.CommitMultiStore // Main (uncached) state
    router     Router               // handle any kind of message
    codespacer *sdk.Codespacer      // handle module codespacing

    // must be set
    txDecoder   sdk.TxDecoder   // unmarshal []byte into sdk.Tx
    anteHandler sdk.AnteHandler // ante handler for fee and auth

    // may be nil
    initChainer      sdk.InitChainer  // initialize state with validators and state blob
    beginBlocker     sdk.BeginBlocker // logic to run before any txs
    endBlocker       sdk.EndBlocker   // logic to run after all txs, and to determine valset changes
    addrPeerFilter   sdk.PeerFilter   // filter peers by address and port
    pubkeyPeerFilter sdk.PeerFilter   // filter peers by public key

    //--------------------
    // Volatile
    // checkState is set on initialization and reset on Commit.
    // deliverState is set in InitChain and BeginBlock and cleared on Commit.
    // See methods setCheckState and setDeliverState.
    // valUpdates accumulate in DeliverTx and are reset in BeginBlock.
    checkState   *state           // for CheckTx
    deliverState *state           // for DeliverTx
    valUpdates   []abci.Validator // cached validator changes from DeliverTx
  }

Apps
----

Apps on the Cosmos-SDK are built on top of the ``BaseApp`` functionalities.

Router
------

A ``Router`` is a struct that provides handlers for each transaction type.

Coin
-----

A Coin is a struct in the SDK that holds some amount of a currency. It also contains methods to do same math operations.

::

    type Coin struct {
    	Denom  string `json:"denom"`
    	Amount int64  `json:"amount"`
    }


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

.. TODO: replaces perms with object capabilities/object capability keys

.. - get rid of IPC

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

Testnet
-------



Middleware
----------
