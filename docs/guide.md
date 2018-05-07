## Introduction

If you want to see some examples, take a look at the [examples/basecoin](/examples/basecoin) directory.

## Design Goals

The design of the Cosmos SDK is based on the principles of "capabilities systems".

## Capabilities systems

### Need for module isolation
### Capability is implied permission
### TODO Link to thesis

## Tx & Msg

The SDK distinguishes between transactions (Tx) and messages
(Msg).  A Tx is a Msg wrapped with authentication and fee data.

### Messages

Users can create messages containing arbitrary information by
implementing the `Msg` interface:

```go
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

```

Messages must specify their type via the `Type()` method. The type should
correspond to the messages handler, so there can be many messages with the same
type.

Messages must also specify how they are to be authenticated. The `GetSigners()`
method return a list of addresses that must sign the message, while the
`GetSignBytes()` method returns the bytes that must be signed for a signature
to be valid.

Addresses in the SDK are arbitrary byte arrays that are hex-encoded when
displayed as a string or rendered in JSON.

Messages can specify basic self-consistency checks using the `ValidateBasic()`
method to enforce that message contents are well formed before any actual logic
begins.

For instance, the `Basecoin` message types are defined in `x/bank/tx.go`: 

```go
type MsgSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type MsgIssue struct {
	Banker  sdk.Address `json:"banker"`
	Outputs []Output       `json:"outputs"`
}
```

Each specifies the addresses that must sign the message:

```go
func (msg MsgSend) GetSigners() []sdk.Address {
	addrs := make([]sdk.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

func (msg MsgIssue) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Banker}
}
```

### Transactions

A transaction is a message with additional information for authentication:

```go
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
```

The `tx.GetSignatures()` method returns a list of signatures, which must match
the list of addresses returned by `tx.Msg.GetSigners()`. The signatures come in
a standard form:

```go
type StdSignature struct {
	crypto.PubKey // optional
	crypto.Signature
	Sequence int64
}
```

It contains the signature itself, as well as the corresponding account's
sequence number.  The sequence number is expected to increment every time a
message is signed by a given account.  This prevents "replay attacks", where
the same message could be executed over and over again.

The `StdSignature` can also optionally include the public key for verifying the
signature.  An application can store the public key for each address it knows
about, making it optional to include the public key in the transaction. In the
case of Basecoin, the public key only needs to be included in the first
transaction send by a given account - after that, the public key is forever
stored by the application and can be left out of transactions.

The address responsible for paying the transactions fee is the first address
returned by msg.GetSigners(). The convenience function `FeePayer(tx Tx)` is provided
to return this.

The standard way to create a transaction from a message is to use the `StdTx`: 

```go
type StdTx struct {
	Msg
	Signatures []StdSignature
}
```

### Encoding and Decoding Transactions

Messages and transactions are designed to be generic enough for developers to
specify their own encoding schemes.  This enables the SDK to be used as the
framwork for constructing already specified cryptocurrency state machines, for
instance Ethereum. 

When initializing an application, a developer must specify a `TxDecoder`
function which determines how an arbitrary byte array should be unmarshalled
into a `Tx`: 

```go
type TxDecoder func(txBytes []byte) (Tx, error)
```

In `Basecoin`, we use the Tendermint wire format and the `go-amino` library for
encoding and decoding all message types.  The `go-amino` library has the nice
property that it can unmarshal into interface types, but it requires the
relevant types to be registered ahead of type. Registration happens on a
`Codec` object, so as not to taint the global name space.

For instance, in `Basecoin`, we wish to register the `MsgSend` and `MsgIssue`
types:

```go
cdc.RegisterInterface((*sdk.Msg)(nil), nil)
cdc.RegisterConcrete(bank.MsgSend{}, "cosmos-sdk/MsgSend", nil)
cdc.RegisterConcrete(bank.MsgIssue{}, "cosmos-sdk/MsgIssue", nil)
```

Note how each concrete type is given a name - these name determine the type's
unique "prefix bytes" during encoding.  A registered type will always use the
same prefix-bytes, regardless of what interface it is satisfying.  For more
details, see the [go-amino documentation](https://github.com/tendermint/go-amino/blob/develop).


## Storage

### MultiStore

MultiStore is like a root filesystem of an operating system, except
all the entries are fully Merkleized.  You mount onto a MultiStore
any number of Stores.  Currently only KVStores are supported, but in
the future we may support more kinds of stores, such as a HeapStore
or a NDStore for multidimensional storage.

The MultiStore as well as all mounted stores provide caching (aka
cache-wrapping) for intermediate state (aka software transactional
memory) during the execution of transactions.  In the case of the
KVStore, this also works for iterators.  For example, after running
the app's AnteHandler, the MultiStore is cache-wrapped (and each
store is also cache-wrapped) so that should processing of the
transaction fail, at least the transaction fees are paid and
sequence incremented.

The MultiStore as well as all stores support (or will support)
historical state pruning and snapshotting and various kinds of
queries with proofs.

### KVStore

Here we'll focus on the IAVLStore, which is a kind of KVStore.

IAVLStore is a fast balanced dynamic Merkle store that also supports
iteration, and of course cache-wrapping, state pruning, and various
queries with proofs, such as proofs of existence, absence, range,
and so on.

Here's how you mount them to a MultiStore.

```go
mainDB, catDB := dbm.NewMemDB(), dbm.NewMemDB()
fooKey := sdk.NewKVStoreKey("foo")
barKey := sdk.NewKVStoreKey("bar")
catKey := sdk.NewKVStoreKey("cat")
ms := NewCommitMultiStore(mainDB)
ms.MountStoreWithDB(fooKey, sdk.StoreTypeIAVL, nil)
ms.MountStoreWithDB(barKey, sdk.StoreTypeIAVL, nil)
ms.MountStoreWithDB(catKey, sdk.StoreTypeIAVL, catDB)
```

In the example above, all IAVL nodes (inner and leaf) will be stored
in mainDB with the prefix of "s/k:foo/" and "s/k:bar/" respectively,
thus sharing the mainDB.  All IAVL nodes (inner and leaf) for the
cat KVStore are stored separately in catDB with the prefix of
"s/\_/".  The "s/k:KEY/" and "s/\_/" prefixes are there to
disambiguate store items from other items of non-storage concern.


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

Transaction processing in the SDK is defined through `Handler` functions:

```go
type Handler func(ctx Context, tx Tx) Result
```

A handler takes a context and a transaction and returns a result.  All
information necessary for processing a transaction should be available in the
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

## AnteHandler

### Handling Fee payment
### Handling Authentication

## Accounts and x/auth

### sdk.Account
### auth.BaseAccount
### auth.AccountMapper

## Wire codec

### Why another codec?
### vs encoding/json
### vs protobuf

## KVStore example

## Basecoin example

The quintessential SDK application is Basecoin - a simple
multi-asset cryptocurrency.  Basecoin consists of a set of
accounts stored in a Merkle tree, where each account may have
many coins. There are two message types: MsgSend and MsgIssue.
MsgSend allows coins to be sent around, while MsgIssue allows a
set of predefined users to issue new coins.

## Conclusion
