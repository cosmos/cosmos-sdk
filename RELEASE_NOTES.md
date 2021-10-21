# Cosmos SDK v0.44.3 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.44 series.

The main performance improvement concerns gRPC queries, which are now able to run concurrently on the node ([\#10045](https://github.com/cosmos/cosmos-sdk/pull/10045)). To benefit from this performance boost, make sure to send your gRPC queries to the gRPC server directly (default port `9090`) instead of using the Tendermint RPC [`abci_query` endpoint](https://docs.tendermint.com/master/rpc/#/ABCI/abci_query) (default port `26657`).

This release notably also:

- bumps Tendermint to [v0.34.14](https://github.com/tendermint/tendermint/releases/tag/v0.34.14).
- bumps the `gin-gonic/gin` version to 1.7.0 to fix the upstream [security vulnerability](https://github.com/advisories/GHSA-h395-qcrw-5vmq).
- adds a null guard with a user-friendly error message for possible nil `Amount` in tx fee `Coins`.

See the [Cosmos SDK v0.44.3 milestone](https://github.com/cosmos/cosmos-sdk/blob/v0.44.3/CHANGELOG.md) on our issue tracker for the exhaustive list of all changes.
