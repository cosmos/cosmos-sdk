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

```golang
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
	Type() string

	// Get some property of the Msg.
	Get(key interface{}) (value interface{})

	// Get the canonical byte representation of the Msg.
	GetSignBytes() []byte

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []crypto.Address
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

Finally, messages can provide generic access to their contents via `Get(key)`,
but this is mostly for convenience and not type-safe.

For instance, the `Basecoin` message types are defined in `x/bank/tx.go`: 

```golang
type SendMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type IssueMsg struct {
	Banker  crypto.Address `json:"banker"`
	Outputs []Output       `json:"outputs"`
}
```

Each specifies the addresses that must sign the message:

```golang
func (msg SendMsg) GetSigners() []crypto.Address {
	addrs := make([]crypto.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

func (msg IssueMsg) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Banker}
}
```

### Transactions

A transaction is a message with additional information for authentication:

```golang
type Tx interface {

	GetMsg() Msg

	// The address that pays the base fee for this message.  The fee is
	// deducted before the Msg is processed.
	GetFeePayer() crypto.Address

	// Get the canonical byte representation of the Tx.
	// Includes any signatures (or empty slots).
	GetTxBytes() []byte

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

```golang
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

Transactions can also specify the address responsible for paying the
transaction's fees using the `tx.GetFeePayer()` method.

The standard way to create a transaction from a message is to use the `StdTx`: 

```golang
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

```golang
type TxDecoder func(txBytes []byte) (Tx, error)
```

In `Basecoin`, we use the Tendermint wire format and the `go-wire` library for
encoding and decoding all message types.  The `go-wire` library has the nice
property that it can unmarshal into interface types, but it requires the
relevant types to be registered ahead of type. Registration happens on a
`Codec` object, so as not to taint the global name space.

For instance, in `Basecoin`, we wish to register the `SendMsg` and `IssueMsg`
types:

```golang
cdc.RegisterInterface((*sdk.Msg)(nil), nil)
cdc.RegisterConcrete(bank.SendMsg{}, "cosmos-sdk/SendMsg", nil)
cdc.RegisterConcrete(bank.IssueMsg{}, "cosmos-sdk/IssueMsg", nil)
```

Note how each concrete type is given a name - these name determine the type's
unique "prefix bytes" during encoding.  A registered type will always use the
same prefix-bytes, regardless of what interface it is satisfying.  For more
details, see the [go-wire documentation](https://github.com/tendermint/go-wire/blob/develop).


## MultiStore

### MultiStore is like a filesystem
### Mounting an IAVLStore

```
TODO:
- IAVLStore: Fast balanced dynamic Merkle store.
  - supports iteration.
- MultiStore: multiple Merkle tree backends in a single store 
  - allows using Ethereum Patricia Trie and Tendermint IAVL in same app
- Provide caching for intermediate state during execution of blocks and transactions (including for iteration)
- Historical state pruning and snapshotting.
- Query proofs (existence, absence, range, etc.) on current and retained historical state.
```

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

```golang
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

```golang
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
many coins. There are two message types: SendMsg and IssueMsg.
SendMsg allows coins to be sent around, while IssueMsg allows a
set of predefined users to issue new coins.

## Conclusion
