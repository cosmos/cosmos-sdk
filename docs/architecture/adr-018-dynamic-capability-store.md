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

Need some notion of module identity, denoted by passing specific closures over module name to `sdk.CapabilityKeeper` and keepers in `app.go`.

`sdk.CapabilityKeeper`

`keeper.AuthenticateCapability(name: string, cap: capability) -> bool`

`keeper.GetCapability(name: string) -> capability` (exposed in closure version with module name)

`keeper.NewCapability(name: string) -> capability` (exposed in closure version with module name)

`keeper.Initialise(ctx sdk.Context)` (actually generates capabilities)

- Capability transfer? Either have a temporary identifier (never stored) or add `keeper.TransferCapability` (but then you need the other module name) or have `keeper.ClaimCapability` for unique use.

## Status

Proposed.

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

- Dynamic capability support.

### Negative

- Additional implementation complexity.
- Possible confusion between capability store & regular store.

### Neutral

(none known)

## References

- [Original discussion](https://github.com/cosmos/cosmos-sdk/pull/5230#discussion_r343978513)
