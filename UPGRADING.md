# Upgrading Cosmos SDK [v0.53.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.0)

This guide provides instructions for upgrading from `v0.50.x` to `v0.53.x` of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

🚨Upgrading to v0.53.x will require a **coordinated** chain upgrade.

### TLDR;

Unordered transactions, x/protocolpool, and x/epoch are the major new features added in v0.53.x.

We also added the ability to add a checkTx handler, enabled ed25519 transaction signature verification, and various bug fixes and DevX changes.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/CHANGELOG.md).

### Unordered Transactions

The Cosmos SDK now supports unordered transactions. Clients that use this feature may now submit their transactions
in a fire-and-forget manner.

To submit an unordered transaction, clients must set the `unordered` flag to
`true` and ensure a reasonable `timeout_timestamp` is set. The `timeout_timestamp` is
used as a TTL for the transaction and provides replay protection. Each transaction's `timeout_timestamp` must be
unique to the account; however, the difference may be as small as a nanosecond. See [ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md) for more details.

#### Enable Unordered Transactions

To enable unordered transactions, set the new `UnorderedNonceManager` field in the `x/auth` `ante.HandlerOptions`.

```go
ante.HandlerOptions{
    UnorderedNonceManager: app.AccountKeeper, // NEW
    // other options...
}
```

By default, unordered transaction support uses a transaction timeout duration of 10 minutes, and charges 2240 gas.
To modify these default values, pass in the corresponding options to the new `UnorderedTxOptions` field in `x/auth's` `ante.HandlerOptions`.

```go
ante.HandlerOptions{
    UnorderedNonceManager: app.AccountKeeper,
    UnorderedTxOptions: []ante.UnorderedTxDecoratorOptions{
        ante.WithTimeoutDuration(10 * time.Minute),
        ante.WithUnorderedTxGasCost(2240),
    },
	// ... other options
}	
```

### SimApp

In this section we describe the changes made in Cosmos SDK's SimApp.
**These changes are directly applicable to your application wiring.**

The `x/auth` module now contains a PreBlocker that must be set in the module manager's `SetOrderPreBlockers` method.

```go
app.ModuleManager.SetOrderPreBlockers(
    upgradetypes.ModuleName,
    authtypes.ModuleName, // NEW
)
```

That's it.

### New Modules

Below are some **optional** new modules you can include in your chain. 
To see a full example of wiring these modules, please see our [SimApp](https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/simapp/app.go).

#### Epochs

⚠️This module requires a `StoreUpgrade`⚠️

The new, supplemental `x/epochs` module provides Cosmos SDK modules functionality to register and execute custom logic at fixed time-intervals.


#### ProtocolPool

⚠️This module requires a `StoreUpgrade`⚠️

The new, supplemental `x/protocolpool` module provides extended functionality for managing and distributing block reward revenue.

