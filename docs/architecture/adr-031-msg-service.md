# ADR 031: Protobuf Msg Services

## Changelog

- 2020-10-05: Initial Draft

## Status

Proposed

## Context

In early conversations [it was also proposed](https://docs.google.com/document/d/1eEgYgvgZqLE45vETjhwIw4VOqK-5hwQtZtjVbiXnIGc/edit)
that Msg` return types be captured using a protobuf extension field, ex:

```protobuf
message MsgSubmitProposal
	option (cosmos_proto.msg_return) = “uint64”;
	bytes delegator_address = 1;
	bytes validator_address = 2;
	repeated sdk.Coin amount = 3;
}
```

This was never adopted, however, and we currently don’t have a mechanism for specifying \
the return type of a `Msg`.

Having a well-specified return value for `Msg`s would improve client UX. For instance,
in `x/gov`,  `MsgSubmitProposal` returns the proposal ID as a big-endian `uint64`.
This isn’t really documented anywhere and clients would need to know the internals
of the SDK to parse that value and return it to users.

Also, there may be cases where we want to use these return values programatically.
For instance, https://github.com/cosmos/cosmos-sdk/issues/7093 proposes a method for
doing inter-module Ocaps using the `Msg` router. A well-defined return type would
improve the developer UX for this approach.

Currently, we are encoding `Msg`s as `Any` in `Tx`s which involves packing the
binary-encoded `Msg` with its type URL.

The type URL for `MsgSubmitProposal` is `/cosmos.gov.MsgSubmitProposal`. 

The fully-qualified RPC name for the `SubmitProposal` RPC call above is
`/cosmos.gov.Msg/SubmitProposal` which is varies by a single `/` character.
We could also pack this type URL into an `Any` with the same `MsgSubmitProposal`
contents and handle it like a “packed RPC” call.


## Decision

We decide to add support for using protobuf `service` definitions for `Msg`s.

A `Msg` `service` definition would be written as follows:

```proto
package cosmos.gov;

service Msg {
  rpc SubmitProposal(MsgSubmitProposal) returns (MsgSubmitProposalResponse);
}

// Note that for backwards compatibility this uses MsgSubmitProposal as the request
// type instead of the more canonical MsgSubmitProposalRequest
message MsgSubmitProposal {
  google.protobuf.Any content = 1;
  bytes proposer = 2;
}

message MsgSubmitProposalResponse {
  uint64 proposal_id;
}
```

While this is most commonly used for gRPC, overloading protobuf `service` definitions like this does not violate
the intent of the [protobuf spec](https://developers.google.com/protocol-buffers/docs/proto3#services) which says:
> If you don’t want to use gRPC, it’s also possible to use protocol buffers with your own RPC implementation.
With this approach, we would get an auto-generated `MsgServer` interface:

In addition to clearly specifying return types, this has the benefit of generating client and server code. On the server
side, this is almost like an automatically generated keeper method and could maybe be used intead of keepers eventually
(see [\#7093](https://github.com/cosmos/cosmos-sdk/issues/7093)):

```go
type MsgServer interface {
  SubmitProposal(context.Context, *MsgSubmitProposal) (*MsgSubmitProposalResponse, error)
}
```

On the client side, developers could take advantage of this by creating RPC implementations that encapsulate transaction
logic. Protobuf libraries that use asynchronous callbacks, like [protobuf.js](https://github.com/protobufjs/protobuf.js#using-services)
could use this to register callbacks for specific messages even for transactions that span include multiple `Msg`s.

For backwards compatibility, existing `Msg` types should be used as the request parameter
for `service` definitions. Newer `Msg` types which only support `service` definitions
should use the more canonical `Msg...Request` names.

### Routing

In the future, `service` definitions may become the primary method for defining
`Msg`s. As a starting point, we need to integrate with the SDK's existing routing
and `Msg` interface. 

To do this we define a `ServiceMsg` struct which wraps any message defined with
`service` definition which contains the fully-qualified method name (from the
`Any.type_url` field) and the request body (from the `Any.value` field):

```go
type ServiceMsg struct {
  // MethodName is the fully-qualified service name
  MethodName string
  // Request is the request payload
  Request MsgRequest
}

type _ sdk.Msg = ServiceMsg{}
```

This `ServiceMsg` implements the `sdk.Msg` interface and its handler does the
actual method routing, allowing this feature to be added incrementally on top of
existing functionality.

### Deserialization

The `TxBody.UnpackageInterfaces` will need a special case
to detect if `Any` type URLs match the service method format (ex. `/cosmos.gov.Msg/SubmitProposal`)
by checking for two `/` characters and those messages will be decoded into `ServiceMsg`:

### `ServiceMsg` interface

All request messages will need to implement the `MsgRequest` interface which the
`ServiceMsg` wrapper will use to implement its own `ValidateBasic` and `GetSigners`
methods:
```go
type MsgRequest interface {
  proto.Message
  ValidateBasic() error
  GetSigners() []AccAddress
}
```

### Module Configuration

In [ADR 021](./adr-021-protobuf-query-encoding.md), we introduced a method `RegisterQueryServer`
to `AppModule` which allows for modules to register gRPC queriers.

For registering `Msg` services, we attempt an approach which is intended to be
more extensible by converting `RegisterQueryServer` to `RegisterServices(Configurator)`:

```go
type Configurator interface {
  QueryServer() grpc.Server
  MsgServer() grpc.Server
}

// example module:
func (am AppModule) RegisterServices(cfg Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper)
	types.RegisterMsgServer(cfg.MsgServer(), keeper)
}
```

This `RegisterServices` method and the `Configurator` method are intended to
evolve to satisfy the use cases discussed in [\#7093](https://github.com/cosmos/cosmos-sdk/issues/7093)
and [\#7122](https://github.com/cosmos/cosmos-sdk/issues/7421).

### `Msg` Service Implementation

Just like query services, `Msg` service methods can retrieve the `sdk.Context`
from the `context.Context` parameter method using the `sdk.UnwrapSDKContext`
method:

```go
func (k Keeper) SubmitProposal(goCtx context.Context, params *types.MsgSubmitProposal) (*MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
    ...
}
```

The `sdk.Context` should have an `EventManager` already attached by the `ServiceMsg`
router.

Separate handler definition is no longer needed with this approach.

## Consequences

### Pros
- communicates return type clearly
- manual handler registration and return type marshaling is no longer needed, just implement the interface and register it
- some keeper code could be automatically generate, would improve the UX of [\#7093](https://github.com/cosmos/cosmos-sdk/issues/7093) approach (1) if we chose to adopt that
- generated client code could be useful for clients

### Cons
- supporting both this and the current concrete `Msg` type approach could be confusing
- using `service` definitions outside the context of gRPC could be confusing (but doesn’t violate the proto3 spec)

## References

- [Initial Github Issue \#7122](https://github.com/cosmos/cosmos-sdk/issues/7122)
- [proto 3 Language Guide: Defining Services](https://developers.google.com/protocol-buffers/docs/proto3#services)
- [Initial pre-`Any` `Msg` designs](https://docs.google.com/document/d/1eEgYgvgZqLE45vETjhwIw4VOqK-5hwQtZtjVbiXnIGc)
- [ADR 020](./adr-020-protobuf-transaction-encoding.md)
- [ADR 021](./adr-021-protobuf-query-encoding.md)
