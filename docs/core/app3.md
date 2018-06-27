# Modules

In the previous app, we introduced a new `Msg` type and used Amino to encode
transactions. We also introduced additional data to the `Tx`, and used a simple 
`AnteHandler` to validate it. 

Here, in `App3`, we introduce two built-in SDK modules to
replace the `Msg`, `Tx`, `Handler`, and `AnteHandler` implementations we've seen
so far. 

The `x/auth` module implements `Tx` and `AnteHandler - it has everything we need to
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


### AccountMapper

TODO

## Transaction


### StdTx

The standard way to create a transaction from a message is to use the `StdTx` struct defined in the `x/auth` module:

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

The `StdTx` includes a list of messages, information about the fee being paid,
and a list of signatures. It also includes an optional `Memo` for additional
data. Note that the list of signatures must match the result of `GetSigners()`
for each `Msg`!

The signatures are provided in a standard form as `StdSignature`:

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

Recall that the `Sequence` is expected to increment every time a
message is signed by a given account in order to prevent "replay attacks" where
the same message could be executed over and over again. The `AccountNumber` is
assigned when the account is created or recreated after being emptied.

The `StdSignature` can also optionally include the public key for verifying the
signature. The public key only needs to be included the first time a transaction
is sent from a given account - from then on it will be stored in the `Account`
and can be left out of transactions.

The fee is provided in a standard form as `StdFee`:

```go
// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount sdk.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}
```

Note that the address responsible for paying the transactions fee is the first address
returned by msg.GetSigners() for the first `Msg`. The convenience function `FeePayer(tx Tx)` is provided
to return this.

### Signing

The standard bytes for signers to sign over is provided by:

```go
TODO
```

## AnteHandler

TODO

## App3

Putting it all together, we get:

```go
TODO
```
