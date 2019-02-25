# Transactions

In the previous app we built a simple bank with one message type `send` for sending
coins and one store for storing accounts.
Here we build `App2`, which expands on `App1` by introducing

- a new message type for issuing new coins
- a new store for coin metadata (like who can issue coins)
- a requirement that transactions include valid signatures

Along the way, we'll be introduced to Amino for encoding and decoding
transactions and to the AnteHandler for processing them.

The complete code can be found in [app2.go](examples/app2.go).


## Message

Let's introduce a new message type for issuing coins:

```go
// MsgIssue to allow a registered issuer
// to issue new coins.
type MsgIssue struct {
	Issuer   sdk.AccAddress
	Receiver sdk.AccAddress
	Coin     sdk.Coin
}

// Implements Msg.
func (msg MsgIssue) Route() string { return "issue" }
```

Note the `Route()` method returns `"issue"`, so this message is of a different
route and will be executed by a different handler than `MsgSend`. The other
methods for `MsgIssue` are similar to `MsgSend`.

## Handler

We'll need a new handler to support the new message type. It just checks if the
sender of the `MsgIssue` is the correct issuer for the given coin type, as per the information
in the issuer store:

```go
// Handle MsgIssue
func handleMsgIssue(keyIssue *sdk.KVStoreKey, keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		issueMsg, ok := msg.(MsgIssue)
		if !ok {
			return sdk.NewError(2, 1, "MsgIssue is malformed").Result()
		}

		// Retrieve stores
		issueStore := ctx.KVStore(keyIssue)
		accStore := ctx.KVStore(keyAcc)

		// Handle updating coin info
		if res := handleIssuer(issueStore, issueMsg.Issuer, issueMsg.Coin); !res.IsOK() {
			return res
		}

		// Issue coins to receiver using previously defined handleTo function
		if res := handleTo(accStore, issueMsg.Receiver, []sdk.Coin{issueMsg.Coin}); !res.IsOK() {
			return res
		}

		return sdk.Result{
			// Return result with Issue msg tags
			Tags: issueMsg.Tags(),
		}
	}
}

func handleIssuer(store sdk.KVStore, issuer sdk.AccAddress, coin sdk.Coin) sdk.Result {
	// the issuer address is stored directly under the coin denomination
	denom := []byte(coin.Denom)
	infoBytes := store.Get(denom)
	if infoBytes == nil {
		return sdk.ErrInvalidCoins(fmt.Sprintf("Unknown coin type %s", coin.Denom)).Result()
	}

	var coinInfo coinInfo
	err := json.Unmarshal(infoBytes, &coinInfo)
	if err != nil {
		return sdk.ErrInternal("Error when deserializing coinInfo").Result()
	}

	// Msg Issuer is not authorized to issue these coins
	if !bytes.Equal(coinInfo.Issuer, issuer) {
		return sdk.ErrUnauthorized(fmt.Sprintf("Msg Issuer cannot issue tokens: %s", coin.Denom)).Result()
	}

	return sdk.Result{}
}

// coinInfo stores meta data about a coin
type coinInfo struct {
	Issuer sdk.AccAddress `json:"issuer"`
}
```

Note we've introduced the `coinInfo` type to store the issuer address for each coin.
We JSON serialize this type and store it directly under the denomination in the
issuer store. We could of course add more fields and logic around this,
like including the current supply of coins in existence, and enforcing a maximum supply,
but that's left as an excercise for the reader :).

## Amino

