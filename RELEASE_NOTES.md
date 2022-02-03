# Cosmos SDK v0.45.0 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.45 series:

Highlights
+ Added the missing `iavl-cache-size` config parameter parsing to set a desired IAVL cache size. The default value is way to small for big chains, and causes OOM failures.
+ Added a check in `x/upgrade` module's `BeginBlock` preventing accidental binary downgrades
+ Fix: the `/cosmos/tx/v1beta1/txs/{hash}` endpoint returns correct return code (404) for a non existing tx.

See the [Cosmos SDK v0.45.1  Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.45.1/CHANGELOG.md) for the exhaustive list of all changes and check other fixes in 0.45.x release series.

**Full Diff**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.0...v0.45.1


