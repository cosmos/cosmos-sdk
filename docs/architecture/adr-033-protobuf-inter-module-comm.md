# ADR 033: Protobuf-based Inter-Module Communication

## Changelog

- 2020-10-05: Initial Draft

## Status

Proposed

## Abstract

This ADR introduces a system for permissioned inter-module communication leveraging the protobuf `Query` and `Msg`
service definitions defined in [ADR 021](./adr-021-protobuf-query-encoding.md) and
[ADR 031](./adr-031-msg-service.md) which provides:
- stable module interfaces to eventually replace the keeper paradigm based on protobuf
- stronger inter-module object capabilities guarantees
- module accounts and sub-account authorization

## Context

In the current Cosmos SDK documentation on the [Object-Capability Model](../docs/core/ocap.md), it is state that:

> We assume that a thriving ecosystem of Cosmos-SDK modules that are easy to compose into a blockchain application will contain faulty or malicious modules.

There is currently not a thriving ecosystem of Cosmos SDK modules. We hypothesize that this is in part due to:
1. lack of a stable v1.0 Cosmos SDK to build modules off of. Module interfaces are changing, sometimes dramatically, from
point release to point release, often for good reasons, but this does not create a stable foundation to build on.
2. lack of a properly implemented object capability or even object-oriented encapsulation system which makes refactors
of module keeper interfaces inevitable because the current interfaces are poorly constrained.

### `x/bank` Case Study

We use `x/bank` of this.

Currently the `x/bank` keeper gives pretty much unrestricted access to any module which references it. For instance, the
`SetBalance` method allows the caller to set the balance of any account to anything, bypassing even proper tracking of supply.

There appears to have been some later attempts to implement some semblance of Ocaps using module-level minting, staking
and burning permissions. These permissions allow a module to mint, burn or delegate tokens with reference to the module’s
own account. These permissions are actually stored as a `[]string` array on the `ModuleAccount` type in state.

However, these permissions don’t really do much. They control what modules can be referenced in the `MintCoins`,
`BurnCoins` and `DelegateCoins***` methods, but for one there is no unique object capability token that controls access
- just a simple string. So the `x/upgrade` module could mint tokens for the `x/staking` module simple by calling
`MintCoins(“staking”)`. Furthermore, all modules which have access to these keeper methods, also have access to
`SetBalance` negating any other attempt at Ocaps and breaking even basic object-oriented encapsulation.

## Decision

Starting from the work in [ADR 31](./adr-031-msg-service.md), we introduce the following inter-module communication system
to replace the existing keeper paradigm. These two pieces together are intended to form the basis of a Cosmos SDK v1.0
that provides the necessary stability and encapsulation guarantees that allow a thriving module ecosystem to emerge.

### New "Keeper" Paradigm

In [ADR 021](./adr-021-protobuf-query-encoding.md), a mechanism for using protobuf service definitions to define queriers
was introduced and in [ADR 31](./adr-031-msg-service.md), a mechanism for using protobuf service to define `Msg`s was added.
Protobuf service definitions generate two golang interfaces representing the client and server sides of a service plus
some helper code. Here is a minimal example for the bank `Send` `Msg`.

```go
package bank

type MsgClient interface {
	Send(context.Context, *MsgSend, opts ...grpc.CallOption) (*MsgSendResponse, error)
}

type MsgServer interface {
	Send(context.Context, *MsgSend) (*MsgSendResponse, error)
}
```

[ADR 021](./adr-021-protobuf-query-encoding.md) and [ADR 31]() specifies how modules can implement the generated `QueryServer`
and `MsgServer` interfaces as replacements for the legacy queriers and `Msg` handlers respectively.

In this ADR we explain how modules can make queries and send `Msg`s to other modules using the generated `QueryClient`
and `MsgClient` interfaces and propose this mechanism as a replacement for the existing `Keeper` paradigm.

