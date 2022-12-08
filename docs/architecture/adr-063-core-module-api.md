# ADR 063: Core Module API

## Changelog

* 2022-08-18 First Draft
* 2022-12-08 First Draft

## Status

PROPOSED Not Implemented

## Abstract

A new core API is proposed as a way to develop cosmos-sdk applications that will eventually replace the existing
`AppModule` and `sdk.Context` frameworks a set of core services and extension interfaces. This core API aims to:
- be simpler,
- more extensible,
- more stable than the current framework,
- enable deterministic events and queries, 
- support event listeners and
- [ADR 033: Protobuf-based Inter-Module Communication](./adr-033-protobuf-inter-module-comm.md) clients.

## Context

Historically modules have exposed their functionality to the framework via the `AppModule` and `AppModuleBasic`
interfaces which have the following shortcomings:
* both `AppModule` and `AppModuleBasic` need to be defined and registered which is counter-intuitive
* apps need to implement the full interfaces, even parts they don't need (although there are workarounds for this),
* interface methods depend heavily on unstable third party dependencies, in particular Tendermint,
* legacy required methods have littered these interfaces for far too long

In order to interact with the state machine, modules have needed to do a combination of these things:
* get store keys from the app
* call methods on `sdk.Context` which contains more or less the full set of capability available to modules.

By isolating all the state machine functionality into `sdk.Context`, the set of functionalities available to
modules are tightly coupled to this type. If there are changes to upstream dependencies (such as Tendermint)
or new functionalities are desired (such as alternate store types), the changes need impact `sdk.Context` and all
consumers of it (basically all modules). Also, all modules now receive `context.Context` and need to convert these
to `sdk.Context`'s with a non-ergonomic unwrapping function.

Any breaking changes to these interfaces, such as ones imposed by third-party dependencies like Tendermint, have the
side effect of forcing all modules in the ecosystem to update in lock-step. This means it is almost impossible to have
a version of the module which can be run with 2 or 3 different versions of the SDK or 2 or 3 different versions of
another module. This lock-step coupling slows down overall development within the ecosystem and causes updates to
components to be delayed longer than they would if things were more stable and loosely coupled.

## Decision

The `core` API proposes a set of core APIs that modules can rely on to interact with the state machine and expose their
functionalities to it that are designed in a principled way such that:
* tight coupling of dependencies and unrelated functionalities is minimized or eliminated
* APIs can have long-term stability guarantees
* the SDK framework is extensible in a safe and straightforward way

The design principles of the core API are as follows:
* everything that a module wants to interact with in the state machine is a service
* all services coordinate state via `context.Context` and don't try to recreate the "bag of variables" approach of `sdk.Context`
* all independent services are isolated in independent packages with minimal APIs and minimal dependencies
* the core API should be minimalistic and designed for long-term support (LTS)
* a "runtime" module will implement all the "core services" defined by the core API and can handle all module
  functionalities exposed by core extension interfaces
* other non-core and/or non-LTS services can be exposed by specific versions of runtime modules or other modules 
following the same design principles, this includes functionality that interacts with specific non-stable versions of
third party dependencies such as Tendermint
* the core API doesn't implement *any* functionality, it just defines types
* go stable API compatibility guidelines are followed: https://go.dev/blog/module-compatibility

A "runtime" module is any module which implements the core functionality of composing an ABCI app, which is currently
handled by `BaseApp` and the `ModuleManager`. Runtime modules which implement the core API are *intentionally* separate
from the core API in order to enable more parallel versions and forks of the runtime module than is possible with the
SDK's current tightly coupled `BaseApp` design while still allowing for a high degree of composability and
compatibility.

Modules which are built only against the core API don't need to know anything about which version of runtime,
`BaseApp` or Tendermint in order to be compatible. Modules from the core mainline SDK could be easily composed
with a forked version of runtime with this pattern.

This design is intended to enable matrices of compatible dependency versions. Ideally a given version of any module
is compatible with multiple versions of the runtime module and other compatible modules. This will allow dependencies
to be selectively updated based on battle-testing. More conservative projects may want to update some dependencies
slower than more fast moving projects.

### Core Services

