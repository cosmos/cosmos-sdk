<!--
order: 13
-->

# Upgrading Modules

[In-Place Store Migrations](../core/upgrade.html) allow your modules to upgrade to new versions that include breaking changes. This document outlines how to build modules to take advantage of this functionality. {synopsis}

## Prerequisite Readings

- [In-Place Store Migration](../core/upgrade.md) {prereq}

## Consensus Version

Successful upgrades of existing modules require each `AppModule` to implement the function `ConsensusVersion() uint64`.

- The versions must be hard-coded by the module developer.
- The initial version **must** be set to 1.

Consensus versions serve as state-breaking versions of app modules and must be incremented when the module introduces breaking changes.

## Registering Migrations

To register the functionality that takes place during a module upgrade, you must register which migrations you want to take place.

Migration registration takes place in the `Configurator` using the `RegisterMigration` method. The `AppModule` reference to the configurator is in the `RegisterServices` method.

You can register one or more migrations. If you register more than one migration script, list the migrations in increasing order and ensure there are enough migrations that lead to the desired consensus version. For example, to migrate to version 3 of a module, register separate migrations for version 1 and version 2 as shown in the following example:

```golang
func (am AppModule) RegisterServices(cfg module.Configurator) {
    // --snip--
    cfg.RegisterMigration(types.ModuleName, 1, func(ctx sdk.Context) error {
        // Perform in-place store migrations from ConsensusVersion 1 to 2.
    })
     cfg.RegisterMigration(types.ModuleName, 2, func(ctx sdk.Context) error {
        // Perform in-place store migrations from ConsensusVersion 2 to 3.
    })
}
```

Since these migrations are functions that need access to a Keeper's store, use a wrapper around the keepers called `Migrator` as shown in this example:

+++ https://github.com/cosmos/cosmos-sdk/blob/6ac8898fec9bd7ea2c1e5c79e0ed0c3f827beb55/x/bank/keeper/migrations.go#L8-L21

## Writing Migration Scripts

To define the functionality that takes place during an upgrade, write a migration script and place the functions in a `migrations/` directory. For example, to write migration scripts for the bank module, place the functions in `x/bank/migrations/`. Use the recommended naming convention for these functions. For example, `v043bank` is the script that migrates the package `x/bank/migrations/v043`:

```golang
// Migrating bank module from version 1 to 2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v043bank.MigrateStore(ctx, m.keeper.storeKey) // v043bank is package `x/bank/migrations/v043`.
}
```

To see example code of changes that were implemented in a migration of balance keys, check out [migrateBalanceKeys](https://github.com/cosmos/cosmos-sdk/blob/36f68eb9e041e20a5bb47e216ac5eb8b91f95471/x/bank/legacy/v043/store.go#L41-L62). For context, this code introduced migrations of the bank store that updated addresses to be prefixed by their length in bytes as outlined in [ADR-028](../architecture/adr-028-public-key-addresses.md).
