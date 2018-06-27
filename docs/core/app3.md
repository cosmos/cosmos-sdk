<<<<<<< HEAD
# Authentication

In the previous app, we introduced a new `Msg` type and used Amino to encode
transactions. In that example, our `Tx` implementation was still just a simple
wrapper of the `Msg`, providing no actual authentication. Here, in `App3`, we
expand on `App2` to provide real authentication in the transactions.

Without loss of generality, the SDK prescribes native 
account and transaction types that are sufficient for a wide range of applications. 
These are implemented in the `x/auth` module, where
all authentication related data structures and logic reside. 
Applications that use `x/auth` don't need to worry about any of the details of
authentication and replay protection, as they are handled automatically. For
completeness, we will explain everything here.

## Account

The `Account` interface provides a model of accounts that have:
=======
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

Applications that use `x/auth` and `x/bank` thus significantly reduce the amount 
of work they have to do so they can focus on their application specific logic in
their own modules.

Here, we'll introduce the important types from `x/auth` and `x/bank`, and show
how to make `App3` by using them. The complete code can be found in [app3.go](examples/app3.go).
For more details, see the [x/auth](TODO) and [x/bank](TODO) API documentation.

## Accounts

The `x/auth` module defines a model of accounts much like Ethereum.
In this model, an account contains:
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

- Address for identification
- PubKey for authentication
- AccountNumber to prune empty accounts
- Sequence to prevent transaction replays
- Coins to carry a balance

<<<<<<< HEAD
It consists of getters and setters for each of these:
=======
Note that the `AccountNumber` is a unique number that is assigned when the account is
created, and the `Sequence` is incremented by one every time a transaction is
sent from the account.

### Account

The `Account` interface captures this account model with getters and setters:
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

```go
// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() sdk.Address
	SetAddress(sdk.Address) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() int64
	SetAccountNumber(int64) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error
}
```

<<<<<<< HEAD
## BaseAccount
=======
Note this is a low-level interface - it allows any of the fields to be over
written. As we'll soon see, access can be restricted using the `Keeper`
paradigm.

### BaseAccount
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

The default implementation of `Account` is the `BaseAccount`:

```go
// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address       sdk.Address   `json:"address"`
	Coins         sdk.Coins     `json:"coins"`
	PubKey        crypto.PubKey `json:"public_key"`
	AccountNumber int64         `json:"account_number"`
	Sequence      int64         `json:"sequence"`
}
```

It simply contains a field for each of the methods.

<<<<<<< HEAD
The `Address`, `PubKey`, and `AccountNumber` of the `BaseAccpunt` cannot be changed once they are set.

The `Sequence` increments by one with every transaction. This ensures that a
given transaction can only be executed once, as the `Sequence` contained in the
transaction must match that contained in the account.

The `Coins` will change according to the logic of each transaction type.

If the `Coins` are ever emptied, the account will be deleted from the store. If
coins are later sent to the same `Address`, the account will be recreated but
with a new `AccountNumber`. This allows us to prune empty accounts from the
store, while still preventing transaction replay if accounts become non-empty
again in the future.



## StdTx

The standard way to create a transaction from a message is to use the `StdTx` struct defined in the `x/auth` module:
=======
### AccountMapper

In previous apps using our `appAccount`, we handled
marshaling/unmarshaling the account from the store ourselves, by performing
operations directly on the KVStore. But unrestricted access to a KVStore isn't really the interface we want
to work with in our applications. In the SDK, we use the term `Mapper` to refer
to an abstaction over a KVStore that handles marshalling and unmarshalling a
particular data type to and from the underlying store.

The `x/auth` module provides an `AccountMapper` that allows us to get and
set `Account` types to the store. Note the benefit of using the `Account`
interface here - developers can implement their own account type that extends
the `BaseAccount` to store additional data without requiring another lookup from
the store.

Creating an AccountMapper is easy - we just need to specify a codec, a
capability key, and a prototype of the object being encoded (TODO: change to
constructor):

