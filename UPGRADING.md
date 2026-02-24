# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).


## Summary

The release of Cosmos SDK v0.54.0 brings exciting new feature previews, an enhanced observability stack, and tightens the 
developer experience of building an application with Cosmos SDK.

## App Wiring Changes

### x/gov

#### Keeper Initialization

The `x/gov` module has been decoupled from `x/staking`. The `keeper.NewKeeper` constructor now requires a `CalculateVoteResultsAndVotingPowerFn` parameter instead of a `StakingKeeper`.

**Before:**
```go
govKeeper := govkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[govtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.StakingKeeper, // REMOVED IN v0.54
    app.DistrKeeper,
    app.MsgServiceRouter(),
    govConfig,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

**After:**
```go
govKeeper := govkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[govtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.DistrKeeper,
    app.MsgServiceRouter(),
    govConfig,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(app.StakingKeeper), // ADDED IN v0.54
)
```

For applications using depinject, the governance module now accepts an optional `CalculateVoteResultsAndVotingPowerFn`. If not provided, it will use the `StakingKeeper` (also optional) to create the default function.

#### GovHooks Interface

The `AfterProposalSubmission` hook now includes the proposer address as a parameter.

**Before:**
```go
func (h MyGovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
    // implementation
}
```

**After:**
```go
func (h MyGovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64, proposerAddr sdk.AccAddress) error {
    // implementation
}
```

### x/epochs

The epochs module's `NewAppModule` function now requires the epoch keeper by pointer instead of value, fixing a bug related to setting hooks via depinject.

### x/bank

The bank module now contains an `EndBlock` method to support the new BlockSTM experimental package. All applications, whether using BlockSTM or not, must add `x/bank`'s `ModuleName` to the `ModuleManager`'s `SetOrderEndBlockers` method as the first entry.

### NodeService

The node service has been updated to return the node's earliest store height in the `Status` query. Please update your registration with the following code (make sure you are already updated to `cosmossdk.io/store/v2`):

```go
func (app *SimApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, func() int64 {
		return app.CommitMultiStore().EarliestVersion()
	})
}
```

## Module Deprecations

Cosmos SDK v0.54.0 drops support for the circuit, nft, and crisis modules. Developers can still use these modules,
however, they will no longer be actively maintained by Cosmos Labs.

### x/circuit

The circuit module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/circuit`. 

### x/nft

The nft module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/nft`.

### x/crisis

The crisis module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/crisis`.

## Cosmos Enterprise

Cosmos Enterprise is Cosmos Labs' new enterprise offering, designed for teams operating production-grade Cosmos-based blockchain networks. It combines hardened protocol modules, on-premises and managed infrastructure components, and direct access to the engineers building the Cosmos technology stack.

### Groups Module

The groups module is now being maintained under the Cosmos Enterprise offering. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) to learn more about using the groups module in applications going forward.

### PoA Module

Cosmos SDK v0.54 includes a Proof of Authority (POA) module under the Cosmos Enterprise offering. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) to learn more about using the PoA module in your application.


## Moved Go Modules

To improve maintainability and unify the import paths of Cosmos SDK's module offerings, all `cosmossdk.io/x/<module>` modules have been moved to the main `github.com/cosmos/cosmos-sdk` go module. The following import paths must be updated:

- `cosmossdk.io/x/evidence` -> `github.com/cosmos/cosmos-sdk/x/evidence`
- `cosmossdk.io/x/feegrant` -> `github.com/cosmos/cosmos-sdk/x/feegrant` 
- `cosmossdk.io/x/upgrade` -> `github.com/cosmos/cosmos-sdk/x/upgrade`
- `cosmossdk.io/x/tx` -> `github.com/cosmos/cosmos-sdk/x/tx`

## Log v2

The log package has been updated to `v2`. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/log/v2`. Usage of the logger itself does not need to be updated.
The v2 release of log adds contextual methods to the logger interface (InfoContext, DebugContext, etc.), allowing logs to be correlated with OpenTelemetry traces. We recommend scraping logs with OpenTelemetry's [FileLog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver).
To learn more about the new features offered in `log/v2`, as well as setting up log correlation, see the log package's [README](log/README.md).

## Store v2

The store package has been updated to `v2`. Store v2 enables support for the new experimental packages: BlockSTM and IAVLX. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/store/v2`.

## Client v2

The `cosmossdk.io/client/v2` package has been updated to ??? to support the new log and store releases. // TODO: THIS MIGHT BE V3???

## Telemetry

The telemetry package has been deprecated and users are encouraged to switch to OpenTelemetry.

## Adoption of OpenTelemetry and Deprecation of `github.com/hashicorp/go-metrics`

Previously, Cosmos SDK telemetry support was provided by `github.com/hashicorp/go-metrics` which was undermaintained and only supported metrics instrumentation.
OpenTelemetry provides an integrated solution for metrics, traces, and logging which is widely adopted and actively maintained.
The existing wrapper functions in the `telemetry` package required acquiring mutex locks and map lookups for every metric operation which is sub-optimal. OpenTelemetry's API uses atomic concurrency wherever possible and should introduce less performance overhead during metric collection.

The [README.md](telemetry/README.md) in the `telemetry` package provides more details on usage, but below is a quick summary:
1. Application developers should follow the official [go OpenTelemetry](https://pkg.go.dev/go.opentelemetry.io/otel) guidelines when instrumenting their applications.
2. Node operators who want to configure OpenTelemetry exporters should set up their otel configuration file in one of two ways:
   - Set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of the otel configuration yaml file.
   - Fill out the automatically generated `otel.yaml` file in `<node_home_dir>/config/otel.yaml`.

Either of these yaml files should follow the OpenTelemetry declarative configuration format specified here: https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf. As long as the `telemetry` package has been imported somewhere (it should already be imported if you are using the SDK), OpenTelemetry will be initialized automatically based on the configuration file.

NOTE: the go implementation of [otelconf](https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf) is still under development, and we will update our usage of it as it matures.

## Experimental Packages

For Q1 of 2026, Cosmos Labs has been focusing on greatly improving performance of Cosmos SDK applications. v0.54 of Cosmos SDK introduces two performance related experimental packages: IAVLX and BlockSTM.

NOTE: It is important to emphasize that these are **experimental** packages. We DO NOT recommend running chains with these packages enabled in production. Their inclusion in this release is for experimentation purposes only.

### IAVLX

IAVLX is a new, WAL-based, ACID storage engine for Cosmos applications. Currently, IAVLX is only available for new applications; we are actively working on IAVL v1 migration paths.
Developers interested in experimenting with IAVLX should read the documentation [here](link/to/docs).

### BlockSTM

BlockSTM enables deterministic, concurrent execution of transactions, improving block execution speeds by up to X%.
Developers interested in experimenting with BlockSTM should read the documentation [here](link/to/docs).
