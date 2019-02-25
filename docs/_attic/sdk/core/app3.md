# Modules

In the previous app, we introduced a new `Msg` type and used Amino to encode
transactions. We also introduced additional data to the `Tx`, and used a simple
`AnteHandler` to validate it.

Here, in `App3`, we introduce two built-in SDK modules to
replace the `Msg`, `Tx`, `Handler`, and `AnteHandler` implementations we've seen
so far: `x/auth` and `x/bank`.

The `x/auth` module implements `Tx` and `AnteHandler` - it has everything we need to
authenticate transactions. It also includes a new `Account` type that simplifies
working with accounts in the store.

The `x/bank` module implements `Msg` and `Handler` - it has everything we need
to transfer coins between accounts.

Here, we'll introduce the important types from `x/auth` and `x/bank`, and use
them to build `App3`, our shortest app yet. The complete code can be found in
[app3.go](examples/app3.go), and at the end of this section.

For more details, see the
[x/auth](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth) and
[x/bank](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank) API documentation.

## Accounts

The `x/auth` module defines a model of accounts much like Ethereum.
In this model, an account contains:

- Address for identification
- PubKey for authentication
- AccountNumber to prune empty accounts
- Sequence to prevent transaction replays
- Coins to carry a balance

Note that the `AccountNumber` is a unique number that is assigned when the account is
created, and the `Sequence` is incremented by one every time a transaction is
sent from the account.

### Account

The `Account` interface captures this account model with getters and setters:

```go
// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error
}
```

Note this is a low-level interface - it allows any of the fields to be over
written. As we'll soon see, access can be restricted using the `Keeper`
paradigm.

### BaseAccount

The default implementation of `Account` is the `BaseAccount`:

```go
// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
type BaseAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	PubKey        crypto.PubKey  `json:"public_key"`
	AccountNumber uint64          `json:"account_number"`
	Sequence      uint64          `json:"sequence"`
}
```

It simply contains a field for each of the methods.

### AccountKeeper

In previous apps using our `appAccount`, we handled
marshaling/unmarshaling the account from the store ourselves, by performing
operations directly on the KVStore. But unrestricted access to a KVStore isn't really the interface we want
to work with in our applications. In the SDK, we use the term `Mapper` to refer
to an abstaction over a KVStore that handles marshalling and unmarshalling a
particular data type to and from the underlying store.

The `x/auth` module provides an `AccountKeeper` that allows us to get and
set `Account` types to the store. Note the benefit of using the `Account`
interface here - developers can implement their own account type that extends
the `BaseAccount` to store additional data without requiring another lookup from
the store.

Creating an AccountKeeper is easy - we just need to specify a codec, a
capability key, and a prototype of the object being encoded

```go
accountKeeper := auth.NewAccountKeeper(cdc, keyAccount, auth.ProtoBaseAccount)
```

Then we can get, modify, and set accounts. For instance, we could double the
amount of coins in an account:

```go
acc := accountKeeper.GetAccount(ctx, addr)
acc.SetCoins(acc.Coins.Plus(acc.Coins))
accountKeeper.SetAccount(ctx, addr)
```

Note that the `AccountKeeper` takes a `Context` as the first argument, and will
load the KVStore from there using the capability key it was granted on creation.

Also note that you must explicitly call `SetAccount` after mutating an account
for the change to persist!

