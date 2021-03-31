<!--
order: 15
-->

# In-Place Store Migrations

Upgrade your app modules smoothly with custom in-place migration logic. {synopsis}

The Cosmos SDK currently has two ways to perform upgrades. The first way is by exporting the entire application state to a JSON file using the `simd export` CLI command, making changes, and then starting a new binary with the changed JSON file as the genesis file. The second way is by performing upgrades in place, significantly decreasing the time needed to perform upgrades for chains with a larger state. The following will guide you on how to setup your application to take advantage of the second method described above.

## Enabling Upgrades

To enable your application to conform to the upgrade module's specifications, a few changes need to be made to your application.

## Genesis State

Each app module's consensus version must be saved to state on the application's genesis. This can be done by adding the following line to the `InitChainer` method in `app.go`

```diff
func (app *MyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
  ...
+ app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
  ...
}
```

### Consensus Version
The consensus version is defined on each app module by the module developer. It serves as the breaking change version of the module. 

### Version Map
The version map is a mapping of module names to consensus versions. The map is persisted to state for use during in-place migrations. 

## Upgrade Handlers

Upgrades utilize an `UpgradeHandler` to facilitate migrations. `UpgradeHandler`s are functions implemented by the app developer that conform to the following function signature.

```golang
type UpgradeHandler func(ctx sdk.Context, plan Plan, versionMap VersionMap) (VersionMap, error)
```

## Running Migrations

In practice, the handlers should simply call and return the values from the `app.mm.RunMigrations` function. The `RunMigrations` function should be passed the `VersionMap` from the `UpgradeHandler`. With this, the `RunMigration` function will loop through the `VersionMap`, and for any current app module who's consensus version is greater than its corresponding value in the `VersionMap`, have its migration scripts ran. To learn how to configure migration scripts, refer to (this guide)[../building-modules/upgrade.md].

When upgrades are executed, they refer to the functionality described in an upgrade handler. All upgrade handlers should describe the logic needed for the upgrade plan, and end with returning the values from a call to `app.RunMigrations(ctx, vm)`. This will return the updated `VersionMap` to be saved to the upgrade module's store.

```golang
app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

    // ...
    // do upgrade logic
    // ...

    // RunMigrations returns the VersionMap
    // with the updated module ConsensusVersions
    return app.RunMigrations(ctx, vm)
})
```

## Adding New Modules In Upgrades

New modules can be introduced to the application during an upgrade. The SDK recognizes new modules during upgrades and will call the corresponding module's `DefaultGenesis` function to setup the its initial state. This can be skipped if the module does not require any inital state. 

If you wish to overwrite the default behavior of running InitGenesis during an upgrade for new modules, make sure to pass the latest `ConsensusVersion` of the new module into the returned `module.VersionMap`. This will then skip running InitGenesis for the module:

```go
// Foo is a new module being introduced
// in this upgrade plan
import foo "github.com/my/module/foo"

app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap)  (module.VersionMap, error) {
    // We make sure to set foo's version to the latest ConsensusVersion in the VersionMap.
    // This will skip running InitGenesis on Foo
    vm["foo"] = foo.AppModule{}.ConsensusVersion()

    return app.mm.RunMigrations(ctx, vm)
})
```

Using the same method, you can also run InitGenesis on your new module with a custom genesis state:

```go
app.UpgradeKeeper.SetUpgradeHandler("my-plan", func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap)  (module.VersionMap, error) {
    vm["foo"] = foo.AppModule{}.ConsensusVersion()

    // Run custom InitGenesis for foo
    app.mm["foo"].InitGenesis(ctx, app.appCodec, myCustomGenesisState)

    return app.mm.RunMigrations(ctx, vm)
})
```