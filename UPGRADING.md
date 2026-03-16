# Upgrade Reference

This document provides a reference for upgrading from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the [App Wiring Changes](#app-wiring-changes) section for more information on application wiring updates.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

## Table of Contents

- [Upgrade Checklist](#upgrade-checklist)
- [Required Changes](#required-changes)
    - [App Wiring Changes](#app-wiring-changes)
        - [x/gov](#xgov)
            - [Keeper Initialization](#keeper-initialization)
            - [GovHooks Interface](#govhooks-interface)
        - [x/epochs](#xepochs)
        - [x/bank](#xbank)
        - [NodeService](#nodeservice)
    - [Moved Go Modules](#moved-go-modules)
    - [Module Version Updates](#module-version-updates)
    - [Log v2](#log-v2)
    - [Store v2](#store-v2)
- [Conditional Changes](#conditional-changes)
    - [Module Deprecations](#module-deprecations)
        - [x/circuit](#xcircuit)
        - [x/nft](#xnft)
        - [x/crisis](#xcrisis)
    - [Cosmos Enterprise](#cosmos-enterprise)
        - [Groups Module](#groups-module)
        - [PoA Module](#poa-module)
- [New Features and Non-Breaking Changes](#new-features-and-non-breaking-changes)
    - [Telemetry](#telemetry)
        - [OpenTelemetry](#opentelemetry)
    - [Centralized Authority via Consensus Params](#centralized-authority-via-consensus-params)
        - [How AuthorityParams Works](#how-authorityparams-works)
- [Upgrade Handler](#upgrade-handler)
- [Experimental Features](#experimental-features)
    - [libp2p](#libp2p)
    - [BlockSTM](#blockstm)

## Upgrade Checklist

Use this checklist first, then read the linked sections for the exact code or wiring changes.

- [ ] Update `x/gov` keeper wiring, as the `x/gov` module has been decoupled from `x/staking`. See [Keeper Initialization](#keeper-initialization).
- [ ] Update your governance hooks if you implement `AfterProposalSubmission`. See [GovHooks Interface](#govhooks-interface).
- [ ] Update `x/epochs.NewAppModule` if your app includes `x/epochs`. See [x/epochs](#xepochs).
- [ ] Put `x/bank` first in `SetOrderEndBlockers`. See [x/bank](#xbank).
- [ ] Update your node service registration if your app exposes `NodeService`. See [NodeService](#nodeservice).
- [ ] Migrate imports for moved `x/` Go modules. See [Moved Go Modules](#moved-go-modules).
- [ ] Update required Cosmos SDK Go module dependencies. See [Module Version Updates](#module-version-updates).
- [ ] Migrate to `contrib/` imports if you use `x/circuit`, `x/nft`, or `x/crisis`. See [Module Deprecations](#module-deprecations).
- [ ] Migrate to Cosmos Enterprise if you use the `x/group` module. See [Groups Module](#groups-module).
- [ ] Update imports to `cosmossdk.io/log/v2` if your app imports the log package directly. See [Log v2](#log-v2).
- [ ] Review Store v2 changes if this release requires your app to adopt `cosmossdk.io/store/v2`. See [Store v2](#store-v2).
- [ ] Review [Centralized Authority via Consensus Params](#centralized-authority-via-consensus-params). No upgrade action is required to keep using per-keeper authorities.
- [ ] Review [Telemetry](#telemetry). No upgrade action is required to keep existing telemetry wiring, but upgrading to OpenTelemetry is strongly encouraged.
- [ ] Review [PoA Module](#poa-module) if you are interested in adopting the new Cosmos Enterprise Proof of Authority module.
- [ ] Review [Experimental Features](#experimental-features) if you are interested in enabling `libp2p` or `BlockSTM`.

## Required Changes

All chains upgrading to `v0.54.x` should review and apply the changes in this section.

### App Wiring Changes

#### x/gov

##### Keeper Initialization

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

##### GovHooks Interface

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

#### x/epochs

The epochs module's `NewAppModule` function now requires the epoch keeper by pointer instead of value, fixing a bug related to setting hooks via depinject.

#### x/bank

The bank module now contains an `EndBlock` method to support the new BlockSTM experimental package. BlockSTM requires coordinating object store access across parallel execution workers, and `x/bank`'s `EndBlock` handles the finalization step for that. **All applications must make this change**, whether or not they enable BlockSTM, because the `EndBlock` registration is now part of the module's standard lifecycle.

```go
	app.ModuleManager.SetOrderEndBlockers(
		banktypes.ModuleName,
        // other modules...
)
```

#### NodeService

The node service has been updated to return the node's earliest store height in the `Status` query. Please update your registration with the following code (make sure you are already updated to `cosmossdk.io/store/v2`):

```go
func (app *SimApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, func() int64 {
		return app.CommitMultiStore().EarliestVersion()
	})
}
```

### Moved Go Modules

Most `cosmossdk.io` vanity URLs for modules under `x/` have been removed. These separate Go modules caused dependency version management to be unpredictable; different modules could be pinned to different SDK versions, leading to compatibility issues. Consolidating everything under `github.com/cosmos/cosmos-sdk` gives developers a single, versioned dependency to manage.

A migration tool ships alongside this release to automate updating these import paths. The following must be updated manually or via the tool:

- `cosmossdk.io/x/evidence` -> `github.com/cosmos/cosmos-sdk/x/evidence`
- `cosmossdk.io/x/feegrant` -> `github.com/cosmos/cosmos-sdk/x/feegrant` 
- `cosmossdk.io/x/upgrade` -> `github.com/cosmos/cosmos-sdk/x/upgrade`
- `cosmossdk.io/x/tx` -> `github.com/cosmos/cosmos-sdk/x/tx`
- `cosmossdk.io/systemtests` -> `github.com/cosmos/cosmos-sdk/testutil/systemtests`

### Module Version Updates

- `cosmossdk.io/client/v2` has been updated to v2.x.x ?? // TODO: Finalize this.
- `cosmossdk.io/api` has been updated to vx.x.x // TODO: Finalize this.

### Log v2

The log package has been updated to `v2`. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/log/v2`. Usage of the logger itself does not need to be updated.
The v2 release of log adds contextual methods to the logger interface (InfoContext, DebugContext, etc.), allowing logs to be correlated with OpenTelemetry traces.
To learn more about the new features offered in `log/v2`, as well as setting up log correlation, see the [log package documentation](https://docs.cosmos.network/sdk/next/learn/advanced/log).

### Store v2

The store package has been updated to `v2`. Store v2 introduces a new async, deferred commit model that is the foundation for both BlockSTM parallel execution and the upcoming IAVLX storage engine — the deferred commit path is what makes concurrent transaction execution safe and allows the WAL-based design in IAVLX. Applications using v0.54.0+ of Cosmos SDK will be required to update imports to `cosmossdk.io/store/v2`.

## Conditional Changes

These changes apply if your chain uses the affected modules, packages, or integrations.

### Module Deprecations

Cosmos SDK v0.54.0 drops support for the circuit, nft, and crisis modules. Developers can still use these modules,
however, they will no longer be actively maintained by Cosmos Labs.

#### x/circuit

The circuit module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/circuit`. 

#### x/nft

The nft module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/nft`.

#### x/crisis

The crisis module is no longer being actively maintained by Cosmos Labs and was moved to `contrib/x/crisis`.

### Cosmos Enterprise

[Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) is Cosmos Labs' new enterprise offering, designed for teams operating production-grade Cosmos-based blockchain networks. It combines hardened protocol modules, on-premises and managed infrastructure components, and direct access to the engineers building the Cosmos technology stack.

#### Groups Module

The groups module is now maintained under the Cosmos Enterprise offering. If your application uses `x/group`, you will need to migrate your code to the Enterprise-distributed package and obtain a Cosmos Enterprise license to continue using it. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/overview) to learn more.

#### PoA Module

Cosmos SDK v0.54 includes a Proof of Authority (POA) module under the Cosmos Enterprise offering. Please see [Cosmos Enterprise](https://docs.cosmos.network/enterprise/components/poa/overview) to learn more about using the PoA module in your application.

## New Features and Non-Breaking Changes

These changes are informational and optional to adopt during the upgrade, but they are not required for a successful migration.

### Telemetry

The telemetry package has been deprecated and users are encouraged to switch to OpenTelemetry.

#### OpenTelemetry

Previously, Cosmos SDK telemetry support was provided by `github.com/hashicorp/go-metrics` which was undermaintained and only supported metrics instrumentation.

OpenTelemetry provides an integrated solution for metrics, traces, and logging which is widely adopted and actively maintained.

The existing wrapper functions in the `telemetry` package required acquiring mutex locks and map lookups for every metric operation which is suboptimal. OpenTelemetry's API uses atomic concurrency wherever possible and should introduce less performance overhead during metric collection.

See the [telemetry documentation](https://docs.cosmos.network/sdk/next/learn/advanced/telemetry) to learn how to set up OpenTelemetry with Cosmos SDK v0.54.0+. 
<!-- todo: update link with correct docs version tag -->


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

### Centralized Authority via Consensus Params

Authority management can now be centralized via the `x/consensus` module. A new `AuthorityParams` field in `ConsensusParams` stores the authority address on-chain. When set, it takes precedence over the per-keeper authority parameter.

**This feature introduces no breaking changes**: Keeper constructors still accept the `authority` parameter. It is now used as a **fallback** when no authority is configured in consensus params. Existing code continues to work without changes.

#### How AuthorityParams Works

When a module validates authority (e.g., in `UpdateParams`), it checks consensus params first. If no authority is set there, it falls back to the keeper's `authority` field:

```go
authority := sdkCtx.Authority() // from consensus params
if authority == "" {
    authority = k.authority       // fallback to keeper field
}
if authority != msg.Authority {
    return nil, errors.Wrapf(...)
}
```

To enable centralized authority, set the `AuthorityParams` in consensus params via a governance proposal targeting the `x/consensus` module's `MsgUpdateParams`.

## Upgrade Handler

This section provides a reference example for implementing the on-chain upgrade itself.

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

## Experimental Features

For Q1 of 2026, Cosmos Labs has been focusing on greatly improving performance of Cosmos SDK applications. v0.54 of Cosmos SDK introduces support for several performance-related features across the stack, including BlockSTM in the SDK and Libp2p support for CometBFT. 

NOTE: It is important to emphasize that the following are **experimental** features. We DO NOT recommend running chains with these features enabled in production. The inclusion in this release is for experimentation purposes only.

### libp2p

libp2p replaces CometBFT's legacy `comet-p2p` transport layer with [go-libp2p](https://libp2p.io/). Unlike other opt-in features, **to opt-in to libp2p, every validator in the network must upgrade together**. CometBFT p2p and libp2p are fundamentally incompatible and cannot interoperate. Because of this, a coordinated network-wide migration at a specific upgrade height is required. Mixed deployments are not supported.

See the [libp2p documentation](https://docs.cosmos.network/cometbft/next/docs/experimental/lib-p2p) for details.
<!-- todo: update link with correct docs version tag -->

### BlockSTM

BlockSTM enables deterministic, concurrent execution of transactions, improving block execution speeds by up to X%. // TODO: REAL NUMBER
Developers interested in experimenting with BlockSTM should read the [documentation](https://docs.cosmos.network/sdk/next/experimental/blockstm). 
<!-- todo: update link with correct docs version tag -->

Below is an example of setting up BlockSTM:

> **⚠️ Warning:** BlockSTM is experimental. Ensure thorough testing before enabling in production.

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
