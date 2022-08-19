# ADR 061: Core Module API

## Changelog

* 2022-08-18 First Draft

## Status

PROPOSED Not Implemented

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

Historically modules have exposed their functionality to the state machine via the `AppModule` and `AppModuleBasic`
interfaces which have the following shortcomings:
* both `AppModule` and `AppModuleBasic` need to be defined which is counter-intuitive
* apps need to implement the full interfaces, even parts they don't need
* interface methods depend heavily on unstable third party dependencies, in particular Tendermint
* legacy required methods have littered these interfaces for far too long

In order to interact with the state machine, modules have needed to do a combination of these things:
* get store keys from the app
* call methods on `sdk.Context` which contains more or less the full set of capability available to modules.

By isolating all the state machine functionality into `sdk.Context`, the set of functionalities available to
modules are tightly coupled to this type. If there are changes to upstream dependencies (such as Tendermint)
or new functionalities are desired (such as alternate store types), the changes need impact `sdk.Context` and all
consumers of it (basically all modules). Also, all modules now receive `context.Context` and need to convert these
to `sdk.Context`'s with a non-ergonomic unwrapping function.

## Decision

The `core` API proposes a set of core APIs that modules can rely on to interact with the state machine and expose their
functionalities to it that are designed in a principled way such that:
* tight coupling of dependencies and unrelated functionalities is minimized or eliminated
* APIs can have long-term stability guarantees
* the SDK framework is extensible in a safe a straightforward way

The design principles of the core API are as follows:
* everything that a module wants to interact with in the state machine is a service
* all services coordinate state via `context.Context` and don't try to recreate the "bag of variables" approach of `sdk.Context`
* all independent services are isolated in independent packages with minimal APIs and minimal dependencies
* the core API should be minimalistic and designed for long-term support (LTS)
* a "runtime" module will implement all the "core services" defined by the core API and can handle all module
  functionalities exposed by the core `Handler` type
* other non-core and/or non-LTS services can be exposed by specific versions of runtime modules or other modules 
following the same design principles, this includes functionality that interacts with specific non-stable versions of
third party dependencies such as Tendermint
* go stable API management principles are followed (TODO: link)

### Core Services

The following "core services" are defined by the core API. All valid runtime module implementations should provide
implementations of these services to modules via both [dependency injection](./adr-057-app-wiring-1.md) and
manual wiring.

#### Store Services

The generic store interface is defined as the current SDK `KVStore` interface. The `StoreService` type is a refactoring
of the existing store key types which inverts the relationship with the context. Instead of expecting a
"bag of variables" context type to explicitly know about stores, `StoreService` uses the general-purpose
`context.Context` just to coordinate state:

```go
type StoreService interface {
    // Open retrieves the KVStore from the context.
	Open(context.Context) KVStore
}
```

Modules can use these services like this:
```go
func (k msgServer) Send(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	store := k.storeService.Open(ctx)
}
```

Three specific types of store services (kv-store, memory and transient) are defined by the core API that modules can
get a reference to via dependency injection or manually:

```go
type KVStoreService interface {
    StoreService
	IsKVStoreService()
}

type MemoryStoreService interface {
    StoreService
	IsMemoryStoreService()
}

type TransientStoreService interface {
	StoreService
	IsTransientStoreService()
}
```

Just as with the current runtime module implementation, modules will not need to explicitly name these store keys,
but rather the runtime module will choose an appropriate name for them and modules just need to request the
type of store they need in their dependency injection (or manual) constructors.

#### Event Service

The event service gives modules access to an event manager which allows modules to emit typed and legacy,
untyped events:

```go
package event
type Service interface {
    GetManager(context.Context) Manager
}

type Manager interface {
    Emit(proto.Message) error
    EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error
}
```

By definition, typed events emitted with `Emit` are part of consensus and can be observed by other modules while
legacy events emitted by `EmitLegacy` are not considered to be part of consensus and cannot be observed by other
modules.

Design questions:
* should we allow overriding event managers (like `sdk.Context.WithEventManager`)?
* should there be a way to "silently" emit typed events that aren't part of consensus?

#### Block Info Service

The `BlockInfo` service allows modules to get a limited set of information about the currently executing block
that does not depend on any specific version of Tendermint:
```go
package blockinfo

type Service interface {
    GetBlockInfo(ctx context.Context) BlockInfo
}

type BlockInfo interface {
    ChainID() string
    Height() int64
    Time() *timestamppb.Timestamp
    Hash() []byte
}
```

Only a limited set of modules need any other information from the Tendermint block header and specific versions
of runtime modules tied to specific Tendermint versions can expose this functionality to modules that need it.
The basic `BlockInfo` service is left fairly generic to insulate the vast majority of other modules from unneeded
exposure to changes in the Tendermint protocol.

#### Gas Service

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

#### `grpc.ClientConnInterface`

Runtime module implementations should provide an instance of `grpc.ClientConnInterface` as core service. This service
will allow modules to send messages to and make queries against other modules as described in [ADR 033](./adr-033-protobuf-inter-module-comm.md).

A single `grpc.ClientConnInterface` will be used for both MsgClient's and QueryClient's and the state machine will
make the presence or absence of the protobuf option `cosmos.msg.v1.service` which will be required for all
valid `Msg` service types.

It will likely be necessary to define which queries and `Msg`'s are available for inter-module communication, possibly
with an `internal` protobuf option.

The router used by the `grpc.ClientConnInterface` will provide a unified interface for sending `Msg`'s and queries
which ensures all required pre-processing steps are uniformly executed by all modules which need such functionality.

### Core `Handler` Struct

Modules will provide their core services to runtime module via the `Handler` struct instead of the existing
`AppModule` interface:
```go
type Handler struct {
    // Services are the Msg and Query services for the module. Msg services
    // must be annotated with the option cosmos.msg.v1.service = true.
    Services []ServiceImpl
	
	DefaultGenesis(GenesisTarget)
    ValidateGenesis(GenesisSource) error
    InitGenesis(context.Context, GenesisSource) error
    ExportGenesis(context.Context, GenesisTarget)

    BeginBlocker func(context.Context) error
	EndBlocker func(context.Context) error
	
	EventListeners []EventListener
	
	UpgradeHandlers []UpgradeHandler
}
```

A struct as opposed to an interface has been chosen for the following reasons:
* it is always possible to add new fields to a struct without breaking backwards compatibility
* it is not necessary to populate all fields in a struct, whereas interface methods must always be implemented
* compared to extension interfaces, struct fields are more explicit

Helper methods will be added to the `Handler` struct to simplify construction of more complex fields.

#### `MsgServer` and `QueryServer` registration

The `Handler` struct type will implement `grpc.ServiceRegistrar` interface which will allow for registering
`MsgServer` and `QueryServer` implementations directly with the `Handler` struct like this:
```go
    h := &appmodule.Handler{}
	types.RegisterMsgServer(h, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(h, am.keeper)
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
or using the `Handler.AddEventListener` builder method:
```go
handler.AddEventListener(func (ctx context.Context, e *EventReceive) error) { ... })
```

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

A helper method on handler will be provided to simplify upgrade handler registration, ex:
```go
func (h *Handler) RegisterUpgradeHandler(fromModule protoreflect.FullName, handler func(context.Context) error)
```

#### Remaining Parts of AppModule


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

TODO:
* interfaces and legacy amino registration
* commands
* grpc gateway
* invariants
* inter-module hooks
* simulations

### Runtime Compatibility Version

### Legacy `AppModule` Runtime Compatibility

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

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
