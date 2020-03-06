# ADR 020: Protocol Buffer Client Encoding

## Changelog

- 2020 March 06: Initial Draft

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
querying.

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
    bank.MsgSend = 1;
    staking.MsgCreateValidator = 2;
    staking.MsgDelegate = 3;
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
  option (cosmos_proto.interface_type) = "github.com/cosmos/cosmos-sdk/types.Tx";

  StdTxBase base = 1;
  repeated Message msgs = 2;
}
```

Note, the `Transaction` type includes `StdTxBase` which will be defined by the SDK
and includes all the core field members that are common across all transaction types.
Developers do not have to include `StdTxBase` if they wish, so it is meant to be
used as an auxiliary type.

### Signing

### CLI, REST, & Querying

## Consequences

### Positive

### Negative

### Neutral

## References
