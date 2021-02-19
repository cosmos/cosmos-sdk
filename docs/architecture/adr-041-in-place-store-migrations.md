# ADR 041: In-Place Store Migrations

## Changelog

- 17.02.2021: Initial Draft

## Status

Proposed

## Abstract

This ADR introduces a mechanism to perform automatic (no manual process involved) in-place store migrations during chain upgrades.

## Context

When a chain upgrade introduces state-breaking changes inside modules, the current procedure consists of exporting the whole state into a JSON file (via the `simd export` command), running migration scripts on the JSON file (`simd migrate` command), clearing the stores (`simd unsafe-reset-all` command), and starting a new chain with the migrated JSON file as new genesis (optionally with a custom initial block height). An example of such a procedure can be seen [in the Cosmos Hub 3->4 migration guide](https://github.com/cosmos/gaia/blob/v4.0.3/docs/migration/cosmoshub-3.md#upgrade-procedure).

While [cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/v0.41.1/cosmovisor) aims to alleviate the difficulty of handling upgrades, this procedure is still cumbersome for multiple reasons:

- The procedure takes time. It can take hours to run the `export` command, plus some additional hours to run `InitChain` on the fresh chain using the migrated JSON.
- The exported JSON file can be heavy (~100MB-1GB), making it difficult to view, edit and transfer.

## Decision

We propose a migration procedure based on modifying the KV store in-place. This procedure does not manipulate intermediary JSON files.

### Module `ConsensusVersion`

We introduce a new method on the `AppModule` interface:

```go
type AppModule interface {
    // --snip--
    ConsensusVersion() uint64
}
```

This methods returns an `uint64` which serves as state-breaking versioning of the module. It should be incremented on each consensus-breaking change introduced by the module. To avoid potential errors with default values, the initial version of a module MUST be set to 1. In the SDK, version 1 corresponds to the modules in the v0.41 series.

### Module-Specific Migration Scripts

For each consensus-breaking change introduced by the module, a migration script from ConsensusVersion `k` to version `k+1` sMUST be registered in the `Configurator` using its `RegisterMigration` method.

```go
configurator.RegisterMigration("bank", 1, func(ctx sdk.Context) error {
    // Perform x/banks's in-place store migrations from ConsensusVersion 1 to 2.
})
```

For example, if the current ConsensusVersion of a module is `N` , then `N-1` migration scripts MUST be registered in the configurator.

In the SDK, the migration scripts are handled by each module's keeper, because the keeper holds the `sdk.StoreKey` used to perform in-place store migrations. A `MigrationKeeper` interface is implemented by each keeper:

```go
// MigrationKeeper is an interface that the keeper implements for handling
// in-place store migrations.
type MigrationKeeper interface {
	// Migrate1 migrates the store from version 1 to 2.
	Migrate1(ctx sdk.Context) error
}
```

Since migration scripts manipulate legacy code, they SHOULD live inside the `legacy/` folder of each module, and be called by the keeper's implementation of `MigrationKeeper`.

```go
func (keeper BankKeeper) Migrate1(ctx sdk.Context) error {
	return v042bank.MigrateStore(ctx, keeper.storeKey) // v042bank is package `x/bank/legacy/v042`.
}
```

Each module's migration scripts are specific to the module's store evolutions, an example of x/bank store key migration due to the introduction of ADR-028 length-prefixed addresses can be seen [here](https://github.com/cosmos/cosmos-sdk/blob/ef8dabcf0f2ecaf26db1c6c6d5922e9399458bb3/x/bank/legacy/v042/store.go#L15).

### Tracking Module Versions in `x/upgrade`

We introduce a new prefix store in `x/upgrade`'s store. This store will track each module's current version.

### Running Migrations

Once all the migration handlers are registered inside the configurator (this happens at startup), running migrations can happen any time by calling the `RunMigrations` method on `module.Manager`.

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

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

- {reference link}
