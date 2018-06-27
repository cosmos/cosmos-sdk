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

- Address for identification
- PubKey for authentication
- AccountNumber to prune empty accounts
- Sequence to prevent transaction replays
- Coins to carry a balance

It consists of getters and setters for each of these:

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

## BaseAccount

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



## StdTx

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

