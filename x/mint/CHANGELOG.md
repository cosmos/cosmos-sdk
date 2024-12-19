<!--
Guiding Principles:
Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.
Usage:
Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:
* (<tag>) [#<issue-number>] Changelog message.
Types of changes (Stanzas):
"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"API Breaking" for breaking exported APIs used by developers building on SDK.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

## [v0.2.0-rc.1](https://github.com/cosmos/cosmos-sdk/releases/tag/x/mint/v0.2.0-rc.1) - 2024-12-18

### Features

* [#20363](https://github.com/cosmos/cosmos-sdk/pull/20363) Implemented epoched minting, configurable through `MintFn`. Now `MintFn` doesn't do any assumptions on how tokens are minted, users can define their own minting logic. 
* [#19896](https://github.com/cosmos/cosmos-sdk/pull/19896) Added a new max supply genesis param to existing params.

### API Breaking Changes

* [#20363](https://github.com/cosmos/cosmos-sdk/pull/20363) Deprecated InflationCalculationFn in favor of MintFn, `keeper.DefaultMintFn` wrapper must be used in order to continue using it in `NewAppModule`. This is not breaking for depinject users, as both `MintFn` and `InflationCalculationFn` are accepted.
* [#19367](https://github.com/cosmos/cosmos-sdk/pull/19398) `appmodule.Environment` is received on the Keeper to get access to different application services.
* [#21858](https://github.com/cosmos/cosmos-sdk/pull/21858) `NewKeeper` now returns a pointer to `Keeper`.
* [#21858](https://github.com/cosmos/cosmos-sdk/pull/21858), [#22979](https://github.com/cosmos/cosmos-sdk/pull/22979) `DefaultMintFn` now takes `StakingKeeper` and `MintKeeper` as arguments to avoid staking keeper being required by mint.   
    * `SetMintFn` is used to replace the default minting function.
