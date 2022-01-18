# Cosmos SDK v0.45.0 Release Notes

Cosmos SDK v0.45.0 is a logical continuation of the v0.44.\* series, but brings a couple of state- and API-breaking changes requested by the community.

### State-Breaking Changes

There are few important changes in **gas consumption**, which improve the gas economics:

- We now charge gas in two new places: on `.Seek()` even if there are no entries, and for the key length (on top of the value length).
- When block gas limit is exceeded, we consume the maximum gas possible (to charge for the performed computation). We also fixed the bug when the last transaction in a block exceeds the block gas limit, it returns an error result, but the tx is actually committed successfully.

Finally, a small improvement in gov, we increased the maximum proposal description size from 5k characters to 10k characters.

### API-Breaking Changes

- The `BankKeeper` interface has a new `HasSupply` method to ensure that input denom actually exists on chain.
- The `CommitMultiStore` interface contains a new `SetIAVLCacheSize` method for a configurable IAVL cache size.
- `AuthKeeper` interface in `x/auth` now includes a function `HasAccount`.
- Moved `TestMnemonic` from `testutil` package to `testdata`.


Finally, when using the `SetOrder*` functions in simapp, e.g. `SetOrderBeginBlocker`, we now require that all modules be present in the function arguments, or else the node panics at startup. We also added a new `SetOrderMigration` function to set the order of running module migrations.

### Improvements

- Speedup improvements (e.g. speedup iterator creation after delete heavy workloads, lower allocations for `Coins.String()`, reduce RAM/CPU usage inside store/cachekv's `Store.Write`) are included in this release.
- Upgrade Rosetta to v0.7.0 .
- Support in-place migration ordering.
- Copied and updated `server.GenerateCoinKey` and `server.GenerateServerCoinKey` functions to the `testutil` package. These functions in `server` package are marked deprecated and will be removed in the next release. In the `testutil.GenerateServerCoinKey` version we  added support for custom mnemonics in in-process testing network.

See our [CHANGELOG](./CHANGELOG.md) for the exhaustive list of all changes, or a full [commit diff](https://github.com/cosmos/cosmos-sdk/compare/v0.44.5...v0.45.0).
