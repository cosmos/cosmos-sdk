# ADR 020: Protocol Buffer Transaction Encoding

## Changelog

- 2020 March 06: Initial Draft
- 2020 March 12: API Updates
- 2020 April 13: Added details on interface `oneof` handling
- 2020 April 20: Describe improved signing/transaction-submission UX

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

### Transaction Encoding

Since the messages that an application knows and is allowed to handle are specific
to the application itself, so must the transaction encoding type be specific to the application
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
  cosmos_sdk.codec.std.v1.StdTxBase base = 1;
  repeated Message               msgs = 2;
}
```

Note, the `Transaction` type includes `StdTxBase` which will be defined by the SDK
and includes all the core field members that are common across all transaction types.
Developers do not have to include `StdTxBase` if they wish, so it is meant to be
used as an auxiliary type.

### Signing & Transaction Submission

To provide an acceptable user experience for client-side developers, we separate
transaction encoding and signing. While every app defines its own transaction
types for encoding purposes, we define a single standard signing `SignDoc` to be used
across apps that leverages `google.protobuf.Any` to handle interface types. `Any`
can be used to wrap any protobuf message by encoding both its type URL and value.
We avoid using `Any` during the encoding phase because these URLs can be long
and we want to keep encoded messages small. We are not, however, sending the
`SignDoc` (encoded in JSON) over the wire or storing it on nodes,
therefore using `Any` for signing does not add any encoding overhead.

```proto
message SignDoc {
  StdSignDocBase base = 1;
  repeated google.protobuf.Any msgs = 2;
}
```

Signing of a `SignDoc` must be canonical across clients and binaries. In order
to provide canonical representation of a `SignDoc` to sign over, clients must
obey the following rules:

- Encode `SignDoc` via [Protobuf's canonical JSON encoding](https://developers.google.com/protocol-buffers/docs/proto3#json).
  - Default must be stripped from the output!
  - JSON keys adhere to their Proto-defined field names.
- Generate canonical JSON to sign via the [JSON Canonical Form Spec](https://gibson042.github.io/canonicaljson-spec/).
  - This spec should be trivial to interpret and implement in any language.

Because signing and encoding are separated and apps will know how to encode
a signing messages, we can provide a gRPC/REST endpoint that allows for generic
app-independent transaction submission. This can be done via a generic
`AnyTransaction` type:

```proto
message AnyTransaction {
  cosmos_sdk.codec.std.v1.StdTxBase base = 1;
  repeated google.protobuf.Any msgs = 2;
}
```

On the backend, this will be converted to the app-specific `Transaction` type,
with any errors returned for unsupported `Msg`s. Client developers that prefer
to use the app-specific `Transaction` and `Message` `oneof`s can also do this
encoding on their own. This provides additional type safety where needed. There
is, however, no security disadvantage for using the generic RPC transaction
endpoint because all clients regardless of whether they have access to the
app-level proto files will know how to correctly sign transactions. This allows
for wallets and block explorers to easily add support for new chains, even
discovering them dynamically.

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

### Interface `oneof` Handling

If a module needs to work with `sdk.Msg`s that use interface types, those interface
types should be dealt with using `google.protobuf.Any` for signing and a `oneof`
at the app-level for encoding.

Using `google.protobuf.Any` for signing will allow client libraries to create
reusable code for dealing with interface types that doesn't require regenerating
protobuf client libraries for every chain. Using `oneof`s at the encoding level
saves space in the Tendermint block store.

Modules should define a generic `sdk.Msg` for interface types at the module
level that uses `Any`. Ex:

```go
// x/gov/types/types.proto
message MsgSubmitAnyProposal {
  MsgSubmitProposalBase base = 1;
  google.protobuf.Any content = 2;
}
```

Apps should define an app-specific `sdk.Msg` that encodes the concrete types
it supports using an `oneof`:

```go
// myapp/types/types.proto
message MsgSubmitProposal {
  MsgSubmitProposalBase base = 1;
  Content content = 1;
}

message Content {
  oneof sum {
    TextProposal text = 1;
    SomeOtherProposal = 2;
  }
}
```

**Client libraries should always sign transactions using the generic module-level
`sdk.Msg` that uses `Any`.** Convenience gRPC/REST methods are described below
that allow module-level `Msg` to be used for transaction
submission, converting it to the app-level encoding `Msg` using the `oneof`.

In order to smoothly allow for conversion between _signing_ and _encoding_
`Msg`s, modules that use interface types should implement the following
interface:

```go
type InterfaceMsgEncoder {
  GetSigningMsg(sdk.Msg) (sdk.Msg, error)
  GetEncodingMsg(sdk.Msg) (sdk.Msg, error)
}
```

`GetSigningMsg` should type switch over the `sdk.Msg`s that use interfaces and
convert `Msg`s in the encoding format (that use `oneof`)
to the signing version (that uses `Any`). `GetEncodingMsg` should
convert `Msg`s used for signing (using `Any`) to those used for encoding
(using `oneof`). Ex:

```go
// x/gov/module.go
var _ InterfaceMsgEncoder = AppModule{}

type AppModule struct {
  ...
  newEncodingMsg fn(MsgSubmitProposalBase, Content) (MsgSubmitProposalI, error)
}

func (am AppModule) GetSigningMsg(msg sdk.Msg) (sdk.Msg, error) {
	switch msg := msg.(type) {
	case MsgSubmitProposalI:
        return NewMsgSubmitAnyProposal(msg.GetBase(), msg.GetContent()), nil
	default:
		return msg, nil
	}
}

func (am AppModule) GetEncodingMsg(sdk.Msg) (sdk.Msg, error) {
	switch msg := msg.(type) {
	case MsgSubmitProposalI:
        return am.newEncodingMsg(msg.GetBase(), msg.GetContent())
	default:
		return msg, nil
	}
}
```

Because apps will know how to convert signing `Msg`s to encoding `Msg`s,
signing `Msg`s can and should be used for transaction submission against a
helper function (for CLI methods) or RPC method that does the actually encoding.

### Generic gRPC/REST Transaction Service

The usage of `Any` by clients as described above allows us to provide a generic
app-independent transaction service (leveraging the generic `AnyTransaction` type
described above) which provides similar functionalities as the existing REST
server. This service will have a gRPC schema roughly as follows with REST
endpoints provided by grpc-gateway:

```go
service TxService {
    // Get a Tx by hash
    rpc QueryTx (QueryTxRequest) returns (QueryTxResponse) {
      option (google.api.http) = {
        get: "/txs/{hash}"
      };
    }
    
    // Search transactions
    rpc QueryTxs (QueryTxsRequest) returns (QueryTxResponse) {
      option (google.api.http) = {
        get: "/txs"
        body: "*"
      };
    }
 
    // Broadcast a signed tx
    rpc BroadcastTx (BroadcastTxRequest) returns (BroadcastTxResponse) {
      option (google.api.http) = {
        post: "/txs"
        body: "*"
      };
    }
}

message BroadcastTxRequest {
    AnyTransaction tx = 1;
    BroadcastTxMode mode = 2;
}

message QueryTxResponse {
    repeated AnyTransaction txs = 1;
    ...
}

...
```

For wallets and block explorers that want to easily target multiple chains, this
generic transaction service will provide a developer UX comparable to what is
currently available with amino except with protobuf.

Apps can also provide an app-level gRPC/REST service that queries and broadcasts
transactions using the app-specific `Transaction` type used for encoding. This
can be used by client apps that are very clearly targetting a single chain.

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

### Neutral

## References
