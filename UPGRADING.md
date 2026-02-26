# Upgrade Reference

This document provides a reference for upgrading from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the [**App Wiring Changes**](#app-wiring-changes) section for more information on application wiring updates.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

## Summary

The release of Cosmos SDK v0.54.0 brings exciting new feature previews, an enhanced observability stack, bug fixes, and developer QoL improvements.

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

```go
	app.ModuleManager.SetOrderEndBlockers(
		banktypes.ModuleName,
        // other modules...
)
```

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

Cosmos SDK v0.54.0 drops support for the circuit, nft, and crisis modules. Developers can still use these modules; however, they will no longer be actively maintained by Cosmos Labs.

### x/circuit

The circuit module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/circuit`. 

### x/nft

The nft module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/nft`.

### x/crisis

The crisis module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/crisis`.

## Cosmos Enterprise

[Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) is Cosmos Labs' new enterprise offering, designed for teams operating production-grade Cosmos-based blockchain networks. It combines hardened protocol modules, on-premises and managed infrastructure components, and direct access to the engineers building the Cosmos technology stack.

### Groups Module

The groups module is now being maintained under the Cosmos Enterprise offering. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) to learn more about using the groups module in applications going forward.

### PoA Module

Cosmos SDK v0.54 includes a Proof of Authority (POA) module under the Cosmos Enterprise offering. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/components/poa/overview) to learn more about using the PoA module in your application.


## Module Version Updates

- `cosmossdk.io/client/v2` has been updated to v2.x.x ?? // TODO: Finalize this.
- `cosmossdk.io/api` has been updated to vx.x.x // TODO: Finalize this.

## Moved Go Modules

To improve maintainability and standardize import paths across Cosmos SDK modules and libraries, some packages have been consolidated under the github.com/cosmos/cosmos-sdk Go module. The following import paths must be updated:

- `cosmossdk.io/x/evidence` -> `github.com/cosmos/cosmos-sdk/x/evidence`
- `cosmossdk.io/x/feegrant` -> `github.com/cosmos/cosmos-sdk/x/feegrant` 
- `cosmossdk.io/x/upgrade` -> `github.com/cosmos/cosmos-sdk/x/upgrade`
- `cosmossdk.io/x/tx` -> `github.com/cosmos/cosmos-sdk/x/tx`
- `cosmossdk.io/systemtests` -> `github.com/cosmos/cosmos-sdk/testutil/systemtests`

## Log v2

The log package has been updated to `v2`. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/log/v2`. Usage of the logger itself does not need to be updated.
The v2 release of log adds contextual methods to the logger interface (`InfoContext`, `DebugContext`, etc.), allowing logs to be correlated with OpenTelemetry traces.
To learn more about the new features offered in `log/v2`, as well as setting up log correlation, see the log package's [README](log/README.md).

## Store v2

The store package has been updated to `v2`. Store v2 enables support for the new experimental packages: BlockSTM and IAVLX. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/store/v2`.

## Telemetry

The telemetry package has been deprecated and users are encouraged to switch to OpenTelemetry.

### Adoption of OpenTelemetry and Deprecation of `github.com/hashicorp/go-metrics`

Previously, Cosmos SDK telemetry support was provided by `github.com/hashicorp/go-metrics`, which was under-maintained and only supported metrics instrumentation.
OpenTelemetry provides an integrated solution for metrics, traces, and logging which is widely adopted and actively maintained.
The existing wrapper functions in the `telemetry` package required acquiring mutex locks and map lookups for every metric operation, which is suboptimal. OpenTelemetry's API uses atomic concurrency wherever possible and should introduce less performance overhead during metric collection.

See the [README.md](./telemetry/README.md) in the `telemetry` package to learn how to set up OpenTelemetry with Cosmos SDK v0.54.0+.

Below is a quick reference on setting up and using meters and traces with OpenTelemetry:

```go
package mymodule

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Declare package-level meter and tracer using otel.Meter() and otel.Tracer().
// Instruments should be created once at package initialization and reused.
var (
	tracer       = otel.Tracer("cosmos-sdk/x/mymodule")
	meter        = otel.Meter("cosmos-sdk/x/mymodule")
	txCounter    metric.Int64Counter
	latencyHist  metric.Float64Histogram
)

func init() {
	var err error
	txCounter, err = meter.Int64Counter(
		"mymodule.tx.count",
		metric.WithDescription("Number of transactions processed"),
	)
	if err != nil {
		panic(err)
	}
	latencyHist, err = meter.Float64Histogram(
		"mymodule.tx.latency",
		metric.WithDescription("Transaction processing latency"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic(err)
	}
}

// ExampleWithContext demonstrates tracing with a standard context.Context.
// Use tracer.Start directly when you have a Go context.
func ExampleWithContext(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "ExampleWithContext",
		trace.WithAttributes(attribute.String("key", "value")),
	)
	defer span.End()

	// Record metrics
	txCounter.Add(ctx, 1)

	if err := doWork(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// ExampleWithSDKContext demonstrates tracing with sdk.Context.
// Use ctx.StartSpan to properly propagate the span through the SDK context.
func ExampleWithSDKContext(ctx sdk.Context) error {
	ctx, span := ctx.StartSpan(tracer, "ExampleWithSDKContext",
		trace.WithAttributes(attribute.String("module", "mymodule")),
	)
	defer span.End()

	// Record metrics (sdk.Context implements context.Context)
	txCounter.Add(ctx, 1)

	// Create child spans for sub-operations
	ctx, childSpan := ctx.StartSpan(tracer, "ExampleWithSDKContext.SubOperation")
	// ... do sub-operation work ...
	childSpan.End()

	return nil
}
```

## Experimental Packages

For Q1 of 2026, Cosmos Labs has been focusing on greatly improving the performance of Cosmos SDK applications. v0.54 of Cosmos SDK introduces two performance-related experimental packages: IAVLX and BlockSTM.

NOTE: It is important to emphasize that these are **experimental** packages. We DO NOT recommend running chains with these packages enabled in production. Their inclusion in this release is for experimentation purposes only.

### IAVLX

IAVLX is a new, WAL-based, ACID storage engine for Cosmos applications. Currently, IAVLX is only available for new applications; we are actively working on IAVL v1 migration paths.
Developers interested in experimenting with IAVLX should read the documentation [here](link/to/docs).

Below is an example of setting up IAVLX:

:::warning
IAVLX is experimental. Test thoroughly before production use. Code for migrating from IAVL v1 is not yet available.
:::

1. Add the following imports
```go
import (
    "github.com/cosmos/cosmos-sdk/baseapp"
    "github.com/cosmos/cosmos-sdk/iavlx"
)
```

2. Create a BaseApp option to configure IAVLX:

```go
func setupIAVLXStore(appOpts servertypes.AppOptions) func(*baseapp.BaseApp) {
    return func(app *baseapp.BaseApp) {
        homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
        dbPath := filepath.Join(homeDir, "data", "iavlx")
		
        opts := &iavlx.Options{
            EvictDepth:            20,
            ReaderUpdateInterval:  1,
            WriteWAL:              true,
            MinCompactionSeconds:  30,
            RetainVersions:        1,
            CompactWAL:            true,
            DisableCompaction:     false,
            CompactionOrphanAge:   200,
            CompactionOrphanRatio: 0.95,
            CompactAfterVersions:  2000,
            ChangesetMaxTarget:    2147483648,
            ZeroCopy:              true,
            FsyncInterval:         1000,
        }

        db, err := iavlx.LoadDB(dbPath, opts, logger)
        if err != nil {
            panic(err)
        }

        app.SetCMS(db)
    }
}
```

3. Add the option to BaseApp initialization

```go
baseAppOptions = append(baseAppOptions, setupIAVLXStore(appOpts))

bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
```

### BlockSTM

BlockSTM enables deterministic, concurrent execution of transactions, improving block execution speed by up to X%. // TODO: REAL NUMBER
Developers interested in experimenting with BlockSTM should read the documentation [here](link/to/docs).

Below is an example of setting up BlockSTM:

:::warning
BlockSTM is experimental. Ensure thorough testing before enabling in production.
:::

```go
import (
    "runtime"
	
    "github.com/cosmos/cosmos-sdk/baseapp/blockstm"
)

oKeys := storetypes.NewObjectStoreKeys(banktypes.ObjectStoreKey)

keys := storetypes.NewKVStoreKeys(
    authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
    // ... other store keys
)

// Collect non-transient store keys
var nonTransientKeys []storetypes.StoreKey
for _, k := range keys {
    nonTransientKeys = append(nonTransientKeys, k)
}
for _, k := range oKeys {
    nonTransientKeys = append(nonTransientKeys, k)
}

// Enable BlockSTM runner
bApp.SetBlockSTMTxRunner(blockstm.NewSTMRunner(
    txConfig.TxDecoder(),
    nonTransientKeys,
    min(runtime.GOMAXPROCS(0), runtime.NumCPU()),
    true,  // debug logging
    sdk.DefaultBondDenom,
))

// Optionally disable block gas meter for better performance
bApp.SetDisableBlockGasMeter(true)

// Set ObjectStoreKey on bank module
app.BankKeeper = app.BankKeeper.WithObjStoreKey(oKeys[banktypes.ObjectStoreKey])
```

## Upgrade Handler

The following is an example upgrade handler for upgrading from **v0.53.6** to **v0.54.0**.

```go
const UpgradeName = "v0.53.6-to-v0.54.0"

func (app SimApp) RegisterUpgradeHandlers() {
    app.UpgradeKeeper.SetUpgradeHandler(
        UpgradeName,
        func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
            return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
        },
    )

    if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
      storeUpgrades := storetypes.StoreUpgrades{
        Added: []string{},
      }
      // configure store loader that checks if version == upgradeHeight and applies store upgrades
      app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
    }
}
```