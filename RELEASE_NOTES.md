# Cosmos SDK v0.50.0-rc.1 Release Notes

There are no release notes for pre-releases.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/CHANGELOG.md) for an exhaustive list of changes.  
Refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) for upgrading your application.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/release/v0.47.x...release/v0.50.x

## Upgrading from v0.50.0-rc.z-0

If you have started integrating with v0.50.0-rc.0, this release candidate contains one breaking change.
This is contrary to our usual policy of not introducing breaking changes in release candidates, but we believe this change is necessary to ensure a smooth upgrade from previous SDK version to v0.50.0. Additionally, it gives a better UX for users integrating vote extensions. Read more about the change [here](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-068-preblock.md).

Update your app.go / app_config.go as instructed in the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md#set-preblocker).
