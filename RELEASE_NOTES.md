# Cosmos SDK v0.42.10 "Stargate" Release Notes

This release includes a new `AnteHandler` that rejects redundant IBC transactions to save relayers fees. This is a backport of [ibc-go \#235](https://github.com/cosmos/ibc-go/pull/235).

v0.42.10 also includes multiple performance improvements, such as:

- improve ABCI performance under concurrent load via [Tendermint v0.34.13](https://github.com/tendermint/tendermint/blob/master/CHANGELOG.md#v03413),
- improve CacheKVStore datastructures / algorithms, to no longer take O(N^2) time when interleaving iterators and insertions.
- bump IAVL to v0.17.1 which includes performance improvements on a batch load.

We fixed the keyring to use pre-configured data for `add-genesis-account` command.

Finally, when using the `client.Context`, we modified ABCI queries to use `abci.QueryRequest`'s `Height` field if it is non-zero, otherwise continue using `client.Context`'s height. This is a minor client-breaking change for users of the `client.Context`.

See the [Cosmos SDK v0.42.10 milestone](https://github.com/cosmos/cosmos-sdk/milestone/55?closed=1) on our issue tracker for the exhaustive list of all changes.
