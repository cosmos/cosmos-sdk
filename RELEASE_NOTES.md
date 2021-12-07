# Cosmos SDK v0.42.11 "Stargate" Release Notes

This release includes a client-breaking change for HTTP users querying account balances by denom:

```diff
- <base_url>/cosmos/bank/v1beta1/balances/<address>/<denom>
+ <base_url>/cosmos/bank/v1beta1/balances/<address>/by_denom?denom=<denom>
```

This change was made to fix querying IBC denoms via HTTP.

v0.42.11 also includes multiple bugfixes and performance improvements, such as:

- Upgrade IAVL to 0.17.3 to solve race condition bug in IAVL.
- Bump Tendermint to [v0.34.14](https://github.com/tendermint/tendermint/releases/tag/v0.34.14).

Finally, when querying for transactions, we added an `Events` field to the `TxResponse` type that captures _all_ events emitted by a transaction, unlike `Logs` which only contains events emitted during message execution. `Logs` and `Events` may currently contain duplicate data, but `Logs` will be deprecated in a future version.

See our [CHANGELOG](./CHANGELOG.md) for the exhaustive list of all changes, or a full [commit diff](https://github.com/cosmos/cosmos-sdk/compare/v0.42.09...v0.42.10).
