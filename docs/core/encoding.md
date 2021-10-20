<!--
order: 6
-->

# Encoding

While encoding in the Cosmos SDK used to be mainly handled by `go-amino` codec, the Cosmos SDK is moving towards using `gogoprotobuf` for both state and client-side encoding. {synopsis}

## Pre-requisite Readings

- [Anatomy of a Cosmos SDK application](../basics/app-anatomy.md) {prereq}

## Encoding

The Cosmos SDK utilizes two binary wire encoding protocols, [Amino](https://github.com/tendermint/go-amino/) which is an object encoding specification and [Protocol Buffers](https://developers.google.com/protocol-buffers), a subset of Proto3 with an extension for
interface support. See the [Proto3 spec](https://developers.google.com/protocol-buffers/docs/proto3)
for more information on Proto3, which Amino is largely compatible with (but not with Proto2).

Due to Amino having significant performance drawbacks, being reflection-based, and
not having any meaningful cross-language/client support, Protocol Buffers, specifically
[gogoprotobuf](https://github.com/gogo/protobuf/), is being used in place of Amino.
Note, this process of using Protocol Buffers over Amino is still an ongoing process.

Binary wire encoding of types in the Cosmos SDK can be broken down into two main
categories, client encoding and store encoding. Client encoding mainly revolves
around transaction processing and signing, whereas store encoding revolves around
types used in state-machine transitions and what is ultimately stored in the Merkle
tree.

For store encoding, protobuf definitions can exist for any type and will typically
have an Amino-based "intermediary" type. Specifically, the protobuf-based type
definition is used for serialization and persistence, whereas the Amino-based type
is used for business logic in the state-machine where they may converted back-n-forth.
Note, the Amino-based types may slowly be phased-out in the future so developers
should take note to use the protobuf message definitions where possible.

In the `codec` package, there exists two core interfaces, `Marshaler` and `ProtoMarshaler`,
where the former encapsulates the current Amino interface except it operates on
types implementing the latter instead of generic `interface{}` types.

In addition, there exists two implementations of `Marshaler`. The first being
`AminoCodec`, where both binary and JSON serialization is handled via Amino. The
second being `ProtoCodec`, where both binary and JSON serialization is handled
via Protobuf.

This means that modules may use Amino or Protobuf encoding but the types must
implement `ProtoMarshaler`. If modules wish to avoid implementing this interface
for their types, they may use an Amino codec directly.

### Amino

Every module uses an Amino codec to serialize types and interfaces. This codec typically
has types and interfaces registered in that module's domain only (e.g. messages),
but there are exceptions like `x/gov`. Each module exposes a `RegisterLegacyAminoCodec` function
that allows a user to provide a codec and have all the types registered. An application
will call this method for each necessary module.

Where there is no protobuf-based type definition for a module (see below), Amino
is used to encode and decode raw wire bytes to the concrete type or interface:

```go
bz := keeper.cdc.MustMarshal(typeOrInterface)
keeper.cdc.MustUnmarshal(bz, &typeOrInterface)
```

Note, there are length-prefixed variants of the above functionality and this is
typically used for when the data needs to be streamed or grouped together
(e.g. `ResponseDeliverTx.Data`)

### Gogoproto

Modules are encouraged to utilize Protobuf encoding for their respective types. In the Cosmos SDK, we use the [Gogoproto](https://github.com/gogo/protobuf) specific implementation of the Protobuf spec that offers speed and DX improvements compared to the official [Google protobuf implementation](https://github.com/protocolbuffers/protobuf).

### Guidelines for protobuf message definitions

In addition to [following official Protocol Buffer guidelines](https://developers.google.com/protocol-buffers/docs/proto3#simple), we recommend using these annotations in .proto files when dealing with interfaces:

- use `cosmos_proto.accepts_interface` to annote fields that accept interfaces
- pass the same fully qualified name as `protoName` to `InterfaceRegistry.RegisterInterface`
- annotate interface implementations with `cosmos_proto.implements_interface`
- pass the same fully qualified name as `protoName` to `InterfaceRegistry.RegisterInterface`

### Transaction Encoding

Another important use of Protobuf is the encoding and decoding of
[transactions](./transactions.md). Transactions are defined by the application or
the Cosmos SDK but are then passed to the underlying consensus engine to be relayed to
other peers. Since the underlying consensus engine is agnostic to the application,
the consensus engine accepts only transactions in the form of raw bytes.

- The `TxEncoder` object performs the encoding.
- The `TxDecoder` object performs the decoding.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/types/tx_msg.go#L83-L87

A standard implementation of both these objects can be found in the [`auth` module](../../x/auth/spec/README.md):

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/x/auth/tx/decoder.go

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/x/auth/tx/encoder.go

See [ADR-020](../architecture/adr-020-protobuf-transaction-encoding.md) for details of how a transaction is encoded.

### Interface Encoding and Usage of `Any`

The Protobuf DSL is strongly typed, which can make inserting variable-typed fields difficult. Imagine we want to create a `Profile` protobuf message that serves as a wrapper over [an account](../basics/accounts.md):

```proto
message Profile {
  // account is the account associated to a profile.
  cosmos.auth.v1beta1.BaseAccount account = 1;
  // bio is a short description of the account.
  string bio = 4;
}
```

In this `Profile` example, we hardcoded `account` as a `BaseAccount`. However, there are several other types of [user accounts related to vesting](../../x/auth/spec/05_vesting.md), such as `BaseVestingAccount` or `ContinuousVestingAccount`. All of these accounts are different, but they all implement the `AccountI` interface. How would you create a `Profile` that allows all these types of accounts with an `account` field that accepts an `AccountI` interface?

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/auth/types/account.go#L307-L330

In [ADR-019](../architecture/adr-019-protobuf-state-encoding.md), it has been decided to use [`Any`](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/any.proto)s to encode interfaces in protobuf. An `Any` contains an arbitrary serialized message as bytes, along with a URL that acts as a globally unique identifier for and resolves to that message's type. This strategy allows us to pack arbitrary Go types inside protobuf messages. Our new `Profile` then looks like:

```protobuf
message Profile {
  // account is the account associated to a profile.
  google.protobuf.Any account = 1 [
    (cosmos_proto.accepts_interface) = "AccountI"; // Asserts that this field only accepts Go types implementing `AccountI`. It is purely informational for now.
  ];
  // bio is a short description of the account.
  string bio = 4;
}
```

To add an account inside a profile, we need to "pack" it inside an `Any` first, using `codectypes.NewAnyWithValue`:

```go
var myAccount AccountI
myAccount = ... // Can be a BaseAccount, a ContinuousVestingAccount or any struct implementing `AccountI`

// Pack the account into an Any
accAny, err := codectypes.NewAnyWithValue(myAccount)
if err != nil {
  return nil, err
}

// Create a new Profile with the any.
profile := Profile {
  Account: accAny,
  Bio: "some bio",
}

// We can then marshal the profile as usual.
bz, err := cdc.Marshal(profile)
jsonBz, err := cdc.MarshalJSON(profile)
```

To summarize, to encode an interface, you must 1/ pack the interface into an `Any` and 2/ marshal the `Any`. For convenience, the Cosmos SDK provides a `MarshalInterface` method to bundle these two steps. Have a look at [a real-life example in the x/auth module](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/auth/keeper/keeper.go#L218-L221).

The reverse operation of retrieving the concrete Go type from inside an `Any`, called "unpacking", is done with the `GetCachedValue()` on `Any`.

```go
profileBz := ... // The proto-encoded bytes of a Profile, e.g. retrieved through gRPC.
var myProfile Profile
// Unmarshal the bytes into the myProfile struct.
err := cdc.Unmarshal(profilebz, &myProfile)

// Let's see the types of the Account field.
fmt.Printf("%T\n", myProfile.Account)                  // Prints "Any"
fmt.Printf("%T\n", myProfile.Account.GetCachedValue()) // Prints "BaseAccount", "ContinuousVestingAccount" or whatever was initially packed in the Any.

// Get the address of the accountt.
accAddr := myProfile.Account.GetCachedValue().(AccountI).GetAddress()
```

It is important to note that for `GetCachedValue()` to work, `Profile` (and any other structs embedding `Profile`) must implement the `UnpackInterfaces` method:

```go
func (p *Profile) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
  if p.Account != nil {
    var account AccountI
    return unpacker.UnpackAny(p.Account, &account)
  }

  return nil
}
```

The `UnpackInterfaces` gets called recursively on all structs implementing this method, to allow all `Any`s to have their `GetCachedValue()` correctly populated.

For more information about interface encoding, and especially on `UnpackInterfaces` and how the `Any`'s `type_url` gets resolved using the `InterfaceRegistry`, please refer to [ADR-019](../architecture/adr-019-protobuf-state-encoding.md).

#### `Any` Encoding in the Cosmos SDK

The above `Profile` example is a fictive example used for educational purposes. In the Cosmos SDK, we use `Any` encoding in several places (non-exhaustive list):

- the `cryptotypes.PubKey` interface for encoding different types of public keys,
- the `sdk.Msg` interface for encoding different `Msg`s in a transaction,
- the `AccountI` interface for encodinig different types of accounts (similar to the above example) in the x/auth query responses,
- the `Evidencei` interface for encoding different types of evidences in the x/evidence module,
- the `AuthorizationI` interface for encoding different types of x/authz authorizations,
- the [`Validator`](https://github.com/cosmos/cosmos-sdk/blob/v0.42.5/x/staking/types/staking.pb.go#L306-L337) struct that contains information about a validator.

A real-life example of encoding the pubkey as `Any` inside the Validator struct in x/staking is shown in the following example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/staking/types/validator.go#L40-L61

## FAQ

1. How to create modules using protobuf encoding?

**Defining module types**

Protobuf types can be defined to encode:

- state
- [`Msg`s](../building-modules/messages-and-queries.md#messages)
- [Query services](../building-modules/query-services.md)
- [genesis](../building-modules/genesis.md)

**Naming and conventions**

We encourage developers to follow industry guidelines: [Protocol Buffers style guide](https://developers.google.com/protocol-buffers/docs/style)
and [Buf](https://buf.build/docs/style-guide), see more details in [ADR 023](../architecture/adr-023-protobuf-naming.md)

2. How to update modules to protobuf encoding?

If modules do not contain any interfaces (e.g. `Account` or `Content`), then they
may simply migrate any existing types that
are encoded and persisted via their concrete Amino codec to Protobuf (see 1. for further guidelines) and accept a `Marshaler` as the codec which is implemented via the `ProtoCodec`
without any further customization.

However, if a module type composes an interface, it must wrap it in the `skd.Any` (from `/types` package) type. To do that, a module-level .proto file must use [`google.protobuf.Any`](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/any.proto) for respective message type interface types.

For example, in the `x/evidence` module defines an `Evidence` interface, which is used by the `MsgSubmitEvidence`. The structure definition must use `sdk.Any` to wrap the evidence file. In the proto file we define it as follows:

```protobuf
// proto/cosmos/evidence/v1beta1/tx.proto

message MsgSubmitEvidence {
  string              submitter = 1;
  google.protobuf.Any evidence  = 2 [(cosmos_proto.accepts_interface) = "Evidence"];
}
```

The Cosmos SDK `codec.Codec` interface provides support methods `MarshalInterface` and `UnmarshalInterface` to easy encoding of state to `Any`.

Module should register interfaces using `InterfaceRegistry` which provides a mechanism for registering interfaces: `RegisterInterface(protoName string, iface interface{})` and implementations: `RegisterImplementations(iface interface{}, impls ...proto.Message)` that can be safely unpacked from Any, similarly to type registration with Amino:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/codec/types/interface_registry.go#L25-L66

In addition, an `UnpackInterfaces` phase should be introduced to deserialization to unpack interfaces before they're needed. Protobuf types that contain a protobuf `Any` either directly or via one of their members should implement the `UnpackInterfacesMessage` interface:

```go
type UnpackInterfacesMessage interface {
  UnpackInterfaces(InterfaceUnpacker) error
}
```

## Next {hide}

Learn about [gRPC, REST and other endpoints](./grpc_rest.md) {hide}