```go
accountMapper := auth.NewAccountMapper(cdc, keyAccount, &auth.BaseAccount{})
```

Then we can get, modify, and set accounts. For instance, we could double the
amount of coins in an account:

```go
acc := GetAccount(ctx, addr)
acc.SetCoins(acc.Coins.Plus(acc.Coins))
acc.SetAccount(ctx, addr)
```

Note that the `AccountMapper` takes a `Context` as the first argument, and will
load the KVStore from there using the capability key it was granted on creation.

Also note that you must explicitly call `SetAccount` after mutating an account
for the change to persist!

See the [AccountMapper API docs](TODO) for more information.

## StdTx

Now that we have a native model for accounts, it's time to introduce the native
`Tx` type, the `auth.StdTx`:
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

```go
// StdTx is a standard way to wrap a Msg with Fee and Signatures.
// NOTE: the first signature is the FeePayer (Signatures must not be nil).
type StdTx struct {
	Msgs       []sdk.Msg      `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
	Memo       string         `json:"memo"`
}
```

<<<<<<< HEAD
The `StdTx` includes a list of messages, information about the fee being paid,
and a list of signatures. It also includes an optional `Memo` for additional
data. Note that the list of signatures must match the result of `GetSigners()`
for each `Msg`!

The signatures are provided in a standard form as `StdSignature`:
=======
This is the standard form for a transaction in the SDK. Besides the Msgs, it
includes:

- a fee to be paid by the first signer 
- replay protecting nonces in the signature
- a memo of prunable additional data

Details on how these components are validated is provided under
[auth.AnteHandler](#antehandler) below.

The standard form for signatures is `StdSignature`:
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

```go
// StdSignature wraps the Signature and includes counters for replay protection.
// It also includes an optional public key, which must be provided at least in
// the first transaction made by the account.
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	crypto.Signature `json:"signature"`
	AccountNumber    int64 `json:"account_number"`
	Sequence         int64 `json:"sequence"`
}
```

<<<<<<< HEAD
Recall that the `Sequence` is expected to increment every time a
message is signed by a given account in order to prevent "replay attacks" where
the same message could be executed over and over again. The `AccountNumber` is
assigned when the account is created or recreated after being emptied.

The `StdSignature` can also optionally include the public key for verifying the
signature. The public key only needs to be included the first time a transaction
is sent from a given account - from then on it will be stored in the `Account`
and can be left out of transactions.

The fee is provided in a standard form as `StdFee`:
=======
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
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514

```go
// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount sdk.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}
```

<<<<<<< HEAD
Note that the address responsible for paying the transactions fee is the first address
returned by msg.GetSigners() for the first `Msg`. The convenience function `FeePayer(tx Tx)` is provided
to return this.

## Signing

The standard bytes for signers to sign over is provided by:

```go
TODO
```

## AnteHandler

The AnteHandler is used to do all transaction-level processing (i.e. Fee payment, signature verification) 
before passing the message to its respective handler.

```go
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
```

The antehandler takes a Context and a transaction and returns a new Context, a Result, and the abort boolean.
As with the handler, all information necessary for processing a message should be available in the
context.

If the transaction fails, then the application should not waste time processing the message. Thus, the antehandler should
return an Error's Result method and set the abort boolean to `true` so that the application knows not to process the message in a handler.

Most applications can use the provided antehandler implementation in `x/auth` which handles signature verification
as well as collecting fees.

Note: Signatures must be over `auth.StdSignDoc` introduced above to use the provided antehandler.

```go
// File: cosmos-sdk/examples/basecoin/app/app.go
app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
```

### Handling Fee payment
### Handling Authentication

The antehandler is responsible for handling all authentication of a transaction before passing the message onto its handler.
This generally involves signature verification. The antehandler should check that all of the addresses that are returned in
`tx.GetMsg().GetSigners()` signed the message and that they signed over `tx.GetMsg().GetSignBytes()`.

# Accounts 

### auth.Account

### auth.AccountMapper

```go
// This AccountMapper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type AccountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical Account concrete type.
	proto Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}