The following "core services" are defined by the core API. All valid runtime module implementations should provide
implementations of these services to modules via both [dependency injection](./adr-057-app-wiring-1.md) and
manual wiring. The individual services described below are all bundled in a convenient `appmodule.Service`
"bundle service" so that for simplicity modules can declare a dependency on a single service.

#### Store Services

Store services will be defined in the `cosmossdk.io/core/store` package.

The generic `store.KVStore` interface is the same as current SDK `KVStore` interface. Store keys have been refactored
into store services which, instead of expecting the context to know about stores, invert the pattern and allow
retrieving a store from a generic context. There are three store services for the three types of currently supported
stores - regular kv-store, memory, and transient:

```go
type KVStoreService interface {
    OpenKVStore(context.Context) KVStore
}

type MemoryStoreService interface {
    OpenMemoryStore(context.Context) KVStore
}
type TransientStoreService interface {
    OpenTransientStore(context.Context) KVStore
}
```

Modules can use these services like this:
```go
func (k msgServer) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
    store := k.kvStoreSvc.OpenKVStore(ctx)
}
```

Just as with the current runtime module implementation, modules will not need to explicitly name these store keys,
but rather the runtime module will choose an appropriate name for them and modules just need to request the
type of store they need in their dependency injection (or manual) constructors.

#### Event Service

The event `Service` and `Manager` will be defined in the `cosmossdk.io/core/event` package.

The event `Service` gives modules access to an event `Manager` which allows modules to emit typed and legacy,
untyped events:

```go
package event
type Service interface {
    GetManager(context.Context) Manager
}

type Manager interface {
    // Emit emits events to both clients and state machine listeners. These events MUST be emitted deterministically
    // and should be assumed to be part of blockchain consensus.
    Emit(proto.Message) error
    
    // EmitLegacy emits legacy untyped events to clients only. These events do not need to be emitted deterministically
    // and are not part of blockchain consensus.
    EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error

    // EmitClientOnly emits events only to clients. These events do not need to be emitted deterministically
    // and are not part of blockchain consensus.
    EmitClientOnly(proto.Message) error
}
```

Typed events emitted with `Emit` can be observed by other modules and thus must be deterministic and should be assumed
to be part of blockchain consensus (whether they are part of the block or app hash is left to the runtime to specify).

Events emitted by `EmitLegacy` and `EmitClientOnly` are not considered to be part of consensus and cannot be observed
by other modules. If there is a client-side need to add events in patch releases, these methods can be used.

Design questions:
* should we allow overriding event managers (like `sdk.Context.WithEventManager`)?

#### Block Info Service

The block info `Service` will be defined in the `cosmossdk.io/core/blockinfo` package.

The `BlockInfo` service allows modules to get a limited set of information about the currently executing block
that does not depend on any specific version of Tendermint:
```go
package blockinfo

type Service interface {
    GetBlockInfo(ctx context.Context) BlockInfo
}

type BlockInfo struct {
    ChainID string
    Height int64
    Time time.Time
    Hash []byte
}
```

Only a limited set of modules need any other information from the Tendermint block header and specific versions
of runtime modules tied to specific Tendermint versions can expose this functionality to modules that need it.
The basic `BlockInfo` service is left fairly generic to insulate the vast majority of other modules from unneeded
exposure to changes in the Tendermint protocol.

#### Gas Service

The gas `Service` and `Meter` types will be defined in the `cosmossdk.io/core/gas` package.

The gas service handles tracking of gas consumptions at the block and transaction level and also allows
for overriding the gas meter passed to child calls:
```go
package gas

type Service interface {
    GetMeter(context.Context) Meter
    GetBlockMeter(context.Context) Meter
    WithMeter(ctx context.Context, meter Meter) context.Context
    WithBlockMeter(ctx context.Context, meter Meter) context.Context
}
```

#### Inter-module Client

Runtime module implementations should provide an instance of the `appmodule.RootInterModuleClient` interface which
will allow modules to send messages to and make queries against other modules as described in [ADR 033](./adr-033-protobuf-inter-module-comm.md).

```go
type InterModuleClient interface {
    grpc.ClientConnInterface
    
    // Address is the ADR-028 address of this client against which messages will be authenticated.
    Address() []byte
}

type RootInterModuleClient interface {
    InterModuleClient
    
    DerivedClient(key []byte) InterModuleClient
}
```

