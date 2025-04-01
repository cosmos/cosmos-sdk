# Upgrade Guide

This document provides a full guide for upgrading a Cosmos SDK chain from `v0.50.x` to `v0.53.x`.

This guide includes one **required** change and three **optional** features.

After completing this guide, applications will have:

- The `x/protocolpool` module
- The `x/epochs` module
- Unordered Transaction support

## Table of Contents

- [App Wiring Changes](#app-wiring-changes)
- [Adding ProtocolPool Module](#adding-protocolpool-module)
  - [Manual Wiring](#protocolpool-manual-wiring)
  - [DI Wiring](#protocolpool-di-wiring)
- [Adding Epochs Module](#adding-epochs-module)
  - [Manual Wiring](#epochs-manual-wiring)
  - [DI Wiring](#epochs-di-wiring)
- [Enable Unordered Transactions](#enable-unordered-transactions)
- [Upgrade Handler](#upgrade-handler)

## App Wiring Changes **REQUIRED**

The `x/auth` module now contains a `PreBlocker` that _must_ be set in the module manager's `SetOrderPreBlockers` method.

```go
app.ModuleManager.SetOrderPreBlockers(
    upgradetypes.ModuleName,
    authtypes.ModuleName, // NEW
)
```

## Adding ProtocolPool Module **OPTIONAL**

### Manual Wiring

Import the following:

```go
import (
    // ...
    "github.com/cosmos/cosmos-sdk/x/protocolpool"
    protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
    protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)
```

Set the module account permissions.

```go
maccPerms = map[string][]string{
    // ...
    protocolpooltypes.ModuleName:                nil,
    protocolpooltypes.ProtocolPoolEscrowAccount: nil,
}
```

Add the protocol pool keeper to your application struct.

```go
ProtocolPoolKeeper protocolpoolkeeper.Keeper
```

Add the store key:

```go
keys := storetypes.NewKVStoreKeys(
    // ...
    protocolpooltypes.StoreKey,
)
```

Instantiate the keeper.

Make sure to do this before the distribution module instantiation, as you will pass the keeper there next.

```go
app.ProtocolPoolKeeper = protocolpoolkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[protocolpooltypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

Pass the protocolpool keeper to the distribution keeper:

```go
app.DistrKeeper = distrkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.StakingKeeper,
    authtypes.FeeCollectorName,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper), // NEW
)
```

Add the protocolpool module to the module manager:

```go
app.ModuleManager = module.NewManager(
    // ...
    protocolpool.NewAppModule(appCodec, app.ProtocolPoolKeeper, app.AccountKeeper, app.BankKeeper),
)
```

Add an entry for SetOrderBeginBlockers, SetOrderEndBlockers, SetOrderInitGenesis, and SetOrderExportGenesis.

```go
app.ModuleManager.SetOrderBeginBlockers(
    // must come AFTER distribution.
    distrtypes.ModuleName,
    protocolpooltypes.ModuleName,
)
```

```go
app.ModuleManager.SetOrderEndBlockers(
    // order does not matter.
    protocolpooltypes.ModuleName,
)
```

```go
app.ModuleManager.SetOrderInitGenesis(
    // order does not matter.
    protocolpooltypes.ModuleName,   
)
```

```go
app.ModuleManager.SetOrderInitGenesis(
    protocolpooltypes.ModuleName, // must be exported before bank.
    banktypes.ModuleName,
)
```

### DI Wiring

First, set up the keeper for the application.

Import the protocolpool keeper:

```go
protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
```

Add the keeper to your application struct:

```go
ProtocolPoolKeeper protocolpoolkeeper.Keeper
```

Add the keeper to the depinject system:

```go
depinject.Inject(
    appConfig,
    &appBuilder,
    &app.appCodec,
    &app.legacyAmino,
    &app.txConfig,
    &app.interfaceRegistry,
    // ... other modules
    &app.ProtocolPoolKeeper, // NEW MODULE!
)
```

Next, set up configuration for the module.

Import the following:

```go
import (
    protocolpoolmodulev1 "cosmossdk.io/api/cosmos/protocolpool/module/v1"
    
    _ "github.com/cosmos/cosmos-sdk/x/protocolpool" // import for side-effects
    protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)
```

The protocolpool module has module accounts that handle funds. Add them to the module account permission configuration:

```go
moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
    // ...
    {Account: protocolpooltypes.ModuleName},
    {Account: protocolpooltypes.ProtocolPoolEscrowAccount},
}
```

Next, add an entry for BeginBlockers, EndBlockers, InitGenesis, and ExportGenesis.

```go
BeginBlockers: []string{
    // ...
    // must be AFTER distribution.
    distrtypes.ModuleName,
    protocolpooltypes.ModuleName,
},
```

```go
EndBlockers: []string{
    // ...
    // order for protocolpool does not matter.
    protocolpooltypes.ModuleName,
},
```

```go
InitGenesis: []string{
    // ... must be AFTER distribution.
    distrtypes.ModuleName,
    protocolpooltypes.ModuleName,
},
```

```go
ExportGenesis: []string{
    // ...
    // Must be exported before x/bank.
    protocolpooltypes.ModuleName, 
    banktypes.ModuleName,
},
```

Lastly, add an entry for protocolpool in the ModuleConfig.

```go
{
    Name:   protocolpooltypes.ModuleName,
    Config: appconfig.WrapAny(&protocolpoolmodulev1.Module{}),
},
```

## Adding Epochs Module **OPTIONAL**

### Manual Wiring

Import the following:

```go
import (
    // ...
    "github.com/cosmos/cosmos-sdk/x/epochs"
    epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
    epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
)
```

Add the epochs keeper to your application struct:

```go
EpochsKeeper epochskeeper.Keeper
```

Add the store key:

```go
keys := storetypes.NewKVStoreKeys(
    // ...
    epochstypes.StoreKey,
)
```

Instantiate the keeper:

```go
app.EpochsKeeper = epochskeeper.NewKeeper(
    runtime.NewKVStoreService(keys[epochstypes.StoreKey]),
    appCodec,
)
```

Set up hooks for the epochs keeper:

To learn how to write hooks for the epoch keeper, see the [x/epoch README](https://github.com/cosmos/cosmos-sdk/blob/main/x/epochs/README.md)

```go
app.EpochsKeeper.SetHooks(
    epochstypes.NewMultiEpochHooks(
        // insert epoch hooks receivers here
        app.SomeOtherModule
    ),
)
```

Add the epochs module to the module manager:

```go
app.ModuleManager = module.NewManager(
    // ...
    epochs.NewAppModule(appCodec, app.EpochsKeeper),
)
```

Add entries for SetOrderBeginBlockers and SetOrderInitGenesis:

```go
app.ModuleManager.SetOrderBeginBlockers(
    // ...
    epochstypes.ModuleName,
)
```

```go
app.ModuleManager.SetOrderInitGenesis(
    // ...
    epochstypes.ModuleName,
)
```

### DI Wiring

First, set up the keeper for the application.

Import the epochs keeper:

```go
epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
```

Add the keeper to your application struct:

```go
EpochsKeeper epochskeeper.Keeper
```

Add the keeper to the depinject system:

```go
depinject.Inject(
    appConfig,
    &appBuilder,
    &app.appCodec,
    &app.legacyAmino,
    &app.txConfig,
    &app.interfaceRegistry,
    // ... other modules
    &app.EpochsKeeper, // NEW MODULE!
)
```

Next, set up configuration for the module.

Import the following:

```go
import (
    epochsmodulev1 "cosmossdk.io/api/cosmos/epochs/module/v1"
    
    _ "github.com/cosmos/cosmos-sdk/x/epochs" // import for side-effects
    epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
)
```

Add an entry for BeginBlockers and InitGenesis:

```go
BeginBlockers: []string{
    // ...
    epochstypes.ModuleName,
},
```

```go
InitGenesis: []string{
    // ...
    epochstypes.ModuleName,
},
```

Lastly, add an entry for epochs in the ModuleConfig:

```go
{
    Name:   epochstypes.ModuleName,
    Config: appconfig.WrapAny(&epochsmodulev1.Module{}),
},
```

## Enable Unordered Transactions **OPTIONAL**

To enable unordered transaction support on an application, the ante handler options must be updated.

```go
options := ante.HandlerOptions{
    // ...
    UnorderedNonceManager: app.AccountKeeper,
    // The following options are set by default.
    // If you do not want to change these, you may remove the UnorderedTxOptions field entirely.
    UnorderedTxOptions: []ante.UnorderedTxDecoratorOptions{
        ante.WithUnorderedTxGasCost(2240),
        ante.WithTimeoutDuration(10 * time.Minute),
    },
}

anteDecorators := []sdk.AnteDecorator{
    ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
    circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
    ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
    ante.NewValidateBasicDecorator(),
    ante.NewTxTimeoutHeightDecorator(),
    ante.NewValidateMemoDecorator(options.AccountKeeper),
    ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
    ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
    ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
    ante.NewValidateSigCountDecorator(options.AccountKeeper),
    ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
    ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
    ante.NewIncrementSequenceDecorator(options.AccountKeeper),
    // NEW !! NEW !! NEW !!
    ante.NewUnorderedTxDecorator(options.UnorderedNonceManager, options.UnorderedTxOptions...)
}
```

## Upgrade Handler

The upgrade handler only requires adding the store upgrades for the modules added above.
If your application is not adding `x/protocolpool` or `x/epochs`, you do not need to add the store upgrade.

```go
// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v050 to v053.
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v0.53.x.
const UpgradeName = "v050-to-v053"

func (app SimApp) RegisterUpgradeHandlers() {
    app.UpgradeKeeper.SetUpgradeHandler(
        UpgradeName,
        func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
            return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
        },
    )

    upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
    if err != nil {
        panic(err)
    }

    if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
        storeUpgrades := storetypes.StoreUpgrades{
            Added: []string{
                epochstypes.ModuleName, // if not adding x/epochs to your chain, remove this line.
                protocolpooltypes.ModuleName, // if not adding x/protocolpool to your chain, remove this line.
            },
        }

        // configure store loader that checks if version == upgradeHeight and applies store upgrades
        app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
    }
}
```