See the [AccountKeeper API
docs](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth#AccountKeeper) for more information.

## StdTx

Now that we have a native model for accounts, it's time to introduce the native
`Tx` type, the `auth.StdTx`:

```go
// StdTx is a standard way to wrap a Msg with Fee and Signatures.
// NOTE: the first signature is the fee payer (Signatures must not be nil).
type StdTx struct {
	Msgs       []sdk.Msg      `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
	Memo       string         `json:"memo"`
}
```

This is the standard form for a transaction in the SDK. Besides the Msgs, it
includes:

- a fee to be paid by the first signer
- replay protecting nonces in the signature
- a memo of prunable additional data

Details on how these components are validated is provided under
[auth.AnteHandler](#antehandler) below.

The standard form for signatures is `StdSignature`:

```go
// StdSignature wraps the Signature and includes counters for replay protection.
// It also includes an optional public key, which must be provided at least in
// the first transaction made by the account.
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	[]byte `json:"signature"`
	AccountNumber    uint64 `json:"account_number"`
	Sequence         uint64 `json:"sequence"`
}
```

The signature includes both an `AccountNumber` and a `Sequence`.
The `Sequence` must match the one in the
corresponding account when the transaction is processed, and will increment by
one with every transaction. This prevents the same
transaction from being replayed multiple times, resolving the insecurity that
remains in App2.

The `AccountNumber` is also for replay protection - it allows accounts to be
deleted from the store when they run out of accounts. If an account receives
coins after it is deleted, the account will be re-created, with the Sequence
reset to 0, but a new AccountNumber. If it weren't for the AccountNumber, the
last sequence of transactions made by the account before it was deleted could be
replayed!

Finally, the standard form for a transaction fee is `StdFee`:

```go
// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount sdk.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}
```

The fee must be paid by the first signer. This allows us to quickly check if the
transaction fee can be paid, and reject the transaction if not.

## Signing

The `StdTx` supports multiple messages and multiple signers.
To sign the transaction, each signer must collect the following information:

- the ChainID
- the AccountNumber and Sequence for the given signer's account (from the
  blockchain)
- the transaction fee
- the list of transaction messages
- an optional memo

Then they can compute the transaction bytes to sign using the
`auth.StdSignBytes` function:

```go
bytesToSign := StdSignBytes(chainID, accNum, accSequence, fee, msgs, memo)
```

Note these bytes are unique for each signer, as they depend on the particular
signers AccountNumber, Sequence, and optional memo. To facilitate easy
inspection before signing, the bytes are actually just a JSON encoded form of
all the relevant information.

## AnteHandler

As we saw in `App2`, we can use an `AnteHandler` to authenticate transactions
before we handle any of their internal messages. While previously we implemented
our own simple `AnteHandler`, the `x/auth` module provides a much more advanced
one that uses `AccountKeeper` and works with `StdTx`:

```go
app.SetAnteHandler(auth.NewAnteHandler(accountKeeper, feeKeeper))
```

The AnteHandler provided by `x/auth` enforces the following rules:

- the memo must not be too big
- the right number of signatures must be provided (one for each unique signer
  returned by `msg.GetSigner` for each `msg`)
- any account signing for the first-time must include a public key in the
  StdSignature
- the signatures must be valid when authenticated in the same order as specified
  by the messages

Note that validating
signatures requires checking that the correct account number and sequence was
used by each signer, as this information is required in the `StdSignBytes`.

If any of the above are not satisfied, the AnteHandelr returns an error.

If all of the above verifications pass, the AnteHandler makes the following
changes to the state:

- increment account sequence by one for all signers
- set the pubkey in the account for any first-time signers
- deduct the fee from the first signer's account

Recall that incrementing the `Sequence` prevents "replay attacks" where
the same message could be executed over and over again.

The PubKey is required for signature verification, but it is only required in
the StdSignature once. From that point on, it will be stored in the account.

The fee is paid by the first address returned by `msg.GetSigners()` for the first `Msg`.

## CoinKeeper

Now that we've seen the `auth.AccountKeeper` and how its used to build a
complete AnteHandler, it's time to look at how to build higher-level
abstractions for taking action on accounts.

Earlier, we noted that `Mappers` are abstactions over KVStores that handle
marshalling and unmarshalling data types to and from underlying stores.
We can build another abstraction on top of `Mappers` that we call `Keepers`,
which expose only limitted functionality on the underlying types stored by the `Mapper`.

For instance, the `x/bank` module defines the canonical versions of `MsgSend`
and `MsgIssue` for the SDK, as well as a `Handler` for processing them. However,
rather than passing a `KVStore` or even an `AccountKeeper` directly to the handler,
we introduce a `bank.Keeper`, which can only be used to transfer coins in and out of accounts.
This allows us to determine up front that the only effect the bank module's
`Handler` can have on the store is to change the amount of coins in an account -
it can't increment sequence numbers, change PubKeys, or otherwise.


A `bank.Keeper` is easily instantiated from an `AccountKeeper`:

```go
bankKeeper = bank.NewBaseKeeper(accountKeeper)
```

We can then use it within a handler, instead of working directly with the
`AccountKeeper`. For instance, to add coins to an account:

```go
// Finds account with addr in AccountKeeper.
// Adds coins to account's coin array.
// Sets updated account in AccountKeeper
app.bankKeeper.AddCoins(ctx, addr, coins)
```

See the [bank.Keeper API
docs](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank#Keeper) for the full set of methods.

Note we can refine the `bank.Keeper` by restricting it's method set. For
instance, the
[bank.ViewKeeper](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank#ViewKeeper)
is a read-only version, while the
[bank.SendKeeper](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank#SendKeeper)
only executes transfers of coins from input accounts to output
accounts.

We use this `Keeper` paradigm extensively in the SDK as the way to define what
kind of functionality each module gets access to. In particular, we try to
follow the *principle of least authority*.
Rather than providing full blown access to the `KVStore` or the `AccountKeeper`,
we restrict access to a small number of functions that do very specific things.

## App3

With the `auth.AccountKeeper` and `bank.Keeper` in hand,
we're now ready to build `App3`.
The `x/auth` and `x/bank` modules do all the heavy lifting:

```go
func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// Create the codec with registered Msg types
	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, logger, db, auth.DefaultTxDecoder(cdc))

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey(auth.StoreKey)
	keyFees := sdk.NewKVStoreKey(auth.FeeStoreKey)  // TODO

	// Set various mappers/keepers to interact easily with underlying stores
	accountKeeper := auth.NewAccountKeeper(cdc, keyAccount, auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountKeeper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to
	app.Router().
		AddRoute("send", bank.NewHandler(bankKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}
```

Note we use `bank.NewHandler`, which handles only `bank.MsgSend`,
and receives only the `bank.Keeper`. See the
[x/bank API docs](https://godoc.org/github.com/cosmos/cosmos-sdk/x/bank)
for more details.

We also use the default txDecoder in `x/auth`, which decodes amino-encoded
`auth.StdTx` transactions.

## Conclusion

Armed with native modules for authentication and coin transfer,
emboldened by the paradigm of mappers and keepers,
and ever invigorated by the desire to build secure state-machines,
we find ourselves here with a full-blown, all-checks-in-place, multi-asset
cryptocurrency - the beating heart of the Cosmos-SDK.
