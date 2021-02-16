# Cosmos SDK v0.41.1 "Stargate" Release Notes

This release includes two security patches, and does not introduce any breaking changes. It is **highly recommended** that all applications using v0.41.0 upgrade to v0.41.1 as soon as possible.

See the [Cosmos SDK v0.41.1 milestone](https://github.com/cosmos/cosmos-sdk/milestone/38?closed=1) on our issue tracker for details.

### Security Patch #1: All gRPC requests are now routed through ABCI

When Tendermint commits a new block, the `versions` map in IAVL MutableTree is updated. If, at the same time, a concurrent gRPC request is performed, it will read the same `versions` map, causing the node to crash.

The patch consists of routing all gRPC requests through ABCI. The Go implementation of ABCI uses global lock on all requests, making them linearizable (received one at a time) which in turn prevents concurrent map reads and writes.

We are exploring on ways of introducing concurrent gRPC queries in [#8591](https://github.com/cosmos/cosmos-sdk/issues/8591).

This bug has been reported via HackerOne.

### Security Patch #2: Remove `GetValidator` cache map

The `x/staking` keeper holds an internal `validatorCache` cache map of validators. When multipile gRPC queries are performed simulataneously, concurrent reads and writes of this map can happen, causing the node to crash.

The patch removes the `validatorCache` altogether. Benchmarks show that the removal of the cache map even increases performance.

Is is important to note that the Security Patch #1 should also fix this bug, as it forces synchronous gRPC queries and therefore synchronous map reads/writes. However, it was deemed useful to include this bugfix too in this release.

### Bug Fixes

Several bug fixes are included in this release.

Tendermint has been bumped to v0.34.4 to address a memory leak.

Environment variables are not correctly populated to CLI flags. When using the Tendermint subcommands `tendermint show-*` from the CLI, the SDK doesn't create new files anymore.

Keyring imports from older versions are fixed.

Additional validation for client denom metadata has been added.

On the IBC side, a `packet_connection` attribute has been added to IBC events to enable relayer filtering.
