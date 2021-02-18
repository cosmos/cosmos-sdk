# Cosmos SDK v0.41.2 "Stargate" Release Notes

This release upgrades Tendermint to v0.34.7, and does not introduce any breaking changes. It is **highly recommended** that all applications using v0.41.1 upgrade to v0.41.2 as soon as possible.

### Tendermint v0.34.7

Operators running nodes that manage their keys through the Tendermint's `FilePV` implementation were
susceptible to leaking private keys material in the logs. The issue is now fixed in Tendermint v0.34.5 and later versions.

For more information regarding the patch, please refer to [Tendermint's changelog](https://github.com/tendermint/tendermint/blob/v0.34.7/CHANGELOG.md#v0345).
