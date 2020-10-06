# ADR 033: Protobuf Module Object Capabilities

## Changelog

- 2020-10-05: Initial Draft

## Status

Proposed

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.


## Context

In the current Cosmos SDK documentation on the [Object-Capability Model](../docs/core/ocap.md), it is state that:

> We assume that a thriving ecosystem of Cosmos-SDK modules that are easy to compose into a blockchain application will contain faulty or malicious modules.

There is currently not a thriving ecosystem of Cosmos SDK modules. We hypothesize that this is in part due to:
1. lack of a stable v1.0 Cosmos SDK to build modules off of. Module interfaces are changing, sometimes dramatically, from
point release to point release, often for good reasons, but this does not create a stable foundation to build on.
2. lack of a properly implemented object capability or even object-oriented encapsulation system.

### `x/bank` Case Study

We use `x/bank as a case study` of this.

Currently the `x/bank` keeper gives pretty much unrestricted access to any module which references it. For instance, the
`SetBalance` method allows the caller to set the balance of any account to anything, bypassing even proper tracking of supply.

There appears to have been some later attempts to implement some semblance of Ocaps using module-level minting, staking
and burning permissions. These permissions allow a module to mint, burn or delegate tokens with reference to the module’s
own account. These permissions are actually stored as a `[]string` array on the `ModuleAccount` type in state.

However, these permissions don’t really do much. They control what modules can be referenced in the `MintCoins`,
`BurnCoins` and `DelegateCoins***` methods, but for one there is no unique object capability token that controls access
- just a simple string. So the `x/upgrade` module could mint tokens for the `x/staking` module simple by calling
`MintCoins(“staking”)`. Furthermore, all modules which have access to these keeper methods, also have access to
`SetBalance` negating any level of ocaps or even basic object-oriented encapsulation.

## Decision

Starting from the work in [ADR 31](Protobuf Msg Services), we introduce the following inter-module communication system
to replace the existing keeper paradigm. These two pieces together are intended to form the basis of a Cosmos SDK v1.0
that provides the necessary stability and encapsulation guarantees that allow a thriving module ecosystem to emerge.

### New "Keeper" Paradigm

In [ADR 021](./adr-021-protobuf-query-encoding.md), a mechanism for using protobuf service definitions to define queriers
was introduced and in [ADR 31](), protobuf service definition representation of `Msg`s was added.
Protobuf service definitions generate two golang interfaces representing the client and server sides of a service plus
some helper code. Ex:

```go
package bank

type MsgClient interface {
	Send(context.Context, *MsgSend, opts ...grpc.CallOption) (*MsgSendResponse, error)
}

type MsgServer interface {
	Send(context.Context, *MsgSend) (*MsgSendResponse, error)
}
```

[ADR 31](Protobuf Msg Services) specifies how modules can implement the `MsgServer` interface as a replacement for the
legacy handler registration.

In this ADR we explain how modules can send `Msg`s to other modules using the generated `MsgClient` interfaces and
propose this mechanism as a replacement for the existing `Keeper` paradigm.

Using this `MsgClient` approach has the following benefits over keepers:
1. Protobuf types are checked for breaking changes using [buf](https://buf.build/docs/breaking-overview) and because of
the way protobuf is designed this will give us strong backwards compatibility guarantees while allowing for forward
evolution.
2. The separation between the client and server interfaces will allow us to insert permission checking code in between
the two which checks if one module is authorized to send the specified `Msg` to the other module providing a proper
object capability system.

### Module Keys

```go
func Invoker(ctx context.Context, signer ModuleID, method string, args, reply interface{}, opts ...grpc.CallOption) error

type ModuleKey interface {
  grpc.ClientConn
  ID() ModuleID
}

type RootModuleKey struct {
  moduleName string
  msgInvoker Invoker()
}

type DerivedModuleKey struct {
  moduleName string
  path []byte
  msgInvoker Invoker()
}

type ModuleID struct {
  ModuleName string
  Path []byte
}

func (key ModuleID) Address() []byte {
  return AddressHash(key.ModuleName, key.Path)
}
```

### Inter-Module Communication

```go
func (k keeper) DoSomething(ctx context.Context, req *MsgDoSomething) (*MsgDoSomethingResponse, error) {
  // make a query
  bankQueryClient := bank.NewQueryClient(sdk.ModuleQueryConn)
  res, err := bankQueryClient.Balance(ctx, &QueryBalanceRequest{
    Denom: "foo",
    Address: ModuleKeyToBech32Address(k.moduleKey),
  })
  
  // send a msg
  bankMsgClient := bank.NewMsgClient(k.moduleKey)
  res, err := bankMsgClient.Send(ctx, &MsgSend{
    FromAddress: ModuleKeyToBech32Address(k.moduleKey),
    ToAddress: ...,
    Amount: ...,
  })

  // send a msg from a derived module account
  derivedKey := k.moduleKey.DerivedKey([]byte("some-sub-pool"))
  res, err := bankMsgClient.Send(ctx, &MsgSend{
    FromAddress: ModuleKeyToBech32Address(derivedKey),
    ToAddress: ...,
    Amount: ...,
  })
}
```

### Hooks

```proto
service Hooks {
  rpc AfterValidatorCreated(AfterValidatorCreatedRequest) returns (AfterValidatorCreatedResponse);
}

message AfterValidatorCreatedRequest {
  string validator_address = 1;
}

message AfterValidatorCreatedResponse { }
```


```go
func (k stakingKeeper) CreateValidator(ctx context.Context, req *MsgCreateValidator) (*MsgCreateValidatorResponse, error) {
  ...

  for moduleId := range k.modulesWithHook {
    hookClient := NewHooksClient(moduleId)
    _, _ := hooksClient.AfterValidatorCreated(ctx, &AfterValidatorCreatedRequest {ValidatorAddress: valAddr})
  }
  ...
}
```

### Module Registration and Requirements

```go
type Configurator interface {
  ModuleKey() RootModuleKey

  MsgServer() grpc.Server
  QueryServer() grpc.Server
  HooksServer() grpc.Server

  RequireMsgServer(msgServerInterface interface{})
  RequireQueryServer(queryServerInterface interface{})
}

type Provisioner interface {
  GetAdminMsgClientConn(msgServerInterface interface{}) grpc.ClientConn
  GetPluginClientConn(pluginServerInterface interface{}) func(ModuleID) grpc.ClientConn
}

type Module interface {
  Configure(Configurator)
  Provision(Provisioner)
}

type ModuleManager interface {
  GrantAdminAccess(module ModuleID, msgServerInterface interface{})
  GrantPluginAccess(module ModuleID, pluginServerInterface interface{})
}
```


## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.


### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.


### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}


## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.


## References

- {reference link}
