# The Basics

Here we introduce the basic components of an SDK by building `App1`, a simple bank.
Users have an account address and an account, and they can send coins around.
It has no authentication, and just uses JSON for serialization.

The complete code can be found in [app1.go](examples/app1.go).

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

    // ValidateBasic does a simple validation check that
    // doesn't require access to any other information.
    ValidateBasic() error

    // Get the canonical byte representation of the Msg.
    // This is what is signed.
    GetSignBytes() []byte

    // Signers returns the addrs of signers that must sign.
    // CONTRACT: All signatures must be present to be valid.
    // CONTRACT: Returns addrs in some deterministic order.
    GetSigners() []Address
}
```


The `Msg` interface allows messages to define basic validity checks, as well as
what needs to be signed and who needs to sign it.

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

Note Addresses in the SDK are arbitrary byte arrays that are
[Bech32](https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki) encoded
when displayed as a string or rendered in JSON. Typically, addresses are the hash of
a public key, so we can use them to uniquely identify the required signers for a
transaction.


The basic validity check ensures the From and To address are specified and the
Amount is positive:

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

Note the `ValidateBasic` method is called automatically by the SDK!

## KVStore

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

 }
```

Note it is unforgiving - it panics on nil keys!

The primary implementation of the KVStore is currently the IAVL store. In the future, we plan to support other Merkle KVStores,
like Ethereum's radix trie.

As we'll soon see, apps have many distinct KVStores, each with a different name and for a different concern.
Access to a store is mediated by *object-capability keys*, which must be granted to a handler during application startup.

## Handlers

Now that we have a message type and a store interface, we can define our state transition function using a handler:

```go
// Handler defines the core of the state transition function of an application.
type Handler func(ctx Context, msg Msg) Result
```

Along with the message, the Handler takes environmental information (a `Context`), and returns a `Result`.
All information necessary for processing a message should be available in the context.

Where is the KVStore in all of this? Access to the KVStore in a message handler is restricted by the Context via object-capability keys.
Only handlers which were given explict access to a store's key will be able to access that store during message processsing.

### Context

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

