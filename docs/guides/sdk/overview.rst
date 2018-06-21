Overview
========

The SDK design optimizes flexibility and security. The
framework is designed around a modular execution stack which allows
applications to mix and match elements as desired. In addition,
all modules are sandboxed for greater application security.

Framework Overview
------------------

Object-Capability Model
~~~~~~~~~~~~~~~~~~~~~~~

When thinking about security, it's good to start with a specific threat model. Our threat model is the following:

::

    We assume that a thriving ecosystem of Cosmos-SDK modules that are easy to compose into a blockchain application will contain faulty or malicious modules.

The Cosmos-SDK is designed to address this threat by being the foundation of an object capability system.

::

    The structural properties of object capability systems favor
    modularity in code design and ensure reliable encapsulation in
    code implementation.

    These structural properties facilitate the analysis of some
    security properties of an object-capability program or operating
    system. Some of these — in particular, information flow properties
    — can be analyzed at the level of object references and
    connectivity, independent of any knowledge or analysis of the code
    that determines the behavior of the objects. As a consequence,
    these security properties can be established and maintained in the
    presence of new objects that contain unknown and possibly
    malicious code.

    These structural properties stem from the two rules governing
    access to existing objects:

    1) An object A can send a message to B only if object A holds a
    reference to B.

    2) An object A can obtain a reference to C only
    if object A receives a message containing a reference to C.  As a
    consequence of these two rules, an object can obtain a reference
    to another object only through a preexisting chain of references.
    In short, "Only connectivity begets connectivity."

See the `wikipedia article <https://en.wikipedia.org/wiki/Object-capability_model>`__ for more information.

Strictly speaking, Golang does not implement object capabilities completely, because of several issues:

* pervasive ability to import primitive modules (e.g. "unsafe", "os")
* pervasive ability to override module vars https://github.com/golang/go/issues/23161
* data-race vulnerability where 2+ goroutines can create illegal interface values

The first is easy to catch by auditing imports and using a proper dependency version control system like Dep.  The second and third are unfortunate but it can be audited with some cost.

Perhaps `Go2 will implement the object capability model <https://github.com/golang/go/issues/23157>`__.

What does it look like?
^^^^^^^^^^^^^^^^^^^^^^^

Only reveal what is necessary to get the work done.

For example, the following code snippet violates the object capabilities principle:

::

    type AppAccount struct {...}
    var account := &AppAccount{
    	Address: pub.Address(),
    	Coins: sdk.Coins{{"ATM", 100}},
    }
    var sumValue := externalModule.ComputeSumValue(account)

The method "ComputeSumValue" implies a pure function, yet the implied capability of accepting a pointer value is the capability to modify that value. The preferred method signature should take a copy instead.

::

    var sumValue := externalModule.ComputeSumValue(*account)

In the Cosmos SDK, you can see the application of this principle in the basecoin examples folder.

::

    // File: cosmos-sdk/examples/basecoin/app/init_handlers.go
    package app
    
    import (
    	"github.com/cosmos/cosmos-sdk/x/bank"
    	"github.com/cosmos/cosmos-sdk/x/sketchy"
    )
    
    func (app *BasecoinApp) initRouterHandlers() {
    
    	// All handlers must be added here.
    	// The order matters.
    	app.router.AddRoute("bank", bank.NewHandler(app.accountMapper))
    	app.router.AddRoute("sketchy", sketchy.NewHandler())
    }

In the Basecoin example, the sketchy handler isn't provided an account mapper, which does provide the bank handler with the capability (in conjunction with the context of a transaction run).

Security Overview
-----------------

For examples, see the `examples <https://github.com/cosmos/cosmos-sdk/tree/develop/examples>`__ directory.

Design Goals
~~~~~~~~~~~~

The design of the Cosmos SDK is based on the principles of "capabilities systems".

Capabilities systems
~~~~~~~~~~~~~~~~~~~~

TODO:

* Need for module isolation
* Capability is implied permission
* Link to thesis

Tx & Msg
~~~~~~~~

The SDK distinguishes between transactions (Tx) and messages
(Msg). A Tx is a Msg wrapped with authentication and fee data.

Messages
^^^^^^^^

Users can create messages containing arbitrary information by
implementing the ``Msg`` interface:

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

For instance, the ``Basecoin`` message types are defined in ``x/bank/tx.go``: 

::

    type SendMsg struct {
    	Inputs  []Input  `json:"inputs"`
    	Outputs []Output `json:"outputs"`
    }
    
    type IssueMsg struct {
    	Banker  sdk.Address `json:"banker"`
    	Outputs []Output       `json:"outputs"`
    }

Each specifies the addresses that must sign the message:

::

    func (msg SendMsg) GetSigners() []sdk.Address {
    	addrs := make([]sdk.Address, len(msg.Inputs))
    	for i, in := range msg.Inputs {
    		addrs[i] = in.Address
    	}
    	return addrs
    }
    
    func (msg IssueMsg) GetSigners() []sdk.Address {
    	return []sdk.Address{msg.Banker}
    }

