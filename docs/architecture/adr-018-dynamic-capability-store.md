# ADR 18: Dynamic Capability Store

## Changelog

- 12 December 2019: Initial version

## Context

Full implementation of the [IBC specification](https://github.com/cosmos/ics) requires the ability to create and authenticate object-capability keys at runtime (i.e., during transaction execution),
as described in [ICS 5](https://github.com/cosmos/ics/tree/master/spec/ics-005-port-allocation#technical-specification). In the IBC specification, capability keys are created for each newly initialised
port & channel, and are used to authenticate future usage of the port or channel. Since channels and potentially ports can be initialised during transaction execution, the state machine must be able to create
object-capability keys at this time.

At present, the Cosmos SDK does not have the ability to do this. Object-capability keys are currently pointers (memory addresses) of `sdk.StoreKey` structs created at application initialisation in `app.go` ([example](https://github.com/cosmos/gaia/blob/master/app/app.go#L132))
and passed to Keepers as fixed arguments ([example](https://github.com/cosmos/gaia/blob/master/app/app.go#L160)). Keepers cannot create or store capability keys during transaction execution — although they could call `sdk.NewKVStoreKey` and take the memory address
of the returned struct, storing this in the Merklised store would result in a consensus fault, since the memory address will be different on each machine (this is intentional — were this not the case, the keys would be predictable and couldn't serve as object capabilities).

Keepers need a way to keep a private map of store keys which can be altered during transacton execution, along with a suitable mechanism for regenerating the unique memory addresses (capability keys) in this map whenever the application is started or restarted.
This ADR proposes such an interface & mechanism.

## Decision

The SDK will include a new `CapabilityKeeper` abstraction, which is responsible for provisioning, tracking, and authenticating capabilities at runtime. During application initialisation in `app.go`, the `sdk.CapabilityKeeper` will
be hooked up to modules through unique function references so that it can identify the calling module when later invoked. When the initial state is loaded from disk, the `sdk.CapabilityKeeper` instance will create new capability keys
for all previously allocated capability identifiers (allocated during execution of past transactions), and keep them in a memory-only store while the chain is running. The SDK will include a new `MemoryStore` store type, similar
to the existing `TransientStore` but without erasure on `Commit()`, which this `CapabilityKeeper` will use to privately store capability keys.

The `sdk.CapabilityKeeper` will use two stores: a regular, persistent `sdk.KVStore`, which will track what capabilities have been created by each module, and an in-memory `sdk.MemoryStore` (described below), which will
store the actual capabilities. The `sdk.CapabilityKeeper` will define the following functions:

```golang
type Capability interface {
  Name() string
  String() string
}
```

```golang
type CapabilityKeeper struct {
  persistentKey sdk.StoreKey
  memoryKey sdk.MemoryStoreKey
  moduleNames map[string]interface{}
}
```

```golang
func (ck CapabilityKeeper) NewCapability(ctx sdk.Context, name string) Capability {
  // check name not taken in memory store
  // set forward map in memory store from capability to name
  // set backward mamp in memory store from name to capability
}
```

```golang
func (ck CapabilityKeeper) AuthenticateCapability(name string, capability Capability) bool {
  // check forward map in memory store === name
}
```

```golang
func (ck CapabilityKeeper) Initialise(ctx sdk.Context) {
  // initialise memory store for all names in persistent store
}
```

The `CapabilityKeeper` also provides the ability to create *scoped* sub-keepers which are tied to a particular module name. These `ScopedCapabilityKeeper`s must be created at application
initialisation and passed to modules, which can then use them to claim capabilities they receive and retrieve capabilities which they own by name.

```golang
type ScopedCapabilityKeeper struct {
  capabilityKeeper CapabilityKeeper
  moduleName string
}
```

`ScopeToModule` is used to create a scoped sub-keeper with a particular name, which must be unique.

```golang
func (ck CapabilityKeeper) ScopeToModule (moduleName string) {
  if _, present := ck.moduleNames[moduleName]; present {
    panic("cannot create multiple scoped capability keepers for the same module name")
  }
  ck.moduleNames[moduleName] = interface{}
  return ScopedCapabilityKeeper{
    capabilityKeeper: ck,
    moduleName: moduleName
  }
}
```

`ClaimCapability` allows a module to claim a capability key which it has received (perhaps by calling `NewCapability`, or from another module), so that future `GetCapability` calls will succeed.

`ClaimCapability` MUST be called, even if `NewCapability` was called by the same module. Capabilities are single-owner, so if multiple modules have a single `Capability` reference, the last module
to call `ClaimCapability` will own it. To avoid confusion, a module which calls `NewCapability` SHOULD either call `ClaimCapability` or pass the capability to another module which will then claim it.

```golang
func (sck ScopedCapabilityKeeper) ClaimCapability(capability Capability) {
  // fetch name from memory store
  // set name to module in persistent store
}
```

`GetCapability` allows a module to fetch a capability which it has previously claimed by name. The module is not allowed to retrieve capabilities which it does not own.

```golang
func (sck ScopedCapabilityKeeper) GetCapability(ctx sdk.Context, name string) (Capability, error) {
  // fetch name from persistent store, check === module
  // fetch capability from memory store, return it
}
```

`keeper.GetCapability(name: string) -> capability` (exposed in closure version with module name)

`keeper.NewCapability(name: string) -> capability` (exposed in closure version with module name)

`keeper.Initialise(ctx sdk.Context)` (actually generates capabilities)

- Capability transfer? Either have a temporary identifier (never stored) or add `keeper.TransferCapability` (but then you need the other module name) or have `keeper.ClaimCapability` for unique use.

### Memory store

- A new store type, `StoreTypeMemory`, will be added to the `store` package.
- A new store key type, `MemoryStoreKey`, will be added to the `store` package.
- The memory store will work just like the current transient store, except that it will not create a new `dbadapter.Store` when `Commit()` is called, but instead retain the current one
  (so that state will persist across blocks).
- Initially the memory store will only be used by the `sdk.CapabilityKeeper`, but it could be used by other modules in the future.

## Status

Proposed.

## Consequences

### Positive

- Dynamic capability support.

### Negative

- Additional implementation complexity.
- Possible confusion between capability store & regular store.

### Neutral

(none known)

## References

- [Original discussion](https://github.com/cosmos/cosmos-sdk/pull/5230#discussion_r343978513)