```

The AccountMapper is responsible for managing and storing the state of all accounts in the application.

Example Initialization:

```go
// File: examples/basecoin/app/app.go
// Define the accountMapper.
app.accountMapper = auth.NewAccountMapper(
	cdc,
	app.keyAccount,      // target store
	&types.AppAccount{}, // prototype
)
```

The accountMapper allows you to retrieve the current account state by `GetAccount(ctx Context, addr auth.Address)` and change the state by 
`SetAccount(ctx Context, acc Account)`.

Note: To update an account you will first have to get the account, update the appropriate fields with its associated setter method, and then call
`SetAccount(ctx Context, acc updatedAccount)`.

Updating accounts is made easier by using the `Keeper` struct in the `x/bank` module.

Example Initialization:

```go
// File: examples/basecoin/app/app.go
app.coinKeeper = bank.NewKeeper(app.accountMapper)
```

Example Usage:

```go
// Finds account with addr in accountmapper
// Adds coins to account's coin array
// Sets updated account in accountmapper
app.coinKeeper.AddCoins(ctx, addr, coins)
```

=======
The fee must be paid by the first signer. This allows us to quickly check if the
transaction fee can be paid, and reject the transaction if not.

## Signing

The `StdTx` supports multiple messages and multiple signers.
To sign the transaction, each signer must collect the following information:

- the ChainID (TODO haven't mentioned this yet)
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
one that uses `AccountMapper` and works with `StdTx`:

```go
TODO: feekeeper :(
app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))
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
The convenience function `FeePayer(tx Tx) sdk.Address` is provided to return this.

## CoinKeeper

Now that we've seen the `auth.AccountMapper` and how its used to build a
complete AnteHandler, it's time to look at how to build higher-level
abstractions for taking action on accounts.

Earlier, we noted that `Mappers` abstactions over a KVStore that handles marshalling and unmarshalling a
particular data type to and from the underlying store. We can build another
abstraction on top of `Mappers` that we call `Keepers`, which expose only
limitted functionality on the underlying types stored by the `Mapper`.

For instance, the `x/bank` module defines the canonical versions of `MsgSend`
and `MsgIssue` for the SDK, as well as a `Handler` for processing them. However, 
rather than passing a `KVStore` or even an `AccountMapper` directly to the handler,
we introduce a `bank.Keeper`, which can only be used to transfer coins in and out of accounts.
This allows us to determine up front that the only effect the bank module's
`Handler` can have on the store is to change the amount of coins in an account -
it can't increment sequence numbers, change PubKeys, or otherwise.


A `bank.Keeper` is easily instantiated from an `AccountMapper`:

```go
coinKeeper = bank.NewKeeper(accountMapper)
```

We can then use it within a handler, instead of working directly with the
`AccountMapper`. For instance, to add coins to an account:

```go
// Finds account with addr in AccountMapper.
// Adds coins to account's coin array.
// Sets updated account in AccountMapper
app.coinKeeper.AddCoins(ctx, addr, coins)
```

See the [bank.Keeper API docs](TODO) for the full set of methods.

Note we can refine the `bank.Keeper` by restricting it's method set. For
instance, the `bank.ViewKeeper` is a read-only version, while the
`bank.SendKeeper` only executes transfers of coins from input accounts to output
accounts.

We use this `Keeper` paradigm extensively in the SDK as the way to define what
kind of functionality each module gets access to. In particular, we try to
follow the *principle of least authority*, where modules only get access to the
absolutely narrowest set of functionality they need to get the job done. Hence,
rather than providing full blown access to the `KVStore` or the `AccountMapper`, 
we restrict access to a small number of functions that do very specific things.

## App3

Armed with an understanding of mappers and keepers, in particular the
`auth.AccountMapper` and the `bank.Keeper`, we're now ready to build `App3`
using the `x/auth` and `x/bank` modules to do all the heavy lifting:

```go
TODO
```
>>>>>>> 6bbe295d7fd75fc3693f68915519b483bdd4e514
