# Cosmos SDK v0.46.3 Release Notes

This release introduces a number of bug fixes, features and improvements.

Highlights:

+ `ApplicationQueryService` was introduced to enable additional query service registration. Applications should implement `RegisterNodeService(client.Context)` method to automatically expose chain information query service implemented in [#13485](https://github.com/cosmos/cosmos-sdk/pull/13485). 

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.2...v0.46.3