Using this `QueryClient`/`MsgClient` approach has the following key benefits over keepers:
1. Protobuf types are checked for breaking changes using [buf](https://buf.build/docs/breaking-overview) and because of
the way protobuf is designed this will give us strong backwards compatibility guarantees while allowing for forward
evolution.
2. The separation between the client and server interfaces will allow us to insert permission checking code in between
the two which checks if one module is authorized to send the specified `Msg` to the other module providing a proper
object capability system.

This mechanism has the added benefits of:
- reducing boilerplate through code generation, and
- allowing for modules in other languages either via a VM like CosmWasm or sub-processes using gRPC 

### Inter-module Communication

In order for code to use the `Client` interfaces generated by the protobuf compiler, a `grpc.ClientConn` implementation
is needed. We introduce a new type, `ModuleKey`, to serve this role which we can conceptualize as the "private key"
corresponding to a module account.

Whereas external clients use their private key to sign transactions containing `Msg`s where they are listed as signers,
modules use their `ModuleKey` to send `Msg`s where they are listed as the sole signer to other modules. For example, modules
could use their `ModuleKey` to "sign" a `/cosmos.bank.Msg/Send` transaction to send coins from the module's account to another
account.

`QueryClient`s could also be made with `ModuleKey`s, except that authentication isn't required.

Here's an example of a hypothetical module `foo` interacting with `x/bank`:
```go
package foo

func (fooMsgServer *MsgServer) Bar(ctx context.Context, req *MsgBar) (*MsgBarResponse, error) {
  bankQueryClient := bank.NewQueryClient(fooMsgServer.moduleKey)
  balance, err := bankQueryClient.Balance(&bank.QueryBalanceRequest{Address: fooMsgServer.moduleKey.Address(), Denom: "foo"})
  
  ...

  bankMsgClient := bank.NewMsgClient(fooMsgServer.moduleKey)
  res, err := bankMsgClient.Balance(ctx, &bank.MsgSend{FromAddress: fooMsgServer.moduleKey.Address(), ...})

  ...
}
```

### `ModuleKey`s and `ModuleID`s

A `ModuleKey` can be thought of as a "private key" for a module account and a `ModuleID` can be thought of as the
corresponding "public key". From [the ADR 028 draft](https://github.com/cosmos/cosmos-sdk/pull/7086), modules can have both a root module account and any number of sub-accounts
or derived accounts that can be used for different pools (ex. staking pools) or managed accounts (ex. group
accounts). We can also think of module sub-accounts as similar to derived keys - there is a root key and then some
derivation path. `ModuleID` is a simple struct which contains the module name and optional "derivation" path,
and forms its address based on the `AddressHash` method from [the ADR 028 draft](https://github.com/cosmos/cosmos-sdk/pull/7086):

```go
type ModuleID struct {
  ModuleName string
  Path []byte
}

func (key ModuleID) Address() []byte {
  return AddressHash(key.ModuleName, key.Path)
}
```

In addition to being able to generate a `ModuleID` and address, a `ModuleKey` contains a special function closure called
the `Invoker` which is the key to safe inter-module access. This function closure corresponds to the `Invoke` method in
the `grpc.ClientConn` interface and under the hood is able to route messages to the appropriate `Msg` and `Query` handlers
performing appropriate security checks on `Msg`s. This allows for even safer inter-module access than keeper's whose
private member variables could be manipulated through reflection. Golang does not support reflection on a function
closure's captured variables and direct manipulation of memory would be needed for a truly malicious module to bypass
the `ModuleKey` security.

The two `ModuleKey` types are `RootModuleKey` and `DerivedModuleKey`:

```go
func Invoker(callInfo CallInfo) func(ctx context.Context, request, response interface{}, opts ...interface{}) error

type CallInfo {
  Method string
  Caller ModuleID
}

type RootModuleKey struct {
  moduleName string
  invoker Invoker
}

type DerivedModuleKey struct {
  moduleName string
  path []byte
  invoker Invoker
}
```

A module can get access to a `DerivedModuleKey`, using the `Derive(path []byte)` method on `RootModuleKey` and then
would use this key to authenticate `Msg`s from a sub-account. Ex:

```go
package foo

func (fooMsgServer *MsgServer) Bar(ctx context.Context, req *MsgBar) (*MsgBarResponse, error) {
  derivedKey := fooMsgServer.moduleKey.Derive(req.SomePath)
  bankMsgClient := bank.NewMsgClient(derivedKey)
  res, err := bankMsgClient.Balance(ctx, &bank.MsgSend{FromAddress: derivedKey.Address(), ...})
  ...
}
```

In this way, a module can gain permissioned access to a root account and any number of sub-accounts and send
authenticated `Msg`s from these accounts. The `Invoker` `callInfo.Caller` parameter is used under the hood to
distinguish between different module accounts, but either way the function returned by `Invoker` only allows `Msg`s
from either the root or a derived module account to pass through.

Note that `Invoker` itself returns a function closure based on the `CallInfo` passed in. This will allow client implementations
in the future that cache the invoke function for each method type avoiding the overhead of hash table lookup.
This would reduce the performance overhead of this inter-module communication method to the bare minimum required for
checking permissions.

Below is a rough sketch of the implementation of `grpc.ClientConn.Invoke` for `RootModuleKey`:

```go
func (key RootModuleKey) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
  f := key.invoker(CallInfo {Method: method, Caller: ModuleID {ModuleName: key.moduleName}})
  return f(ctx, args, reply)
}
```

### `AppModule` Wiring and Requirements

In [ADR 031](./adr-031-msg-service.md), the `AppModule.RegisterService(Configurator)` method was introduced. To support
inter-module communication, we extend the `Configurator` interface to pass in the `ModuleKey` and to allow modules to
specify their dependencies on other modules using `RequireServer()`:


```go
type Configurator interface {
   QueryServer() grpc.Server
   MsgServer() grpc.Server

   ModuleKey() ModuleKey
   RequireServer(serverInterface interface{})
}
```

The `ModuleKey` is passed to modules in the `RegisterService` method itself so that `RegisterServices` serves as a single
entry point for configuring module services. This is intended to also have the side-effect of reducing boilerplate in
`app.go`. For now, `ModuleKey`s will be created based on `AppModuleBasic.Name()`, but a more flexible system may be
introduced in the future. The `ModuleManager` will handle creation of module accounts behind the scenes.

Because modules do not get direct access to each other anymore, modules may have unfulfilled dependencies. To make sure
that module dependencies are resolved at startup, the `Configurator.RequireServer` method should be added. The `ModuleManager`
will make sure that all dependencies declared with `RequireServer` can be resolved before the app starts. An example
module `foo` could declare it's dependency on `x/bank` like this:

```go
package foo

func (am AppModule) RegisterServices(cfg Configurator) {
  cfg.RequireServer((*bank.QueryServer)(nil))
  cfg.RequireServer((*bank.MsgServer)(nil))
}
```

### Security Considerations

In addition to checking for `ModuleKey` permissions, a few additional security precautions will need to be taken by
the underlying router infrastructure.

#### Recursion and Re-entry

Recursive or re-entrant method invocations pose a potential security threat. This can be a problem if Module A
calls Module B and Module B calls module A again in the same call.

One basic way for the router system to deal with this is to maintain a call stack which prevents a module from
being referenced more than once in the call stack so that there is no re-entry. A `map[string]interface{}` table
in the router could be used to perform this security check.

#### Queries

Queries in Cosmos SDK are generally un-permissioned so allowing one module to query another module should not pose
any major security threats assuming basic precautions are taken. The basic precaution that the router system will
need to take is making sure that the `sdk.Context` passed to query methods does not allow writing to the store. This
can be done for now with a `CacheMultiStore` as is currently done for `BaseApp` queries.

### Future Work

Separate ADRs will address the use cases of:
* unrestricted, "admin" access
* dynamic interface routing (ex. `x/gov` `Content` routing)
* inter-module hooks (ex. `x/staking/keeper/hooks.go`)

Other future improvements may include:
* combining `StoreKey`s and `ModuleKey`s into a single interface so that modules have a single Ocaps handle
* code generation which makes inter-module communication more performant
* decoupling `ModuleKey` creation from `AppModuleBasic.Name()` so that app's can override root module account names

## Consequences

### Backwards Compatibility

This ADR is intended to provide a pathway to a scenario where there is greater long term compatibility between modules.
In the short-term, this will likely result in breaking certain `Keeper` interfaces which are too permissive and/or
replacing `Keeper` interfaces altogether.

### Positive

- proper inter-module Ocaps
- an alternative to keepers which can more easily lead to stable inter-module interfaces

### Negative

- modules which adopt this will need significant refactoring

### Neutral

## Test Cases [optional]

## References

- [ADR 021](./adr-021-protobuf-query-encoding.md)
- [ADR 031](./adr-031-msg-service.md)
- [ADR 028 draft](https://github.com/cosmos/cosmos-sdk/pull/7086)
- [Object-Capability Model](../docs/core/ocap.md)