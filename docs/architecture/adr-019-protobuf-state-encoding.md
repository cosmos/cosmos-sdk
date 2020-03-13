# ADR 019: Protocol Buffer State Encoding

## Changelog

- 2020 Feb 15: Initial Draft
- 2020 Feb 24: Updates to handle messages with interface fields

## Status

Accepted

## Context

Currently, the Cosmos SDK utilizes [go-amino](https://github.com/tendermint/go-amino/) for binary
and JSON object encoding over the wire bringing parity between logical objects and persistence objects.

From the Amino docs:

> Amino is an object encoding specification. It is a subset of Proto3 with an extension for interface
> support. See the [Proto3 spec](https://developers.google.com/protocol-buffers/docs/proto3) for more
> information on Proto3, which Amino is largely compatible with (but not with Proto2).
>
> The goal of the Amino encoding protocol is to bring parity into logic objects and persistence objects.

Amino also aims to have the following goals (not a complete list):

- Binary bytes must be decode-able with a schema.
- Schema must be upgradeable.
- The encoder and decoder logic must be reasonably simple.

However, we believe that Amino does not fulfill these goals completely and does not fully meet the
needs of a truly flexible cross-language and multi-client compatible encoding protocol in the Cosmos SDK.
Namely, Amino has proven to be a big pain-point in regards to supporting object serialization across
clients written in various languages while providing virtually little in the way of true backwards
compatibility and upgradeability. Furthermore, through profiling and various benchmarks, Amino has
been shown to be an extremely large performance bottleneck in the Cosmos SDK <sup>1</sup>. This is
largely reflected in the performance of simulations and application transaction throughput.

Thus, we need to adopt an encoding protocol that meets the following criteria for state serialization:

- Language agnostic
- Platform agnostic
- Rich client support and thriving ecosystem
- High performance
- Minimal encoded message size
- Codegen-based over reflection-based
- Supports backward and forward compatibility

Note, migrating away from Amino should be viewed as a two-pronged approach, state and client encoding.
This ADR focuses on state serialization in the Cosmos SDK state machine. A corresponding ADR will be
made to address client-side encoding.

## Decision

We will adopt [Protocol Buffers](https://developers.google.com/protocol-buffers) for serializing
persisted structured data in the Cosmos SDK while providing a clean mechanism and developer UX for
applications wishing to continue to use Amino. We will provide this mechanism by updating modules to
accept a codec interface, `Marshaler`, instead of a concrete Amino codec. Furthermore, the Cosmos SDK
will provide three concrete implementations of the `Marshaler` interface: `AminoCodec`, `ProtoCodec`,
and `HybridCodec`.

- `AminoCodec`: Uses Amino for both binary and JSON encoding.
- `ProtoCodec`: Uses Protobuf for or both binary and JSON encoding.
- `HybridCodec`: Uses Amino for JSON encoding and Protobuf for binary encoding.

Until the client migration landscape is fully understood and designed, modules will use a `HybridCodec`
as the concrete codec it accepts and/or extends. This means that all client JSON encoding, including
genesis state, will still use Amino. The ultimate goal will be to replace Amino JSON encoding with
Protbuf encoding and thus have modules accept and/or extend `ProtoCodec`.

### Module Design

Modules that do not require the ability to work with and serialize interfaces, the path to Protobuf
migration is pretty straightforward. These modules are to simply migrate any existing types that
are encoded and persisted via their concrete Amino codec to Protobuf and have their keeper accept a
`Marshaler` that will be a `HybridCodec`. This migration is simple as things will just work as-is.

Note, any business logic that needs to encode primitive types like `bool` or `int64` should use
[gogoprotobuf](https://github.com/gogo/protobuf) Value types.

Example:

```go
  ts, err := gogotypes.TimestampProto(completionTime)
  if err != nil {
    // ...
  }

  bz := cdc.MustMarshalBinaryBare(ts)
```

However, modules can vary greatly in purpose and design and so we must support the ability for modules
to be able to encode and work with interfaces (e.g. `Account` or `Content`). For these modules, they
must define their own codec interface that extends `Marshaler`. These specific interfaces are unique
to the module and will contain method contracts that know how to serialize the needed interfaces.

Example:

```go
// x/auth/types/codec.go

type Codec interface {
  codec.Marshaler

  MarshalAccount(acc exported.Account) ([]byte, error)
  UnmarshalAccount(bz []byte) (exported.Account, error)

  MarshalAccountJSON(acc exported.Account) ([]byte, error)
  UnmarshalAccountJSON(bz []byte) (exported.Account, error)
}
```

Note, concrete types implementing these interfaces can be defined outside the scope of the module
that defines the interface (e.g. `ModuleAccount` in `x/supply`). To handle these cases, a Protobuf
message must be defined at the application-level along with a single codec that will be passed to _all_
modules using a `oneof` approach.

Example:

```protobuf
// app/codec/codec.proto

import "third_party/proto/cosmos-proto/cosmos.proto";
import "x/auth/types/types.proto";
import "x/auth/vesting/types/types.proto";
import "x/supply/types/types.proto";

message Account {
  option (cosmos_proto.interface_type) = "*github.com/cosmos/cosmos-sdk/x/auth/exported.Account";

  // sum defines a list of all acceptable concrete Account implementations.
  oneof sum {
    cosmos_sdk.x.auth.v1.BaseAccount                      base_account               = 1;
    cosmos_sdk.x.auth.vesting.v1.ContinuousVestingAccount continuous_vesting_account = 2;
    cosmos_sdk.x.auth.vesting.v1.DelayedVestingAccount    delayed_vesting_account    = 3;
    cosmos_sdk.x.auth.vesting.v1.PeriodicVestingAccount   periodic_vesting_account   = 4;
    cosmos_sdk.x.supply.v1.ModuleAccount                  module_account             = 5;
  }

  // ...
}
```

```go
// app/codec/codec.go

type Codec struct {
  codec.Marshaler


  amino *codec.Codec
}

func NewAppCodec(amino *codec.Codec) *Codec {
  return &Codec{Marshaler: codec.NewHybridCodec(amino), amino: amino}
}

func (c *Codec) MarshalAccount(accI authexported.Account) ([]byte, error) {
  acc := &Account{}
  if err := acc.SetAccount(accI); err != nil {
    return nil, err
  }

  return c.Marshaler.MarshalBinaryBare(acc)
}

func (c *Codec) UnmarshalAccount(bz []byte) (authexported.Account, error) {
  acc := &Account{}
  if err := c.Marshaler.UnmarshalBinaryBare(bz, acc); err != nil {
    return nil, err
  }

  return acc.GetAccount(), nil
}
```

Since the `Codec` implements `auth.Codec` (and all other required interfaces), it is passed to _all_
the modules and satisfies all the interfaces. Now each module needing to work with interfaces will know
about all the required types. Note, the use of `interface_type` allows us to avoid a significant
amount of code boilerplate when implementing the `Codec`.

A similar concept is to be applied for messages that contain interfaces fields. The module will
define a "base" concrete message type that the application-level codec will extend via `oneof` that
fulfills the required message interface.

Example:

The `MsgSubmitEvidence` defined by the `x/evidence` module contains a field `Evidence` which is an
interface.

```go
type MsgSubmitEvidence struct {
  Evidence  exported.Evidence
  Submitter sdk.AccAddress
}
```

Instead, we will implement a "base" message type and an interface which the concrete message type
must implement.

```protobuf
// x/evidence/types/types.proto

message MsgSubmitEvidenceBase {
  bytes submitter = 1
    [
      (gogoproto.casttype) = "github.com/cosmos/cosmos-sdk/types.AccAddress"
    ];
}
```

```go
// x/evidence/exported/evidence.go

type MsgSubmitEvidence interface {
  sdk.Msg

  GetEvidence() Evidence
  GetSubmitter() sdk.AccAddress
}
```

Notice the `MsgSubmitEvidence` interface extends `sdk.Msg` and allows for the `Evidence` interface
to be retrieved from the concrete message type.

Now, the application-level codec will define the concrete `MsgSubmitEvidence` type and will have it
fulfill the `MsgSubmitEvidence` interface defined by `x/evidence`.

```protobuf
// app/codec/codec.proto

message Evidence {
  option (gogoproto.equal)             = true;
  option (cosmos_proto.interface_type) = "github.com/cosmos/cosmos-sdk/x/evidence/exported.Evidence";

  oneof sum {
    cosmos_sdk.x.evidence.v1.Equivocation equivocation = 1;
  }
}

message MsgSubmitEvidence {
  option (gogoproto.equal)           = true;
  option (gogoproto.goproto_getters) = false;

  Evidence                                       evidence = 1;
  cosmos_sdk.x.evidence.v1.MsgSubmitEvidenceBase base     = 2
    [
      (gogoproto.nullable) = false,
      (gogoproto.embed)    = true
    ];
}
```

```go
// app/codec/msgs.go

func (msg MsgSubmitEvidence) GetEvidence() eviexported.Evidence {
  return msg.Evidence.GetEvidence()
}

func (msg MsgSubmitEvidence) GetSubmitter() sdk.AccAddress {
  return msg.Submitter
}
```

Note, however, the module's message handler must now handle the interface `MsgSubmitEvidence` in
addition to any concrete types.

### Why Wasn't X Chosen Instead

For a more complete comparison to alternative protocols, see [here](https://codeburst.io/json-vs-protocol-buffers-vs-flatbuffers-a4247f8bda6f).

### Cap'n Proto

While [Cap’n Proto](https://capnproto.org/) does seem like an advantageous alternative to Protobuf
due to it's native support for interfaces/generics and built in canonicalization, it does lack the
rich client ecosystem compared to Protobuf and is a bit less mature.

### FlatBuffers

[FlatBuffers](https://google.github.io/flatbuffers/) is also a potentially viable alternative, with the
primary difference being that FlatBuffers does not need a parsing/unpacking step to a secondary
representation before you can access data, often coupled with per-object memory allocation.

However, it would require great efforts into research and full understanding the scope of the migration
and path forward -- which isn't immediately clear. In addition, FlatBuffers aren't designed for
untrusted inputs.

## Future Improvements & Roadmap

The landscape and roadmap to restructuring queriers and tx generation to fully support
Protobuf isn't fully understood yet. Once all modules are migrated, we will have a better
understanding on how to proceed with client improvements (e.g. gRPC) <sup>2</sup>.

## Consequences

### Positive

- Significant performance gains.
- Supports backward and forward type compatibility.
- Better support for cross-language clients.

### Negative

- Learning curve required to understand and implement Protobuf messages.
- Less flexibility in cross-module type registration. We now need to define types
at the application-level.
- Client business logic and tx generation may become a bit more complex.

### Neutral

## References

1. https://github.com/cosmos/cosmos-sdk/issues/4977
2. https://github.com/cosmos/cosmos-sdk/issues/5444
