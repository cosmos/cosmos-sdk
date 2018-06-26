# App1

The first application is a simple bank. Users have an account address and an account,
and they can send coins around. It has no authentication, and just uses JSON for
serialization.

## Messages

Messages are the primary inputs to the application state machine.
They define the content of transactions and can contain arbitrary information.
Developers can create messages by implementing the `Msg` interface:

```go
type Msg interface {

	// Return the message type.
	// Must be alphanumeric or empty.
    // Must correspond to name of message handler (XXX).
	Type() string

	// Get the canonical byte representation of the Msg.
    // This is what is signed.
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


The `Msg` interface allows messages to define basic validity checks, as well as
what needs to be signed and who needs to sign it.

Addresses in the SDK are arbitrary byte arrays that are hex-encoded when
displayed as a string or rendered in JSON. Typically, addresses are the hash of
a public key.

For instance, take the simple token sending message type from app1.go: 

```go
// MsgSend to send coins from Input to Output
type MsgSend struct {
	From   sdk.Address `json:"from"`
	To     sdk.Address `json:"to"`
	Amount sdk.Coins   `json:"amount"`
}

// Implements Msg.
func (msg MsgSend) Type() string { return "bank" }
```

It specifies that the message should be JSON marshaled and signed by the sender:

```go
// Implements Msg. JSON encode the message.
func (msg MsgSend) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// Implements Msg. Return the signer.
func (msg MsgSend) GetSigners() []sdk.Address {
	return []sdk.Address{msg.From}
}
```

The basic validity check ensures the From and To address are specified and the
amount is positive:

```go
// Implements Msg. Ensure the addresses are good and the
// amount is positive.
func (msg MsgSend) ValidateBasic() sdk.Error {
	if len(msg.From) == 0 {
		return sdk.ErrInvalidAddress("From address is empty")
	}
	if len(msg.To) == 0 {
		return sdk.ErrInvalidAddress("To address is empty")
	}
	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidCoins("Amount is not positive")
	}
	return nil
}
```

# KVStore

The basic persistence layer for an SDK application is the KVStore:

```go
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

    // TODO Not yet implemented.
    // CreateSubKVStore(key *storeKey) (KVStore, error)

    // TODO Not yet implemented.
    // GetSubKVStore(key *storeKey) KVStore
 }
