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

Note this is a low-level interface - it allows any of the fields to be over
written. As we'll soon see, access can be restricted using the `Keeper`
paradigm.

### BaseAccount

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
acc := GetAccount(ctx, addr)` 
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

This is the standard form for a transaction in the SDK. Besides the Msgs, it
includes:

- a fee to be paid by the first signer 
- replay protecting nonces in the signature
- a memo of prunable additional data

Details on how these components are validated is provided under
[auth.AnteHandler](#ante-handler) below.

The standard form for signatures is `StdSignature`:

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

If any of the above are not satisfied, it returns an error. 

If all of the above verifications pass, the AnteHandler makes the following
changes to the state:

- increment account sequence by one for all signers
- set the pubkey for any first-time signers
- deduct the fee from the first signer

Recall that incrementing the `Sequence` prevents "replay attacks" where
the same message could be executed over and over again. 

The PubKey is required for signature verification, but it is only required in
the StdSignature once. From that point on, it will be stored in the account.

The fee is paid by the first address returned by msg.GetSigners() for the first `Msg`. 
The convenience function `FeePayer(tx Tx) sdk.Address` is provided to return this.

## CoinKeeper

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


## App3

Putting it all together, we get:

```go
TODO
```
