# Cosmos SDK v0.47.11 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/6)

## 🚀 Highlights

For this month patch release of the v0.47.x line, a few bug were fixed in the SDK.
Notably:

* `secp256r1` keys now implement gogoproto's customtype interface.
* CLI now throws an error when signing with an incorrect Ledger.
* Fixing [GHSA-4j93-fm92-rp4m](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-4j93-fm92-rp4m) in `x/feegrant` and `x/authz` modules. The upgrade instuctions were provided in the [v0.47.9 release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.9).

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.47.11/CHANGELOG.md) for an exhaustive list of changes or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.47.10...v0.47.11) from last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.x`.

## Maintenance Policy

v0.50 has been released which means the v0.47.x line is now supported for bug fixes only, as per our release policy. Earlier versions are not maintained.  

Start integrating with [Cosmos SDK Eden (v0.50)](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.6) and enjoy and the new features and performance improvements.
