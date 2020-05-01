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

Based on detailed discussions ([\#6030](https://github.com/cosmos/cosmos-sdk/issues/6030)
and [\#6078](https://github.com/cosmos/cosmos-sdk/issues/6078)), the original 
design for transactions was changed substantially from an `oneof` /JSON-signing
approach to the approach described below.

## Decision

### Transactions

Since interface values are encoded with `google.protobuf.Any` in state (see [ADR 019](adr-019-protobuf-state-encoding.md)),
 `sdk.Msg`s are encoding with `Any` in transactions.
 
 One of the main goals of using `Any` to encode interface values is to have a
 core set of types which is reused by apps so that
 clients can safely be compatible with as many chains as possible.
 
 It is one of the goals of this specification to provide a flexible cross-chain transaction
 format that can serve a wide variety of use cases without breaking client 
 compatibility.

In order to facilitate signing, transactions are separated into `TxBody`,
which will be re-used by `SignDoc` below, and `signatures`:

```proto
// types/types.proto
package cosmos_sdk.v1;

message Tx {
    TxBody body = 1;
    repeated bytes signatures = 2;
}

message TxBody {
    repeated google.protobuf.Any messages = 1;
    repeated SignerInfo signer_info = 2;
    Fee fee = 3;
    string memo = 4;
    int64 timeout_height = 5;
    repeated google.protobuf.Any extension_options = 6;
}

message SignerInfo {
    PublicKey pub_key = 1;
    SignMode mode = 2;
}

enum SignMode {
    SIGN_MODE_DEFAULT = 0;
    SIGN_MODE_LEGACY_AMINO = -1;
    SIGN_MODE_EXTENDED = 1;
}
```

As will be discussed below, in order to include as much of the `Tx` as possible
in the `SignDoc`, `SignerInfo` is separated from signatures so that only the
raw signatures themselves live outside of `TxBody`.

Because we are aiming for be a flexible, extensible cross-chain transaction
format, new transaction processing options should be added to `TxBody` as soon
those use cases are discovered, even if they can't be implemented yet.

Because there is coordination overhead in this, `TxBody` includes an
`extension_options` field which can be used for any transaction processing
options that are not already covered. App developers should, nevertheless,
attempt to upstream important improvements to `Tx`.

### Signing

Signatures are structured using the `SignDoc` below which reuses `TxBody` and only
adds the fields which are needed for signatures but not present on `TxBody`:

```proto
// types/types.proto
message SignDoc {
    TxBody body = 1;
    string chain_id = 2;
    uint64 account_number = 3;
    uint64 account_sequence = 4;
}
```

#### `SIGN_MODE_DEFAULT`

The default signing behavior is to sign the raw `TxBody` bytes as broadcast over
the wire. This has the advantages of:

* requiring the minimum additional client capabilities beyond a standard protocol
buffers implementation
* leaving effectively zero holes for transaction malleability (i.e. there are no
subtle differences between the signing and encoding formats which could 
potentially be exploited by an attacker)

In order to sign in the default mode, clients take the following steps:

1. Encode `TxBody`

2. Sign `SignDocRaw`

The raw encoded `TxBody` bytes are encoded into `SignDocRaw` below so that the
encoded body bytes exactly match the signed body bytes with no need for
["canonicalization"](https://github.com/regen-network/canonical-proto3) of that
message.

```proto
// types/types.proto
message SignDocRaw {
    bytes body_bytes = 1;
    string chain_id = 2;
    uint64 account_number = 3;
    uint64 account_sequence = 4;
}
```

The only requirements are that the client _must_ encode `SignDocRaw` itself canonically.
 This means that:
 
* all of the fields must be encoded in order
* default values (i.e. empty/zero values) must be omitted 

If a protobuf implementation does not by default encode `SignDocRaw` canonically,
the client _must_ manually encode `SignDocRaw` following the guidelines above.

Again, it does not matter if `TxBody` was encoded canonically or not.

Note that in order to decode `SignDocRaw` for displaying contents to users,
the regular `SignDoc` type should be used.

3. Broadcast `TxRaw`

In order to make sure that the signed body bytes exactly match the encoded body
bytes, clients should encode and broadcast `TxRaw` with the same body bytes used
in `SignDocRaw`.

```proto
// types/types.proto
message TxRaw {
    bytes body_bytes = 1;
    repeated bytes signatures = 2;
}
```

Signature verifiers should verify signatures by decoding `TxRaw` and then
encoding `SignDocRaw` with the raw body bytes.

The standard `Tx` type (which is byte compatible with `TxRaw`) can be used to
decode transactions for all other use cases.

#### `SIGN_MODE_LEGACY_AMINO`

In order to support legacy wallets and exchanges, Amino JSON will be emporarily
supported transaction signing. Once wallets and exchanges have had a
chance to upgrade to protobuf based signing, this option will be disabled. In
the meantime, it is foreseen that disabling the current Amino signing would cause
too much breakage to be feasible.

Legacy clients will be able to sign a transaction using the current Amino
JSON format and have it encoded to protobuf using the REST `/tx/encode`
endpoint before broadcasting.

#### `SIGN_MODE_EXTENDED`

As was discussed extensively in [\#6078](https://github.com/cosmos/cosmos-sdk/issues/6078),
there is a desire for a human-readable signing encoding, especially for hardware
wallets like the [Ledger](https://www.ledger.com) which display
transaction contents to users before signing. JSON was an attempt at this but 
falls short of the ideal.

`SIGN_MODE_EXTENDED` is intended as a placeholder for a human-readable
encoding which will replace Amino JSON. This new encoding should be even more
focused on readability than JSON, possibly based on formatting strings like
[MessageFormat](http://userguide.icu-project.org/formatparse/messages).

In order to ensure that the new human-readable format does not suffer from
transaction malleability issues, `SIGN_MODE_EXTENDED`
requires that the _human-readable bytes are concatenated with the raw `SignDoc`_
to generate sign bytes.

Multiple human-readable formats (maybe even localized messages) may be supported
by `SIGN_MODE_EXTENDED` when it is implemented.

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

A concrete implementation of `SIGN_MODE_EXTENDED` is intended as a near-term
future improvement.

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
