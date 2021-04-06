<!--
order: 15
-->

# In-Place Store Migrations

::: warning
Please make sure you read this whole document and fully understand in-place store migrations before running them on a live chain.
:::

Upgrade your app modules smoothly with custom in-place migration logic. {synopsis}

The Cosmos SDK currently has two methods to perform upgrades. The first method is by exporting the entire application state to a JSON file using the `export` CLI command, making changes, and then starting a new binary with the changed JSON file as the genesis file. More details on this method can be found in the [chain upgrade guide](../migrations/chain-upgrade-guide-040.md#upgrade-procedure). The second method, introduced in v0.43, works by performing upgrades in place, significantly decreasing the time needed to perform upgrades for chains with a larger state. The following guide will provide you with the necessary information in order to setup your application to take advantage of in-place upgrades.

## Tracking Module Versions

Each module gets assigned a consensus version by the module developer, which serves as the breaking change version of the module. The SDK keeps track of all modules' consensus versions in the x/upgrade's `VersionMap` store. During an upgrade, the SDK calculates the difference between the old `VersionMap` stored in state and the new `VersionMap`. For each difference found, the SDK will run module-specific migrations and increment the respective consensus version of each upgraded module.

The next paragraphs detail each component of the in-place store migration process, and gives instructions on how to update your app to take advantage of this functionality.

## Genesis State

When starting a new chain, each module's consensus version must be saved to state during the application's genesis. This can be done by adding the following line to the `InitChainer` method in `app.go`

```diff
func (app *MyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
  ...
+ app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
  ...
}
```

Using this information, the SDK will be able to detect when modules with newer versions are introduced to the app. 

### Consensus Version
The consensus version is defined on each app module by the module developer. It serves as the breaking change version of the module. The consensus version helps to inform the SDK on which modules need to be upgraded. For example, if the bank module was version 2 and an upgrade introduces bank module 3, the SDK will know to upgrade the bank module and run its "version 2 to 3" migration script.

### Version Map
The version map is a mapping of module names to consensus versions. The map is persisted to x/upgrade's state for use during in-place migrations. When migrations finish, the updated version map is persisted to state. 

## Upgrade Handlers

Upgrades utilize an `UpgradeHandler` to facilitate migrations. `UpgradeHandler`s are functions implemented by the app developer that conform to the following function signature. These functions retrieve the `VersionMap` from x/upgrade's state, and return the new `VersionMap` to be stored in x/upgrade after the upgrade. The diff between the two `VersionMap`s determines which modules need upgrading.

```golang
type UpgradeHandler func(ctx sdk.Context, plan Plan, fromVM VersionMap) (VersionMap, error)
```

Inside these functions, you should perform any upgrade logic you wish to include in the provided `plan`. All upgrade handler functions should end with the following line of code:

```golang
  return app.mm.RunMigrations(ctx, cfg, fromVM)
```

## Running Migrations

Migrations are run inside of an `UpgradeHandler` via `app.mm.RunMigrations(ctx, cfg, vm)`. As described above, `UpgradeHandler`s are functions which describe the functionality to occur during an upgrade. The `RunMigration` function will loop through the `VersionMap` argument, and run the migration scripts for any versions that are less than the new binary's app module versions. Once the migrations are finished, a new `VersionMap` will be returned to persist the upgraded module versions to state.

```golang
cfg := module.NewConfigurator(...)
app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

    // ...
    // do upgrade logic
    // ...

    // RunMigrations returns the VersionMap
    // with the updated module ConsensusVersions
    return app.mm.RunMigrations(ctx, vm)
})
```

To learn more about configuring migration scripts for your modules, refer to this [guide](../building-modules/upgrade.md).

## Adding New Modules In Upgrades

Entirely new modules can be introduced to the application during an upgrade. The SDK recognizes new modules during upgrades and will call the corresponding module's `InitGenesis` function to setup its initial state. This can be skipped if the module does not require any inital state. Otherwise, it is important to implement `InitGenesis` for new modules to successfully upgrade your application without error.

In the scenario where your application does not need any inital state via `InitGenesis`, you must take extra steps to ensure `InitGenesis` is skipped to avoid errors. To do so, you simply need to update the value of the module version in the `VersionMap` in the `UpgradeHandler`. 

```go
// Foo is a new module being introduced
// in this upgrade plan
import foo "github.com/my/module/foo"

app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap)  (module.VersionMap, error) {
    // We make sure to set foo's version to the latest ConsensusVersion in the VersionMap.
    // This will skip running InitGenesis on Foo
    vm["foo"] = foo.AppModule{}.ConsensusVersion()

    return app.mm.RunMigrations(ctx, cfg, vm)
})
```

Using a similar method, you can also run InitGenesis on your new module with a custom genesis state:

```go
import foo "github.com/my/module/foo"

app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap)  (module.VersionMap, error) {
    vm["foo"] = foo.AppModule{}.ConsensusVersion()

    // Run custom InitGenesis for foo
    app.mm["foo"].InitGenesis(ctx, app.appCodec, myCustomGenesisState)

    return app.mm.RunMigrations(ctx, cfg, vm)
})
```
