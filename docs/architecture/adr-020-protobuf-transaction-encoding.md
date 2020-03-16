# ADR 020: Protocol Buffer Transaction Encoding

## Changelog

- 2020 March 06: Initial Draft
- 2020 March 12: API Updates

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

Since the messages that an application is known and allowed to handle are specific
to the application itself, so must the transactions be specific to the application
itself. Similar to how we described in [ADR 019](./adr-019-protobuf-state-encoding.md),
the concrete types will be defined at the application level via Protobuf `oneof`.

The application will define a single canonical `Message` Protobuf message
with a single `oneof` that implements the SDK's `Msg` interface.

Example:

```protobuf
// app/codec/codec.proto

message Message {
  option (cosmos_proto.interface_type) = "github.com/cosmos/cosmos-sdk/types.Msg";

  oneof sum {
    cosmos_sdk.x.bank.v1.MsgSend              msg_send             = 1;
    cosmos_sdk.x.bank.v1.MsgMultiSend         msg_multi_send       = 2;
    cosmos_sdk.x.crisis.v1.MsgVerifyInvariant msg_verify_invariant = 3;
    // ...
  }
}
```

Because an application needs to define it's unique `Message` Protobuf message, it
will by proxy have to define a `Transaction` Protobuf message that encapsulates this
`Message` type. The `Transaction` message type must implement the SDK's `Tx` interface.

Example:

```protobuf
// app/codec/codec.proto

message Transaction {
  cosmos_sdk.x.auth.v1.StdTxBase base = 1;
  repeated Message               msgs = 2;
}
```

Note, the `Transaction` type includes `StdTxBase` which will be defined by the SDK
and includes all the core field members that are common across all transaction types.
Developers do not have to include `StdTxBase` if they wish, so it is meant to be
used as an auxiliary type.

### Signing

Signing of a `Transaction` must be canonical across clients and binaries. In order
to provide canonical representation of a `Transaction` to sign over, clients must
obey the following rules:

- Encode `SignDoc` (see below) via [Protobuf's canonical JSON encoding](https://developers.google.com/protocol-buffers/docs/proto3#json).
  - Default must be stripped from the output!
  - JSON keys adhere to their Proto-defined field names.
- Generate canonical JSON to sign via the [JSON Canonical Form Spec](https://gibson042.github.io/canonicaljson-spec/).
  - This spec should be trivial to interpret and implement in any language.

```Protobuf
// app/codec/codec.proto

message SignDoc {
  StdSignDocBase base = 1;
  repeated Message msgs = 2;
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

We then update `CLIContext` to have two new fields: `Generator` and `Marshler`.

Then, each module will at the minimum accept a `Marshaler` instead of a concrete
Amino codec. If the module needs to work with any interface types, it will use
the `Codec` interface defined by the module which also extends `Marshaler`.

## Future Improvements

Requiring application developers to have to redefine their `Message` Protobuf types
can be extremely tedious and may increase the surface area of bugs by potentially
missing one or more messages in the `oneof`.

To circumvent this, an optional strategy can be taken that has each module define
it's own `oneof` and then the application-level `Message` simply imports each module's
`oneof`. However, this requires additional tooling and the use of reflection.

Example:

```protobuf
// app/codec/codec.proto

message Message {
  option (cosmos_proto.interface_type) = "github.com/cosmos/cosmos-sdk/types.Msg";

  oneof sum {
    bank.Msg = 1;
    staking.Msg = 2;
    // ...
  }
}
```

## Consequences

### Positive

- Significant performance gains.
- Supports backward and forward type compatibility.
- Better support for cross-language clients.

### Negative

- Learning curve required to understand and implement Protobuf messages.
- Less flexibility in cross-module type registration. We now need to define types
at the application-level.
- Client business logic and tx generation become a bit more complex as developers
have to define more types and implement more interfaces.

### Neutral

## References
