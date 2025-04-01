# Upgrading Cosmos SDK [v0.53.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.0)

This guide provides instructions for upgrading from `v0.50.x` to `v0.53.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

🚨Upgrading to v0.53.x will require a **coordinated** chain upgrade.🚨

### TLDR;

Unordered transactions, `x/protocolpool`, and `x/epoch` are the major new features added in v0.53.x.

We also added the ability to add a `CheckTx` handler and enabled ed25519 signature verification.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/CHANGELOG.md).

### Unordered Transactions

The Cosmos SDK now supports unordered transactions. _This is an opt-in feature_.

Clients that use this feature may now submit their transactions in a fire-and-forget manner to chains that enabled unordered transactions.

To submit an unordered transaction, clients must set the `unordered` flag to
`true` and ensure a reasonable `timeout_timestamp` is set. The `timeout_timestamp` is
used as a TTL for the transaction and provides replay protection. Each transaction's `timeout_timestamp` must be
unique to the account; however, the difference may be as small as a nanosecond. See [ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md) for more details.

#### Enabling Unordered Transactions

To enable unordered transactions, set the new `UnorderedNonceManager` field in the `x/auth` `ante.HandlerOptions`.

```go
ante.HandlerOptions{
    UnorderedNonceManager: app.AccountKeeper, // NEW
}
```

By default, unordered transactions use a transaction timeout duration of 10 minutes and a default gas charge of 2240 gas.
To modify these default values, pass in the corresponding options to the new `UnorderedTxOptions` field in `x/auth's` `ante.HandlerOptions`.

```go
ante.HandlerOptions{
    UnorderedNonceManager: app.AccountKeeper,
    UnorderedTxOptions: []ante.UnorderedTxDecoratorOptions{
        ante.WithTimeoutDuration(XXXX * time.Minute),
        ante.WithUnorderedTxGasCost(XXXX),
    },
}	
```

### App Wiring Changes

In this section, we describe the required app wiring changes to run a v0.53.x Cosmos SDK application.

**These changes are directly applicable to your application wiring.**

The `x/auth` module now contains a `PreBlocker` that _must_ be set in the module manager's `SetOrderPreBlockers` method.

```go
app.ModuleManager.SetOrderPreBlockers(
    upgradetypes.ModuleName,
    authtypes.ModuleName, // NEW
)
```

That's it.

### New Modules

Below are some **optional** new modules you can include in your chain. 
To see a full example of wiring these modules, please check out the [SimApp](https://github.com/cosmos/cosmos-sdk/blob/release/v0.53.x/simapp/app.go).

#### Epochs

⚠️Adding this module requires a `StoreUpgrade`⚠️

The new, supplemental `x/epochs` module provides Cosmos SDK modules functionality to register and execute custom logic at fixed time-intervals.

Required wiring:
- Keeper Instantiation
- StoreKey addition
- Hooks Registration 
- App Module Registration
- entry in SetOrderBeginBlockers
- entry in SetGenesisModuleOrder
- entry in SetExportModuleOrder

#### ProtocolPool

⚠️Adding this module requires a `StoreUpgrade`⚠️

The new, supplemental `x/protocolpool` module provides extended functionality for managing and distributing block reward revenue.

Required wiring:
- Module Account Permissions
  - protocolpooltypes.ModuleName (nil)
  - protocolpooltypes.ProtocolPoolEscrowAccount (nil)
- Keeper Instantiation
- StoreKey addition
- Passing the keeper to the Distribution Keeper
  - `distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper)`
- App Module Registration
- entry in SetOrderBeginBlockers
- entry in SetOrderEndBlockers
- entry in SetGenesisModuleOrder
- entry in SetExportModuleOrder **before `x/bank`**

### Misc Changes

#### Testnet's init-files Command

Some changes were made to `testnet`'s `init-files` command to support our new testing framework, `Systemtest`.

##### Flag Changes

- The flag for validator count was changed from `--v` to `--validator-count`(shorthand: `-v`).

##### Flag Additions
- `--staking-denom` allows changing the default stake denom, `stake`.
- `--commit-timeout` enables changing the commit timeout of the chain.
- `--single-host` enables running a multi-node network on a single host. This bumps each subsequent node's network addresses by 1. For example, node1's gRPC address will be 9090, node2's 9091, etc...