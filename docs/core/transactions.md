### Transactions

A message is a set of instructions for a state transition.

For a message to be valid, it must be accompanied by at least one 
digital signature. The signatures required are determined solely 
by the contents of the message.

A transaction is a message with additional information for authentication:

```go
type Tx interface {

	GetMsg() Msg

}
```

The standard way to create a transaction from a message is to use the `StdTx` struct defined in the `x/auth` module.

```go
type StdTx struct {
	Msg        sdk.Msg        `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
}
```

The `StdTx.GetSignatures()` method returns a list of signatures, which must match
the list of addresses returned by `tx.Msg.GetSigners()`. The signatures come in
a standard form:

```go
type StdSignature struct {
	crypto.PubKey // optional
	crypto.Signature
	AccountNumber int64
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

The standard bytes for signers to sign over is provided by:

```go
func StdSignByes(chainID string, accnums []int64, sequences []int64, fee StdFee, msg sdk.Msg) []byte
```

in `x/auth`. The standard way to construct fees to pay for the processing of transactions is:

```go
// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount sdk.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}
```

### Encoding and Decoding Transactions

Messages and transactions are designed to be generic enough for developers to
specify their own encoding schemes.  This enables the SDK to be used as the
framwork for constructing already specified cryptocurrency state machines, for
instance Ethereum. 

When initializing an application, a developer can specify a `TxDecoder`
function which determines how an arbitrary byte array should be unmarshalled
into a `Tx`: 

```go
type TxDecoder func(txBytes []byte) (Tx, error)
```

The default tx decoder is the Tendermint wire format which uses the go-amino library
for encoding and decoding all message types.

In `Basecoin`, we use the default transaction decoder.  The `go-amino` library has the nice
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

If you wish to use a custom encoding scheme, you must define a TxDecoder function
and set it as the decoder in your extended baseapp using the `SetTxDecoder(decoder sdk.TxDecoder)`.

Ex:

```go
app.SetTxDecoder(CustomTxDecodeFn)
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
