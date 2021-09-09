# Cosmos SDK v0.44.0 Release Notes

v0.44 is a security release which contains a consensus breaking change.
It doesn't bring any new feature and it's a logical continuation of v0.43.

Consequences:
+ v0.43 is discontinued;
+ all chains should upgrade to v0.44. Update from v0.43 doesn't require any migration. Chains can upgrade directly from v0.42, in that case v0.43 migrations must be executed when upgrading to v0.44;
+ all previously planned features for v0.44 are going to land in v0.45, with the same release schedule.

Please see [Cosmos SDK v0.43.0 Release Notes](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/RELEASE_NOTES.md).

## Updates

For a comprehensive list of all breaking changes and improvements since the v0.42 "Stargate" release series, please see the **[CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/CHANGELOG.md)**.

### Client Breaking Changes

* Remove broadcast & encode legacy REST endpoints. Both requests should use the new gRPC-Gateway REST endpoints. Please see the [REST Endpoints Migration guide](https://docs.cosmos.network/master/migrations/rest.html) to migrate to the new REST endpoints.
