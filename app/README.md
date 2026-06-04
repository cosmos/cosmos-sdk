# SDK App Package

The `app` package provides a reusable `SDKApp` implementation that wires standard Cosmos SDK modules and baseapp setup behind a focused configuration API.

## Quick Start

```go
cfg := app.DefaultSDKAppConfig("MyApp", appOpts, baseAppOptions...)
sdkApp := app.NewSDKApp(logger, db, traceStore, cfg)
sdkApp.LoadModules()
```

## Execution Configuration

`SDKAppConfig` exposes two execution controls:

* `BlockSTM *BlockSTMConfig`
    * `nil` means serial execution.
    * non-nil enables BlockSTM parallel execution.
* `OptimisticExecutionEnabled bool`
    * enables BaseApp optimistic execution.

These settings are mutually exclusive. `Validate()` returns an error if both are enabled.

## Optional Modules

Use the module toggles on `SDKAppConfig` to include or exclude optional modules:

* `WithAuthz`
* `WithFeeGrant`
* `WithMint`
* `WithEpochs`

When a module is disabled, the config removes it from module account permissions and lifecycle ordering slices.

## Extending With Custom Modules

After constructing the app, you can register custom modules with:

```go
if err := sdkApp.AddModules(customModules...); err != nil {
    panic(err)
}
```

Then call `LoadModules()` exactly once to build module managers, register services, handlers, and mount stores.