Now that we have two implementations of `Msg`, we won't know before hand
which type is contained in a serialized `Tx`. Ideally, we would use the
`Msg` interface inside our `Tx` implementation, but the JSON decoder can't
decode into interface types. In fact, there's no standard way to unmarshal
into interfaces in Go. This is one of the primary reasons we built
[Amino](https://github.com/tendermint/go-amino) :).

While SDK developers can encode transactions and state objects however they
like, Amino is the recommended format. The goal of Amino is to improve over the latest version of Protocol Buffers,
`proto3`. To that end, Amino is compatible with the subset of `proto3` that
excludes the `oneof` keyword.

While `oneof` provides union types, Amino aims to provide interfaces.
The main difference being that with union types, you have to know all the types
up front. But anyone can implement an interface type whenever and however
they like.

To implement interface types, Amino allows any concrete implementation of an
interface to register a globally unique name that is carried along whenever the
type is serialized. This allows Amino to seamlessly deserialize into interface
types!

The primary use for Amino in the SDK is for messages that implement the
`Msg` interface. By registering each message with a distinct name, they are each
given a distinct Amino prefix, allowing them to be easily distinguished in
transactions.

Amino can also be used for persistent storage of interfaces.

To use Amino, simply create a codec, and then register types:

```
func NewCodec() *codec.Codec {
	cdc := codec.New()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "example/MsgSend", nil)
	cdc.RegisterConcrete(MsgIssue{}, "example/MsgIssue", nil)
	crypto.RegisterAmino(cdc)
	return cdc
}
```

Note: We also register the types in the `tendermint/tendermint/crypto` module so that `crypto.PubKey`
is encoded/decoded correctly.

Amino supports encoding and decoding in both a binary and JSON format.
See the [codec API docs](https://godoc.org/github.com/tendermint/go-amino#Codec) for more details.

## Tx

Now that we're using Amino, we can embed the `Msg` interface directly in our
`Tx`. We can also add a public key and a signature for authentication.

```go
// Simple tx to wrap the Msg.
type app2Tx struct {
    sdk.Msg

    PubKey    crypto.PubKey
    Signature []byte
}

// This tx only has one Msg.
func (tx app2Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.Msg}
}

// Amino decode app2Tx. Capable of decoding both MsgSend and MsgIssue
func tx2Decoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx app2Tx
		err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode(err.Error())
		}
		return tx, nil
	}
}
```

## AnteHandler

Now that we have an implementation of `Tx` that includes more than just the Msg,
we need to specify how that extra information is validated and processed. This
is the role of the `AnteHandler`. The word `ante` here denotes "before", as the
`AnteHandler` is run before a `Handler`. While an app can have many Handlers,
one for each set of messages, it can have only a single `AnteHandler` that
corresponds to its single implementation of `Tx`.


The AnteHandler resembles a Handler:


```go
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
```

Like Handler, AnteHandler takes a Context that restricts its access to stores
according to whatever capability keys it was granted. Instead of a `Msg`,
however, it takes a `Tx`.

Like Handler, AnteHandler returns a `Result` type, but it also returns a new
`Context` and an `abort bool`.

For `App2`, we simply check if the PubKey matches the Address, and the Signature validates with the PubKey:

```go
// Simple anteHandler that ensures msg signers have signed.
// Provides no replay protection.
func antehandler(ctx sdk.Context, tx sdk.Tx) (_ sdk.Context, _ sdk.Result, abort bool) {
	appTx, ok := tx.(app2Tx)
	if !ok {
		// set abort boolean to true so that we don't continue to process failed tx
		return ctx, sdk.ErrTxDecode("Tx must be of format app2Tx").Result(), true
	}

	// expect only one msg and one signer in app2Tx
	msg := tx.GetMsgs()[0]
	signerAddr := msg.GetSigners()[0]

	signBytes := msg.GetSignBytes()

	sig := appTx.GetSignature()

	// check that submitted pubkey belongs to required address
	if !bytes.Equal(appTx.PubKey.Address(), signerAddr) {
		return ctx, sdk.ErrUnauthorized("Provided Pubkey does not match required address").Result(), true
	}

	// check that signature is over expected signBytes
	if !appTx.PubKey.VerifyBytes(signBytes, sig) {
		return ctx, sdk.ErrUnauthorized("Signature verification failed").Result(), true
	}

	// authentication passed, app to continue processing by sending msg to handler
	return ctx, sdk.Result{}, false
}
```

## App2

Let's put it all together now to get App2:

```go
func NewApp2(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app2Name, logger, db, txDecoder(cdc))

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey(auth.StoreKey)
	// Create a key for accessing the issue store.
	keyIssue := sdk.NewKVStoreKey("issue")

	// set antehandler function
	app.SetAnteHandler(antehandler)

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("send", handleMsgSend(keyAccount)).
		AddRoute("issue", handleMsgIssue(keyAccount, keyIssue))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyIssue)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}
```

The main difference here, compared to `App1`, is that we use a second capability
key for a second store that is *only* passed to a second handler, the
`handleMsgIssue`. The first `handleMsgSend` has no access to this second store and cannot read or write to
it, ensuring a strong separation of concerns.

Note now that we're using Amino, we create a codec, register our types on the codec, and pass the
codec into our TxDecoder constructor, `tx2Decoder`. The SDK takes care of the rest for us!

## Conclusion

We've expanded on our first app by adding a new message type for issuing coins,
and by checking signatures. We learned how to use Amino for decoding into
interface types, allowing us to support multiple Msg types, and we learned how
to use the AnteHandler to validate transactions.

Unfortunately, our application is still insecure, because any valid transaction
can be replayed multiple times to drain someones account! Besides, validating
signatures and preventing replays aren't things developers should have to think
about.

In the next section, we introduce the built-in SDK modules `auth` and `bank`,
which respectively provide secure implementations for all our transaction authentication
and coin transfering needs.