The `RootInterModuleClient` allows a module to make inter-module calls using its root [ADR 028](./adr-028-public-key-addresses.md)
address (as returned by the `Address` method).

This client can be used with both `MsgClient`'s and `QueryClient`'s and the protobuf option
`cosmos.msg.v1.service` will allow the router to detect whether a given service is a query or msg service and
route requests appropriately.

Module clients that use [ADR 028](./adr-028-public-key-addresses.md) derived module addresses
can be created using the `DerivedClient` method.

By default, all `Msg`'s registered using the core API will be available for inter-module routing. Queries will only
be exposed for inter-module routing if their `rpc` methods have the protobuf option `cosmos.query.v1.internal`
set to true.

The inter-module router will provide a unified interface for sending `Msg`'s and queries
which ensures all required pre-processing steps are uniformly executed by all modules which need such functionality.

Runtime modules may choose to allow different types of "admin" access to individual modules which may allow for
bypassing certain authentication checks. This could allow `x/authz` to allow sending messages on behalf of all
addresses, for example.

#### `appmodule.Service` Bundle Service

To allow modules to declare a dependency on a single "bundle" service, runtime modules should provide an implementation
of the `appmodule.Service` interface:

```go
package appmodule

type Service interface {
    store.KVStoreService
    store.MemoryStoreService
    store.TransientStoreService
    event.Service
    blockinfo.Service
    gas.Service
    RootInterModuleClient
}
```
To maintain API compatibility, if new core services are added, a new `cosmossdk.io/core/appmodule/v2.Service` should
be added which extends this service and bundles the new core services, ex:

```go
package v2

import "cosmossdk.io/core/appmodule"

type Service interface {
    appmodule.Service
    SomeNewService
    AnotherNewService
}
```

### Core `AppModule` extension interfaces


Modules will provide their core services to the runtime module via extension interfaces built on top of the
`cosmossdk.io/core/appmodule.AppModule` tag interface. This tag interface requires only two empty methods which
allow `depinject` to identify implementors as `depinject.OnePerModule` types and as app module implementations:
```go
type AppModule interface {
  depinject.OnePerModuleType

  // IsAppModule is a dummy method to tag a struct as implementing an AppModule.
  IsAppModule()
}
```

Other core extension interfaces will be defined in `cosmossdk.io/core` should be supported by valid runtime
implementations.

#### `MsgServer` and `QueryServer` registration

The `Handler` struct type will implement `grpc.ServiceRegistrar` interface which will allow for registering
`MsgServer` and `QueryServer` implementations directly with the `Handler` struct like this:
```go
    h := &appmodule.Handler{}
    types.RegisterMsgServer(h, keeper.NewMsgServerImpl(am.keeper))
    types.RegisterQueryServer(h, am.keeper)
```

Service registrations will be stored in the `Handler.Services` field using the `ServiceImpl` struct:
```go
type ServiceImpl struct {
    Desc *grpc.ServiceDesc
    Impl interface{}
}
```

Because of the `cosmos.msg.v1.service` protobuf option, required for `Msg` services, the same handler struct can be
used to register both `Msg` and query services.

#### Genesis

The genesis `Handler` functions - `DefaultGenesis`, `ValidateGenesis`, `InitGenesis` and `ExportGenesis` - are specified
against the `GenesisSource` and `GenesisTarget` interfaces which will abstract over genesis sources which may be single
JSON files or collections of JSON objects that can be efficiently streamed. They will also include helper functions for
interacting with genesis data represented by single `proto.Message` types.

```go
type GenesisSource interface {
    ReadMessage(proto.Message) error
    OpenReader(field string) (io.ReadCloser, error)
    ReadRawJSON() (json.RawMessage, error)
}

type GenesisTarget interface {
    WriteMessage(proto.Message) error
    OpenWriter(field string) (io.WriteCloser, error)
    WriteRawJSON(json.RawMessage) error
}
```

All genesis objects for a given module are expected to conform to the semantics of a JSON object. That object may
represent a single genesis state `proto.Message` which is the case for most current modules, or it may represent a set
of nested JSON structures which may optionally be streamed as arrays from individual files which is the way the
[ORM](./adr-055-orm.md) is designed. This streaming genesis support may be beneficial for chains with extremely large
amounts of state that can't fit into memory.

