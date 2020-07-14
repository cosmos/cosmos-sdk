# Cosmos-SDK v0.39.0 Release Notes

This is the inaugural release of the **Cosmos SDK 0.39 «Launchpad»** release series.

See the [Cosmos SDK 0.39.0 milestone](https://github.com/cosmos/cosmos-sdk/milestone/27?closed=1) on our issue tracker for details.

## Changes to IAVL and store pruning

The pruning features introduced in the `0.38` release series are buggy and might lead to data loss,
even after upgrading to `v0.39.0`. When upgrading from `0.38` it is important to follow the instructions
below, to prevent data loss and database corruption.

**Note: there are are several breaking changes with regard to IAVL, stores, and pruning settings that affect command line clients, server configuration, and Golang API.**

### Migrate an application from 0.38.5 to 0.39.0

The IAVL's `v0.13.0` release introduced a pruning functionality that turned out to be buggy and flawed.
IAVL's new `v0.14.0` release now commits and flushes every state to disk as it did in pre-`v0.13.0` release.
The SDK's multi-store will track and ensure the proper heights are pruned. The operator can now set the pruning
options by passing a `pruning` configuration via command line option or `app.toml`. The `pruning` flag supports the following
options: `default`, `everything`, `nothing`, `custom` - see docs for further details. If the operator chooses `custom`, they
may want to provide either of the granular pruning values:
- `pruning-keep-recent`
- `pruning-keep-every`
- `pruning-interval`

The former two options dictate how many recent versions are kept on disk and the offset of what versions are kept after that
respectively, and the latter defines the height interval in which versions are deleted in a batch. **Note: there are are some
client application breaking changes with regard to IAVL, stores, and pruning settings.**

### Migrate a node from 0.38.5 to 0.39.0

Note: **do not modify pruning settings with any release prior to `v0.39.0` as that may cause data corruption**.
The following instructions assume that **pruning settings have not been modified since the node started using 0.38.x. Note: the default pruning setting `syncable` used `KeepEvery:100`.

* The simple upgrade strategy: **perform a full sync of the node from scratch**. Else, follow one of the other strategies.

* If your node had started with using `KeepEvery:1` (e.g. pruning settings `nothing` or `everything`), upgrading to `v0.39.0` is safe.

* Otherwise, do halt block processing with `--halt-height` after committing a height divisible by `KeepEvery` - e.g. at block 147600 with `KeepEvery:100`. The **node must never have processed a height beyond that at any time in its past**. Upgrading to `v0.39.0` is then safe.

* Otherwise, set the `KeepEvery` setting to the same as the previous `KeepEvery` setting (both `<=v0.38.5` and `v0.39.0` default to `KeepEvery:100`). Upgrade to `v0.39.0` is then safe as long as you wait one `KeepEvery` interval plus one `KeepRecent` interval **plus** one pruning `Interval` before changing pruning settings or deleting the last `<=v0.38.5` height (so wait *210* heights with the default configuration).

* Otherwise, make sure the last version persisted with `<=v0.38.5` is never deleted after upgrading to `v0.39.0`, as doing so may cause data loss and data corruption.

## Changes to ABCI Query's "app/simulate" path

The `app/simulate` query path is used to simulate the execution transactions in order to obtain an estimate
of the gas consumption that would be required to actually execute them. The response used to return only
the amount of gas, it now returns the result of the transaction as well.

## bank.send event comes with sender information

The `bank.send` event used to carry only the recipient and amount. It was assumed that the sender of the funds was `message.sender`.
This is often not true when a module call the bank keeper directly. This may be due to staking distribution, or via a cosmwasm contract that released funds (where I discovered the issue).

`bank.send` now contains the entire triple `(sender, recipient, amount)`.

## trace option is no longer ignored

The `--trace` option is reintroduced. It comes in very handy for debugging as it causes the full stack trace to be included in the ABCI error logs.

## appcli keys parse command didn't honor client application's bech32 prefixes

The `key parse` command ignored the application-specific address bech32
prefixes and used to return `cosmos*1`-prefixed addresses regardless
of the client application's configuration.
