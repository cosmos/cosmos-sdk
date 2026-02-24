# Upgrade Guide

This document provides a guide for upgrading a Cosmos SDK chain from `v0.53.x` to `v0.54.x`.

## Table of Contents

* [Go Mod Changes](#go-mod-changes)
* [Keeper Changes](#keeper-changes)
    * [EpochsKeeper Pointer Type](#epochskeeper-pointer-type)
* [Experimental Features](#experimental-features)
    * [BlockSTM Parallel Execution](#blockstm-parallel-execution)
    * [IAVLX Storage Backend](#iavlx-storage-backend)
* [Upgrade Handler](#upgrade-handler)

## Go Mod Changes

The following imports must be updated:

- `cosmossdk.io/x/evidence` -> `github.com/cosmos/cosmos-sdk/x/evidence`
- `cosmossdk.io/x/feegrant` -> `github.com/cosmos/cosmos-sdk/x/feegrant`
- `cosmossdk.io/x/upgrade` -> `github.com/cosmos/cosmos-sdk/x/upgrade`
- `cosmossdk.io/x/tx` -> `github.com/cosmos/cosmos-sdk/x/tx`

It is recommended to use a find and replace tool to do a global replacement of these modules.

## New EndBlocker in Bank Module

The Bank module now contains an EndBlocker method that must be registered with the ModuleManager:

```go
app.ModuleManager.SetOrderEndBlockers(
    banktypes.ModuleName, // IMPORTANT: Must be the first entry.
	// other modules..
)
```

## NodeService Change

The NodeService's `Status` query now returns the Earliest store height the node can service in queries. Registering the node service now requires a function that returns the earliest store height.

NOTE: This method will not be available unless you have updated your `cosmossdk.io/store` dependency to `cosmossdk.io/store/v2`.

```go

func (app *SimApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, func() int64 {
		return app.CommitMultiStore().EarliestVersion()
	})
}
```

## Keeper Changes

### GovKeeper Constructor Change


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

### AfterProposalSubmission New Parameter


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

### EpochsKeeper Pointer Type

The `epochs.NewAppModule` function now takes a pointer to the EpochsKeeper.

**Update ModuleManager:**

```go
// BEFORE (v0.53.x):
epochs.NewAppModule&app.EpochsKeeper)

// AFTER (v0.54.x):
epochs.NewAppModule(&app.EpochsKeeper)  // No & needed, already a pointer
```

## OpenTelemetry

 TODO

## Experimental Features

The following features are experimental and optional. They can significantly improve performance but should be thoroughly tested before production use.

### BlockSTM Parallel Execution

BlockSTM enables parallel transaction execution using software transactional memory. This can significantly improve throughput for workloads with independent transactions.

:::warning
BlockSTM is experimental. Ensure thorough testing before enabling in production.
:::

```go
import (
    "runtime"
    
    "github.com/cosmos/cosmos-sdk/baseapp/blockstm"
)
```

```go
    // setup ObjectStoreKeys
    oKeys := storetypes.NewObjectStoreKeys(banktypes.ObjectStoreKey
```

```go
// Collect non-transient store keys
var nonTransientKeys []storetypes.StoreKey
for _, k := range keys {
    nonTransientKeys = append(nonTransientKeys, k)
}
for _, k := range oKeys {
    nonTransientKeys = append(nonTransientKeys, k)
}

// Enable BlockSTM
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

### IAVLX Storage Backend

IAVLX is an experimental, high-performance storage backend.

:::warning
IAVLX is experimental. Test thoroughly before production use. Code for migrating from IAVL v1 is not yet available.
:::

```go
import (
    "github.com/cosmos/cosmos-sdk/baseapp"
    "github.com/cosmos/cosmos-sdk/iavlx"
)
```

Create a BaseApp option to configure IAVLX:

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

Add the option to BaseApp initialization:

```go
baseAppOptions = append(baseAppOptions, iavlxStorage(appOpts))

bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
```

## Upgrade Handler

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