Modules which use a single `proto.Message` can use the `GenesisSource.ReadMessage` and `GenesisTarget.ReadMessage`
methods without needing to deal with raw JSON or codecs.

Modules which use streaming genesis can use the
`GenesisSource.OpenReader` and `GenesisTarget.OpenWriter` methods. In the case of modules using the ORM, all JSON
handling is done automatically by the ORM and a single line will be sufficient for registering all of that functionality
(ex. `ormDb.RegisterGenesis(handler)`).

Low-level `ReadRawJSON` and `WriteRawJSON` are provided in case modules use some other sort of legacy genesis logic
we haven't anticipated.

#### Begin and End Blockers

The `BeginBlock` and `EndBlock` methods will simply take a `context.Context`, because:
* most modules don't need Tendermint information other than `BlockInfo` so we can eliminate dependencies on specific
Tendermint versions
* for the few modules that need Tendermint block headers and/or return validator updates, specific versions of the
runtime module will provide specific functionality for interacting with the specific version(s) of Tendermint
supported

In order for `BeginBlock`, `EndBlock` and `InitGenesis` to send back validator updates and retrieve full Tendermint
block headers, the runtime module for a specific version of Tendermint could provide services like this:
```go
type ValidatorUpdateService interface {
    SetValidatorUpdates(context.Context, []abci.ValidatorUpdate)
}

type BeginBlockService interface {
    GetBeginBlockRequest(context.Context) abci.RequestBeginBlock 
}
```

We know these types will change at the Tendermint level and that also a very limited set of modules actually need this
functionality, so they are intentionally kept out of core to keep core limited to the necessary, minimal set of stable
APIs.

#### Event Listeners

Handlers allow event listeners for typed events to be registered that will be called deterministically whenever those
events are called in the state machine. Event listeners are functions that take a context, the protobuf message type of
the event and optionally return an error. If an event listener returns a non-nil error, this error will fail the process
that emitted the event and revert any state transitions that this process would have caused.

Event listeners for typed events can be added to the handler struct either by populating the `EventListeners` field
or using generic helper methods:
```go
appmodule.AddEventListener(handler, func (ctx context.Context, e *EventReceive)) { ... })
```

Event listeners provide a standardized alternative to module hooks, such as `StakingHooks`, even though these hooks
are still supported and described in [ADR 057: App Wiring](./adr-057-app-wiring-1.md)

Only one event listener per event type can be defined per module. Event listeners will be called in a deterministic
order based on alphabetically sorting module names. If a more customizable order is desired, additional parameters
to the runtime module config object can be added to support optional custom ordering.

Only events emitted using `EventManager.Emit` are observable using event listeners.

#### Upgrade Handlers

Upgrade handlers can be specified using the `UpgradeHandlers` field that takes an array of `UpgradeHandler` structs:
```go
type UpgradeHandler struct {
    FromModule protoreflect.FullName
    Handler func(context.Context) error
}
```

In the [ADR 057: App Wiring](./adr-057-app-wiring-1.md) it is specified that each version of each module should be
identified by a unique protobuf message type called the module config object. For example for the consensus version 3
of the bank module, we would actually have a unique type `cosmos.bank.module.v3.Module` to identify this module.
In the [ModuleDescriptor](../proto/cosmos/app/v1alpha1/module.proto) for the module config object, `MigrateFromInfo`
objects should be provided which specify the full-qualified names of the modules that this module can migrate from.
For example, `cosmos.bank.module.v3.Module` might specify that it can migrate from `cosmos.bank.module.v2.Module`.
The app wiring framework will ensure that if the `ModuleDescriptor` specifies that a module can upgrade from another
module, an `UpgradeHandler` specifying that `FromModule` must be provided.

This pattern of identifying from module versions will allow a module to upgrade from a version it is not a direct
descendant of. For example, someone could fork staking and allow upgrading from the mainline SDK staking. The mainline
SDK staking could later merge some of the forked changes and allow upgrading from the fork back to mainline. This
is intended to allow smooth forking and merging patterns in the ecosystem for simpler and more diverse innovation.

A helper method on handler will be provided to simplify upgrade handler registration, ex:
```go
func (h *Handler) RegisterUpgradeHandler(fromModule protoreflect.FullName, handler func(context.Context) error)
```

#### Remaining Parts of AppModule

