<!--
order: 15
-->

# In-Place Store Migrations

Upgrade your app modules smoothly with custom in-place migration logic.

The Cosmos SDK currently has two ways to perform upgrades. The first way is by exporting the entire application state to a JSON file using the `simd export` CLI command, making changes, and then starting a new binary with the changed JSON file as the genesis file. The second way is by performing upgrades in place, significantly decreasing the time needed to perform upgrades for chains with a larger state. The following will guide you on how to setup your application to take advantage of the second method described above.

# Enabling Upgrades

To enable your application to conform to the upgrade module's specifications, a few changes need to be made to your application.

# Genesis State

Each app module's consensus version must be saved to state on the application's genesis. This can be done by adding the following line to the `InitChainer` method in `app.go`

```diff
func (app *MyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
  ...
+ app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
  ...
}
```

# Upgrade Handlers

Upgrades utilize an `UpgradeHandler` to facilitate migrations. `UpgradeHandler`s are functions implemented by the app developer that conform to the following function signature.

```golang
type UpgradeHandler func(ctx sdk.Context, plan Plan, versionMap VersionMap) (VersionMap, error)
```

# Running Migrations

In practice, the handlers should simply call and return the values from the `app.mm.RunMigrations` function. The `RunMigrations` function should be passed the `VersionMap` from the `UpgradeHandler`. With this, the `RunMigration` function will loop through the `VersionMap`, and for any current app module who's consensus version is greater than its corresponding value in the `VersionMap`, have its migration scripts ran. To learn how to configure migration scripts, refer to (this guide)[../building-modules/upgrade.md].
