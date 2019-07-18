# ADR 4: Multiple Genesis modes

## Changelog

* 07/18/2019: Initial draft

## Context

### Problem

Currently, each some of the modules `GenesisState` are populated if the corresponding
object passed through the Genesis JSON is empty (_i.e_ as an empty array) or if it's
undefined or the value is zero.

This leads to confussion as you migh pass a valid genesis file which contains zero/empty
value sand have `InitGenenesis` perform this uncessesary check.

### Proposed Solution

The proposed solution is to have two ways of initializing the daemon:

1. **New (empty) chain**: the user provides the minimum amount of genesis fields required to
start a chain. The rest is loaded directly from the mandatory genesis parameters.
2. **Restarted chain**: the user provides all the fields from genesis.

If an empty chain from `1.` recives _any_ field that should be populated, the initialization
of the app should `panic`. Same applies for `2.` if a single field is missing.

The fields that are not required are the ones in general are checked with simulation
invariants.

## Decision

Add a flag `FlagPopulateGenesis = "populate-genesis"` on `server/start.go` that is
passed to the `AppCreator` (eg: `NewSimApp` or `NewGaiaApp`) when `<app>d start` is called.
This tells the app to fill the missing genesis properties with the mandatory fields.

<!-- TODO: Add BaseApp options changes -->

```go
// populateGenesis is the parameter added. It tells the module to populate or not the
// missing values of the genesis from what's given on the state
func (am AppModule) InitGenesis(ctx sdk.Context, populateGenesis bool, data json.RawMessage)
```

## Status

Proposed

## Consequences

<!-- TODO: -->

### Positive

### Negative

### Neutral


## References

* [#2862](https://github.com/cosmos/cosmos-sdk/issues/2862)
* [#4568](https://github.com/cosmos/cosmos-sdk/issues/4568)
