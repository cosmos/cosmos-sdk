# ADR 020: Protocol Buffer Transaction Encoding

## Changelog

- 2020 March 06: Initial Draft
- 2020 March 12: API Updates
- 2020 April 13: Added details on interface `oneof` handling
- 2020 April 30: Switch to `Any`

## Status

Proposed

## Context

This ADR is a continuation of the motivation, design, and context established in
[ADR 019](./adr-019-protobuf-state-encoding.md), namely, we aim to design the
Protocol Buffer migration path for the client-side of the Cosmos SDK.

Specifically, the client-side migration path primarily includes tx generation and
signing, message construction and routing, in addition to CLI & REST handlers and
business logic (i.e. queriers).

With this in mind, we will tackle the migration path via two main areas, txs and
querying. However, this ADR solely focuses on transactions. Querying should be
addressed in a future ADR, but it should build off of these proposals.

## Decision

### Transactions

Since are encoding interface values using `google.protobuf.any` (see [ADR 019](adr-019-protobuf-state-encoding.md)),
`sdk.Msg`s are encoding with `Any` in transactions. One of the primary goals of
using `Any` to encode interface values which vary from chain to chain is to have
a core set of types which is reused by apps so that clients can safely be
compatible with as many chains as possible. It is one of the goals of this
specification to provide a flexible cross-chain transaction format that can
serve a wide variety of use cases without breaking compatibility.

In order to facilitate signing, transactions are separated into a body (`TxBody`),
which will be re-used by `SignDoc` below, and `Signature`s:

```proto
// types/types.proto
package cosmos_sdk.v1;

message Tx {
    TxBody body = 1;
    repeated Signature signatures = 2;
}

message TxBody {
    repeated google.protobuf.Any messages = 1;
    Fee fee = 2;
    string memo = 3;
    int64 timeout_height = 4;
    repeated google.protobuf.Any extension_options = 5;
}
```

Because we are aiming for be a flexible, extensible cross-chain transaction
format, new transaction processing options should be added to `TxBody` as those
use cases are discovered. Because there is coordination overhead in this,
however, `TxBody` includes an `extension_options` field which can be used by apps to
add custom transaction processing options that have not yet been upstreamed
into the canonical `TxBody` as fields. Apps may use this field as necessary
using their own processing middleware, but _should_ attempt to upstream useful
features to core SDK .proto files even if the SDK does not yet support these
options.

### Signing

```proto
// types/types.proto

message SignDoc {
    TxBody body = 1;
    string chain_id = 2;
    uint64 account_number = 3;
    uint64 account_sequence = 4;
}

message Signature {
    PublicKey public_key = 1;
    bytes signature = 2;
    SignMode mode = 3;
}

enum SignMode {
    SIGN_MODE_DEFAULT = 0;
    SIGN_MODE_EXTENDED = 1;
    SIGN_MODE_LEGACY_AMINO = 1024;
}
```

### CLI & REST

Currently, the REST and CLI handlers encode and decode types and txs via Amino
JSON encoding using a concrete Amino codec. Being that some of the types dealt with
in the client can be interfaces, similar to how we described in [ADR 019](./adr-019-protobuf-state-encoding.md),
the client logic will now need to take a codec interface that knows not only how
to handle all the types, but also knows how to generate transactions, signatures,
and messages.

```go
type AccountRetriever interface {
  EnsureExists(addr sdk.AccAddress) error
  GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error)
}

type Generator interface {
  NewTx() ClientTx
}

type ClientTx interface {
  sdk.Tx
  codec.ProtoMarshaler

  SetMsgs(...sdk.Msg) error
  GetSignatures() []sdk.Signature
  SetSignatures(...sdk.Signature)
  GetFee() sdk.Fee
  SetFee(sdk.Fee)
  GetMemo() string
  SetMemo(string)

  CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error)
}
```

We then update `CLIContext` to have a new field: `Marshaler`.

Then, each module's client handler will at the minimum accept a `Marshaler` instead
of a concrete Amino codec and a `Generator` along with an `AccountRetriever` so
that account fields can be retrieved for signing.


## Future Improvements

## Consequences

### Positive

- Significant performance gains.
- Supports backward and forward type compatibility.
- Better support for cross-language clients.

### Negative

- `google.protobuf.Any` type URLs increase transaction size although the effect
may be negligible or compression may be able to mitigate it.

### Neutral

## References
