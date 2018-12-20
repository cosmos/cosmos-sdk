## Types

Besides accounts (specified in [State](state.md)), the types exposed by the auth module
are `StdFee`, the combination of an amount and gas limit, `StdSignature`, the combination
of an optional public key and a cryptographic signature as a byte array, `StdTx`,
a struct which implements the `sdk.Tx` interface using `StdFee` and `StdSignature`, and
`StdSignDoc`, a replay-prevention structure for `StdTx` which transaction senders must sign over.

### StdFee

A `StdFee` is simply the combination of a fee amount, in any number of denominations,
and a gas limit (where dividing the amount by the gas limit gives a "gas price").

```golang
type StdFee struct {
  Amount Coins
  Gas    uint64
}
```

### StdSignature

A `StdSignature` is the combination of an optional public key and a cryptographic signature
as a byte array. The SDK is agnostic to particular key or signature formats and supports any
supported by the `PubKey` interface.

```golang
type StdSignature struct {
  PubKey    PubKey
  Signature []byte
}
```

### StdTx

A `StdTx` is a struct which implements the `sdk.Tx` interface, and is likely to be generic
enough to serve the purposes of many Cosmos SDK blockchains.

```golang
type StdTx struct {
  Msgs        []sdk.Msg
  Fee         StdFee  
  Signatures  []StdSignature
  Memo        string
}
```

### StdSignDoc

A `StdSignDoc` is a replay-prevention structure to be signed over, which ensures that
any submitted transaction (which is simply a signature over a particular bytestring)
will only be executable once on a particular blockchain.

`json.RawMessage` is preferred over using the SDK types for future compatibility.

```golang
type StdSignDoc struct {
  AccountNumber uint64
  ChainID       string
  Fee           json.RawMessage
  Memo          string
  Msgs          []json.RawMessage
  Sequence      uint64
}
```