```

Note it is unforgiving - it panics on nil keys!

The primary implementation of the KVStore is currently the IAVL store. In the future, we plan to support other Merkle KVStores, 
like Ethereum's radix trie.

As we'll soon see, apps have many distinct KVStores, each with a different name and for a different concern.
Access to a store is mediated by *object-capability keys*, which must be granted to a handler during application startup.

# Handlers

Now that we have a message type and a store interface, we can define our state transition function using a handler:

```go
// Handler defines the core of the state transition function of an application.
type Handler func(ctx Context, msg Msg) Result
```

Along with the message, the Handler takes environmental information (a `Context`), and returns a `Result`.

Where is the KVStore in all of this? Access to the KVStore in a message handler is restricted by the Context via object-capability keys.
Only handlers which were given explict access to a store's key will be able to access that store during message processsing.

## Context

The SDK uses a `Context` to propogate common information across functions. 
Most importantly, the `Context` restricts access to KVStores based on object-capability keys.
Only handlers which have been given explicit access to a key will be able to access the corresponding store.

For instance, the FooHandler can only load the store it's given the key for:

```go
// newFooHandler returns a Handler that can access a single store.
func newFooHandler(key sdk.StoreKey) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
        store := ctx.KVStore(key)
        // ...
    }
}
```

`Context` is modeled after the Golang [context.Context](TODO), which has
become ubiquitous in networking middleware and routing applications as a means
to easily propogate request context through handler functions.
Many methods on SDK objects receive a context as the first argument. 

The Context also contains the [block header](TODO), which includes the latest timestamp from the blockchain and other information about the latest block.

See the [Context API docs](TODO) for more details.

## Result

Result is motivated by the corresponding [ABCI result](TODO). It contains any return values, error information, logs, and meta data about the transaction:

```go
// Result is the union of ResponseDeliverTx and ResponseCheckTx.
type Result struct {

	// Code is the response code, is stored back on the chain.
	Code ABCICodeType

	// Data is any data returned from the app.
	Data []byte

	// Log is just debug information. NOTE: nondeterministic.
	Log string

	// GasWanted is the maximum units of work we allow this tx to perform.
	GasWanted int64

	// GasUsed is the amount of gas actually consumed. NOTE: unimplemented
	GasUsed int64

	// Tx fee amount and denom.
	FeeAmount int64
	FeeDenom  string

	// Tags are used for transaction indexing and pubsub.
	Tags Tags
}
```

## Handler

Let's define our handler for App1:

```go
func NewApp1Handler(keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, keyAcc, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
```

We have only a single message type, so just one message-specific function to define, `handleMsgSend`.

Note this handler has unfettered access to the store specified by the capability key `keyAcc`. So it must also define items in the store are encoded.
For this first example, we will define a simple account that is JSON encoded:

```go
type acc struct {
	Coins sdk.Coins `json:"coins"`
}
```

Coins is a useful type provided by the SDK for multi-asset accounts. While we could just use an integer here for a single coin type,
it's worth [getting to know `Coins`](TODO).


Now we're ready to handle the MsgSend:

```
// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, key *sdk.KVStoreKey, msg MsgSend) sdk.Result {
	// NOTE: from, to, and amount were already validated

	store := ctx.KVStore(key)
	bz := store.Get(msg.From)
	if bz == nil {
		// TODO
	}

	var acc acc
	err := json.Unmarshal(bz, &acc)
	if err != nil {
		// InternalError
	}

	// TODO: finish the logic

	return sdk.Result{
	// TODO: Tags
	}
}
```

The handler is straight forward:

- get the KVStore from the context using the granted capability key
- lookup the From address in the KVStore, and JSON unmarshal it into an `acc`,
- check that the account balance is greater than the `msg.Amount`
- transfer the `msg.Amount`

And that's that!

# BaseApp

Finally, we stitch it all together using the `BaseApp`.

The BaseApp is an abstraction over the [Tendermint
ABCI](https://github.com/tendermint/abci) that
simplifies application development by handling common low-level concerns.
It serves as the mediator between the two key components of an SDK app: the store
and the message handlers.

The BaseApp implements the
[`abci.Application`](https://godoc.org/github.com/tendermint/abci/types#Application) interface. 
It uses a `MultiStore` to manage the state, a `Router` for transaction handling, and 
`Set` methods to specify functions to run at the beginning and end of every
block. It's quite a work of art :).


Here is the complete setup for App1:

```go
func NewApp1(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// TODO: make this an interface or pass in
	// a TxDecoder instead.
	cdc := wire.NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app1Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")

	// Determine how transactions are decoded.
	app.SetTxDecoder(txDecoder)

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp1Handler(keyAccount))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}
```

Every app will have such a function that defines the setup of the app.
It will typically be contained in an `app.go` file.
We'll talk about how to connect this app object with the CLI, a REST API, 
the logger, and the filesystem later in the tutorial. For now, note that this is where we grant handlers access to stores.
Here, we have only one store and one handler, and the handler is granted access by giving it the capability key.
In future apps, we'll have multiple stores and handlers, and not every handler will get access to every store.

Note also the call to `SetTxDecoder`. While `Msg` contains the content for particular functionality in the application, the actual input
provided by the user is a serialized `Tx`. Applications may have many implementations of the `Msg` interface, but they should have only 
a single implementation of `Tx`:


```go
// Transactions wrap messages.
type Tx interface {
	// Gets the Msgs.
	GetMsgs() []Msg
}
```

The `Tx` just wraps a `[]Msg`, and may include additional authentication data, like signatures and account nonces. 
Applications must specify how their `Tx` is decoded, as this is the ultimate input into the application.
We'll talk more about `Tx` types later in the tutorial, specifically when we introduce the `StdTx`.

For this example, we have a dead-simple `Tx` type that contains the `MsgSend` and is JSON decoded:

```go
// Simple tx to wrap the Msg.
type app1Tx struct {
	MsgSend
}

// This tx only has one Msg.
func (tx app1Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.MsgSend}
}

// JSON decode MsgSend.
func txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx app1Tx
	err := json.Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode(err.Error())
	}
	return tx, nil
}
```

This means the input to the app must be a JSON encoded `app1Tx`.

In the next tutorial, we'll introduce Amino, a superior encoding scheme that lets us decode into interface types!

The last step in `NewApp1` is to mount the stores and load the latest version. Since we only have one store, we only mount one:

```go
	app.MountStoresIAVL(keyAccount)
```

We now have a complete implementation of a simple app. Next, we'll add another Msg type and another store, and use Amino for encoding!