Transactions
^^^^^^^^^^^^

A transaction is a message with additional information for authentication:

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
    	AccountNumber int64
        Sequence int64
    }

It contains the signature itself, as well as the corresponding account's account and 
sequence numbers.  The sequence number is expected to increment every time a
message is signed by a given account. The account number stays the same and is assigned
when the account is first generated.  These prevent "replay attacks", where
the same message could be executed over and over again.

The ``StdSignature`` can also optionally include the public key for verifying the
signature.  An application can store the public key for each address it knows
about, making it optional to include the public key in the transaction. In the
case of Basecoin, the public key only needs to be included in the first
transaction send by a given account - after that, the public key is forever
stored by the application and can be left out of transactions.

The standard way to create a transaction from a message is to use the ``StdTx``: 

::

    type StdTx struct {
    	Msg
    	Signatures []StdSignature
    }

Encoding and Decoding Transactions
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Messages and transactions are designed to be generic enough for developers to
specify their own encoding schemes.  This enables the SDK to be used as the
framwork for constructing already specified cryptocurrency state machines, for
instance Ethereum. 

When initializing an application, a developer must specify a ``TxDecoder``
function which determines how an arbitrary byte array should be unmarshalled
into a ``Tx``: 

::

    type TxDecoder func(txBytes []byte) (Tx, error)

In ``Basecoin``, we use the Tendermint wire format and the ``go-amino`` library for
encoding and decoding all message types.  The ``go-amino`` library has the nice
property that it can unmarshal into interface types, but it requires the
relevant types to be registered ahead of type. Registration happens on a
``Codec`` object, so as not to taint the global name space.

For instance, in ``Basecoin``, we wish to register the ``SendMsg`` and ``IssueMsg``
types:

::

    cdc.RegisterInterface((*sdk.Msg)(nil), nil)
    cdc.RegisterConcrete(bank.SendMsg{}, "cosmos-sdk/SendMsg", nil)
    cdc.RegisterConcrete(bank.IssueMsg{}, "cosmos-sdk/IssueMsg", nil)

Note how each concrete type is given a name - these name determine the type's
unique "prefix bytes" during encoding.  A registered type will always use the
same prefix-bytes, regardless of what interface it is satisfying.  For more
details, see the `go-amino documentation <https://github.com/tendermint/go-amino/tree/develop>`__.


MultiStore
~~~~~~~~~~

MultiStore is like a filesystem
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Mounting an IAVLStore
^^^^^^^^^^^^^^^^^^^^^

TODO:

* IAVLStore: Fast balanced dynamic Merkle store.

  * supports iteration.

* MultiStore: multiple Merkle tree backends in a single store 
  
  * allows using Ethereum Patricia Trie and Tendermint IAVL in same app

* Provide caching for intermediate state during execution of blocks and transactions (including for iteration)
* Historical state pruning and snapshotting.
* Query proofs (existence, absence, range, etc.) on current and retained historical state.

Context
-------

The SDK uses a ``Context`` to propogate common information across functions. The
``Context`` is modelled after the Golang ``context.Context`` object, which has
become ubiquitous in networking middleware and routing applications as a means
to easily propogate request context through handler functions.

The main information stored in the ``Context`` includes the application
MultiStore (see below), the last block header, and the transaction bytes.
Effectively, the context contains all data that may be necessary for processing
a transaction.

Many methods on SDK objects receive a context as the first argument. 

Handler
-------

Transaction processing in the SDK is defined through ``Handler`` functions:

::

    type Handler func(ctx Context, tx Tx) Result

A handler takes a context and a transaction and returns a result.  All
information necessary for processing a transaction should be available in the
context.

While the context holds the entire application state (all referenced from the
root MultiStore), a particular handler only needs a particular kind of access
to a particular store (or two or more). Access to stores is managed using
capabilities keys and mappers.  When a handler is initialized, it is passed a
key or mapper that gives it access to the relevant stores.

::

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

AnteHandler
-----------

Handling Fee payment
~~~~~~~~~~~~~~~~~~~~

Handling Authentication
~~~~~~~~~~~~~~~~~~~~~~~

Accounts and x/auth
-------------------

sdk.Account
~~~~~~~~~~~

auth.BaseAccount
~~~~~~~~~~~~~~~~

auth.AccountMapper
~~~~~~~~~~~~~~~~~~

Wire codec
----------

Why another codec?
~~~~~~~~~~~~~~~~~~

vs encoding/json
~~~~~~~~~~~~~~~~

vs protobuf
~~~~~~~~~~~

KVStore example
---------------

Basecoin example
----------------

The quintessential SDK application is Basecoin - a simple
multi-asset cryptocurrency.  Basecoin consists of a set of
accounts stored in a Merkle tree, where each account may have
many coins. There are two message types: SendMsg and IssueMsg.
SendMsg allows coins to be sent around, while IssueMsg allows a
set of predefined users to issue new coins.
