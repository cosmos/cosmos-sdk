# Cosmos SDK v0.44.0 Release Notes

v0.44 is a security release which contains a possible consensus breaking change.
It doesn't bring any new feature and it's a logical continuation of the v0.43.

Consequences:
+ v0.43 is discontinued;
+ all chains should upgrade to v0.44. Update from v0.43 doesn't require any migration. Chains can upgrade directly from v0.42, in that case v0.43 migrations will be executed as a part of v0.44;
+ all planned features for v0.44 are going to land in v0.45, with the same release schedule.

NOTE: v0.42 release will get to the end of life on September 8, 2021.

Please see [Cosmos SDK v0.43.0 Release Notes](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/RELEASE_NOTES.md).

## Updates

For a comprehsive list of all breaking changes and improvements since the v0.42 "Stargate" release series, please see the **[CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/CHANGELOG.md)**.

### API Breaking Changes
* (client/tx) `BuildUnsignedTx`, `BuildSimTx`, `PrintUnsignedStdTx` functions are moved to the Tx Factory as methods.

### Client Breaking Changes

* Remove broadcast & encode legacy REST endpoints. Both requests should use the new gRPC-Gateway REST endpoints. Please see the [REST Endpoints Migration guide](https://docs.cosmos.network/master/migrations/rest.html) to migrate to the new REST endpoints.
