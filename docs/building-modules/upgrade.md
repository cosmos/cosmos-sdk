<!--
order: 13
-->

# In-Place Store Migrations

In-place store migrations allow your modules to smoothly transition to new versions with breaking changes. This document outlines how to build modules to take advantage of this functionality. 

# Consensus Version

In order to successfully upgrade your existing modules, your `AppModule`s must implement the function `ConsensusVersion() uint64`. The `uint64` returned will serve as the consensus version and should be hard coded by the module developer. This number will serve as a state-breaking version of each app module, so it *MUST* be incremented on each consensus-breaking change introduced by the module. The initial version *MUST* be set to 1.

# Registering Migrations

To register the functionality that takes place during a module upgrade, we must register which migrations we want to take place. This takes place in the `Configurator` via the `RegisterMigration` method. The `AppModule`s have a reference to the configurator in the `RegisterServices` method. We can register a number of migrations here, however, if more than one migration script is registered, it is important they are listed in increasing order. Additionally, we must ensure there are enough migrations that will lead to the desired consensus version. For example, if we wanted to migrate to version 3 of a module, we would need to register a separate migration for both version 1 and 2 as shown below.

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

Since these migrations are functions that need access to a Keeper's store, we use a wrapper around the keepers called `Migrator`. An example of this can be found (here)[https://github.com/cosmos/cosmos-sdk/blob/master/x/bank/keeper/migrations.go]. In addition to the `Migrator` wrapper, we also define our migration scripts. More on that below.

# Writing Migration Scripts

In order to define the functionality that takes place during an upgrade, we will write a migration script. Since migration scripts will manipulate legacy code, we place these functions in a `legacy/` directory. For example, if we wanted to write migration scripts for a module named `bank`, we would place the functions in `x/bank/legacy/`. We recommend the following naming convention for these functions:

```golang
// Migrating bank module from version 1 to 2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v043bank.MigrateStore(ctx, m.keeper.storeKey) // v043bank is package `x/bank/legacy/v043`.
}
```

If you would like to see example code of changes implemented in a migration, you can check out the code [here](https://github.com/cosmos/cosmos-sdk/blob/36f68eb9e041e20a5bb47e216ac5eb8b91f95471/x/bank/legacy/v043/store.go#L41-L62). For context, this introduced migrations of the bank store that updated addresses to be prefixed by their length in bytes as perscribed in (ADR-028)[https://docs.cosmos.network/master/architecture/adr-028-public-key-addresses.html].