`Context` is modeled after the Golang
[context.Context](https://golang.org/pkg/context/), which has
become ubiquitous in networking middleware and routing applications as a means
to easily propogate request context through handler functions.
Many methods on SDK objects receive a context as the first argument.

The Context also contains the
[block header](https://github.com/tendermint/tendermint/blob/master/docs/spec/blockchain/blockchain.md#header),
which includes the latest timestamp from the blockchain and other information about the latest block.

See the [Context API
docs](https://godoc.org/github.com/cosmos/cosmos-sdk/types#Context) for more details.

### Result

Handler takes a Context and Msg and returns a Result.
Result is motivated by the corresponding [ABCI result](https://github.com/tendermint/tendermint/blob/master/abci/types/types.proto#L165).
It contains return values, error information, logs, and meta data about the transaction:

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

We'll talk more about these fields later in the tutorial. For now, note that a
`0` value for the `Code` is considered a success, and everything else is a
failure. The `Tags` can contain meta data about the transaction that will allow
us to easily lookup transactions that pertain to particular accounts or actions.

### Handler

Let's define our handler for App1:

```go
// Handle MsgSend.
// NOTE: msg.From, msg.To, and msg.Amount were already validated
// in ValidateBasic().
func handleMsgSend(ctx sdk.Context, key *sdk.KVStoreKey, msg MsgSend) sdk.Result {
	// Load the store.
	store := ctx.KVStore(key)

	// Debit from the sender.
	if res := handleFrom(store, msg.From, msg.Amount); !res.IsOK() {
		return res
	}

	// Credit the receiver.
	if res := handleTo(store, msg.To, msg.Amount); !res.IsOK() {
		return res
	}

	// Return a success (Code 0).
	// Add list of key-value pair descriptors ("tags").
	return sdk.Result{
		Tags: msg.Tags(),
	}
}
```

We have only a single message type, so just one message-specific function to define, `handleMsgSend`.

Note this handler has unrestricted access to the store specified by the capability key `keyAcc`,
so it must define what to store and how to encode it. Later, we'll introduce
higher-level abstractions so Handlers are restricted in what they can do.
For this first example, we use a simple account that is JSON encoded:

```go
type appAccount struct {
	Coins sdk.Coins `json:"coins"`
}
```

Coins is a useful type provided by the SDK for multi-asset accounts.
We could just use an integer here for a single coin type, but
it's worth [getting to know
Coins](https://godoc.org/github.com/cosmos/cosmos-sdk/types#Coins).


Now we're ready to handle the two parts of the MsgSend:

```go
func handleFrom(store sdk.KVStore, from sdk.Address, amt sdk.Coins) sdk.Result {
	// Get sender account from the store.
	accBytes := store.Get(from)
	if accBytes == nil {
		// Account was not added to store. Return the result of the error.
		return sdk.NewError(2, 101, "Account not added to store").Result()
	}

	// Unmarshal the JSON account bytes.
	var acc appAccount
	err := json.Unmarshal(accBytes, &acc)
	if err != nil {
		// InternalError
		return sdk.ErrInternal("Error when deserializing account").Result()
	}

	// Deduct msg amount from sender account.
	senderCoins := acc.Coins.Minus(amt)

	// If any coin has negative amount, return insufficient coins error.
	if !senderCoins.IsNotNegative() {
		return sdk.ErrInsufficientCoins("Insufficient coins in account").Result()
	}

	// Set acc coins to new amount.
	acc.Coins = senderCoins

	// Encode sender account.
	accBytes, err = json.Marshal(acc)
	if err != nil {
		return sdk.ErrInternal("Account encoding error").Result()
	}

	// Update store with updated sender account
	store.Set(from, accBytes)
	return sdk.Result{}
}

func handleTo(store sdk.KVStore, to sdk.Address, amt sdk.Coins) sdk.Result {
	// Add msg amount to receiver account
	accBytes := store.Get(to)
	var acc appAccount
	if accBytes == nil {
		// Receiver account does not already exist, create a new one.
		acc = appAccount{}
	} else {
		// Receiver account already exists. Retrieve and decode it.
		err := json.Unmarshal(accBytes, &acc)
		if err != nil {
			return sdk.ErrInternal("Account decoding error").Result()
		}
	}

	// Add amount to receiver's old coins
	receiverCoins := acc.Coins.Plus(amt)

	// Update receiver account
	acc.Coins = receiverCoins

	// Encode receiver account
	accBytes, err := json.Marshal(acc)
	if err != nil {
		return sdk.ErrInternal("Account encoding error").Result()
	}

	// Update store with updated receiver account
	store.Set(to, accBytes)
	return sdk.Result{}
}
```

The handler is straight forward. We first load the KVStore from the context using the granted capability key.
Then we make two state transitions: one for the sender, one for the receiver.
Each one involves JSON unmarshalling the account bytes from the store, mutating
the `Coins`, and JSON marshalling back into the store.

And that's that!

## Tx

The final piece before putting it all together is the `Tx`.
While `Msg` contains the content for particular functionality in the application, the actual input
provided by the user is a serialized `Tx`. Applications may have many implementations of the `Msg` interface,
but they should have only a single implementation of `Tx`:


```go
// Transactions wrap messages.
type Tx interface {
	// Gets the Msgs.
	GetMsgs() []Msg
}
```

The `Tx` just wraps a `[]Msg`, and may include additional authentication data, like signatures and account nonces.
Applications must specify how their `Tx` is decoded, as this is the ultimate input into the application.
We'll talk more about `Tx` types later, specifically when we introduce the `StdTx`.

In this first application, we won't have any authentication at all. This might
make sense in a private network where access is controlled by alternative means,
like client-side TLS certificates, but in general, we'll want to bake the authentication
right into our state machine. We'll use `Tx` to do that
in the next app. For now, the `Tx` just embeds `MsgSend` and uses JSON:


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

## BaseApp

Finally, we stitch it all together using the `BaseApp`.

The BaseApp is an abstraction over the [Tendermint
ABCI](https://github.com/tendermint/tendermint/tree/master/abci) that
simplifies application development by handling common low-level concerns.
It serves as the mediator between the two key components of an SDK app: the store
and the message handlers. The BaseApp implements the
[`abci.Application`](https://godoc.org/github.com/tendermint/tendermint/abci/types#Application) interface.
See the [BaseApp API
documentation](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp) for more details.

Here is the complete setup for App1:

```go
func NewApp1(ctx *sdk.ServerContext, db dbm.DB) *bapp.BaseApp {
    cdc := wire.NewCodec()

    // Create the base application object.
    app := bapp.NewBaseApp(app1Name, cdc, ctx, db)

    // Create a capability key for accessing the account store.
    keyAccount := sdk.NewKVStoreKey("acc")

    // Determine how transactions are decoded.
    app.SetTxDecoder(txDecoder)

    // Register message routes.
    // Note the handler receives the keyAccount and thus
    // gets access to the account store.
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
the logger, and the filesystem later in the tutorial. For now, note that this is where we
register handlers for messages and grant them access to stores.

Here, we have only a single Msg type, `bank`, a single store for accounts, and a single handler.
The handler is granted access to the store by giving it the capability key.
In future apps, we'll have multiple stores and handlers, and not every handler will get access to every store.

After setting the transaction decoder and the message handling routes, the final
step is to mount the stores and load the latest version.
Since we only have one store, we only mount one.

## Execution

We're now done the core logic of the app! From here, we can write tests in Go
that initialize the store with accounts and execute transactions by calling
the `app.DeliverTx` method.

In a real setup, the app would run as an ABCI application on top of the
Tendermint consensus engine. It would be initialized by a Genesis file, and it
would be driven by blocks of transactions committed by the underlying Tendermint
consensus. We'll talk more about ABCI and how this all works a bit later, but
feel free to check the
[specification](https://github.com/tendermint/tendermint/blob/master/docs/abci-spec.md).
We'll also see how to connect our app to a complete suite of components
for running and using a live blockchain application.

For now, we note the follow sequence of events occurs when a transaction is
received (through `app.DeliverTx`):

- serialized transaction is received by `app.DeliverTx`
- transaction is deserialized using `TxDecoder`
- for each message in the transaction, run `msg.ValidateBasic()`
- for each message in the transaction, load the appropriate handler and execute
  it with the message

## Conclusion

We now have a complete implementation of a simple app!

In the next section, we'll add another Msg type and another store. Once we have multiple message types
we'll need a better way of decoding transactions, since we'll need to decode
into the `Msg` interface. This is where we introduce Amino, a superior encoding scheme that lets us decode into interface types!