The current `AppModule` framework handles a number of additional concerns which aren't addressed by this core API.
These include the registration of:
* gogo proto and amino interface types
* cobra query and tx commands
* gRPC gateway 
* crisis module invariants
* simulations

The design proposed here relegates the registration of these things to other structures besides the `Handler` struct
defined here. In the case of gogo proto and amino interfaces, the registration of these generally should happen as early
as possible during initialization and in [ADR 057: App Wiring](./adr-057-app-wiring-1.md), protobuf type registration  
happens before dependency injection (although this could alternatively be done dedicated DI providers).

Commands should likely be handled at a different level of the framework as they are purely a client concern. Currently,
we use the cobra framework, but there have been discussions of potentially using other frameworks, automatically
generating CLI commands, and even letting libraries outside the SDK totally manage this.

gRPC gateway registration should probably be handled by the runtime module, but the core API shouldn't depend on gRPC
gateway types as 1) we are already using an older version and 2) it's possible the framework can do this registration
automatically in the future. So for now, the runtime module should probably provide some sort of specific type for doing
this registration ex:
```go
type GrpcGatewayInfo struct {
    Handlers []GrpcGatewayHandler
}

type GrpcGatewayHandler func(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error
```

which modules can return in a provider:
```go
func ProvideGrpcGateway() GrpcGatewayInfo {
    return GrpcGatewayinfo {
        Handlers: []Handler {types.RegisterQueryHandlerClient}
    }
}
```

Crisis module invariants and simulations are subject to potential redesign and should be managed with types
defined in the crisis and simulation modules respectively.

#### Example Usage

Here is an example of setting up a hypothetical `foo` v2 module which uses the [ORM](./adr-055-orm.md) for its state
management and genesis.

```go
func ProvideApp(config *foomodulev2.Module, evtSvc event.EventService, db orm.ModuleDB) (Keeper, *Handler){
    k := &Keeper{db: db, evtSvc: evtSvc}
    h := &Handler{}
    foov1.RegisterMsgServer(h, k)
    foov1.RegisterQueryServer(h, k)
    h.RegisterBeginBlocker(k.BeginBlock)
    db.RegisterGenesis(h)
    h.RegisterUpgradeHandler("foo.module.v1.Module", k.MigrateFromV1)
    return k, h
}
```

### Runtime Compatibility Version

The `core` module will define a static integer var, `cosmossdk.io/core.RuntimeCompatibilityVersion`, which is
a minor version indicator of the core module that is accessible at runtime. Correct runtime module implementations
should check this compatibility version and return an error if the current `RuntimeCompatibilityVersion` is higher
than the version of the core API that this runtime version can support. When new features are adding to the `core`
module API that runtime modules are required to support, this version should be incremented.

## Consequences

### Backwards Compatibility

Early versions of runtime modules should aim to support as much as possible modules built with the existing
`AppModule`/`sdk.Context` framework. As the core API is more widely adopted, later runtime versions may choose to
drop support and only support the core API plus any runtime module specific APIs (like specific versions of Tendermint).

The core module itself should strive to remain at the go semantic version `v1` as long as possible and follow design
principles that allow for strong long-term support (LTS).

### Positive

* better API encapsulation and separation of concerns
* more stable APIs
* more framework extensibility
* deterministic events and queries
* event listeners
* inter-module msg and query execution support
* more explicit support for forking and merging of module versions (including runtime)

### Negative

### Neutral

* modules will need to be refactored to use this API
* some replacements for `AppModule` functionality still need to be defined in follow-ups
  (type registration, commands, invariants, simulations) and this will take additional design work
* the upgrade module may need upgrades to support the new upgrade handler system described here

## Further Discussions

* how to register:
  * gogo proto and amino interface types
  * cobra commands
  * invariants
  * simulations
* should event and gas services allow callers to replace the event manager and gas meter in the context?
* should we allow a way for emitting typed events which aren't part of consensus?

## References

* [ADR 033: Protobuf-based Inter-Module Communication](./adr-033-protobuf-inter-module-comm.md)
* [ADR 057: App Wiring](./adr-057-app-wiring-1.md)
* [ADR 055: ORM](./adr-055-orm.md)
* [ADR 028: Public Key Addresses](./adr-028-public-key-addresses.md)
* [Keeping Your Modules Compatible](https://go.dev/blog/module-compatibility)
