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

### API Breaking Changes

* [#19450](https://github.com/cosmos/cosmos-sdk/pull/19450) Migrate module to use `appmodule.Environment` instead of passing individual services.

### Consens Breaking Changes

* [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188) Remove creation of `BaseAccount` when sending a message to an account that does not exist

### Features

* [#14649](https://github.com/cosmos/cosmos-sdk/pull/14649) The `x/feegrant` module is extracted to have a separate go.mod file which allows it to be a standalone module.

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/x/feegrant/v0.1.0) - 2023-11-07

### Features

* [#18047](https://github.com/cosmos/cosmos-sdk/pull/18047) Added a limit of 200 grants pruned per EndBlock and the method PruneAllowances that prunes 75 expired grants on every run.

### Improvements

* [#18767](https://github.com/cosmos/cosmos-sdk/pull/18767) Ensure we only execute revokeAllowance if there is no error is the grant is to be removed.

### API Breaking Changes

* [#15606](https://github.com/cosmos/cosmos-sdk/pull/15606) `NewKeeper` now takes a `KVStoreService` instead of a `StoreKey` and methods in the `Keeper` now take a `context.Context` instead of a `sdk.Context`.
* [#15347](https://github.com/cosmos/cosmos-sdk/pull/15347) Remove global bech32 usage in keeper.
* [#15347](https://github.com/cosmos/cosmos-sdk/pull/15347) `ValidateBasic` is treated as a no op now with with acceptance of RFC001
* [#17869](https://github.com/cosmos/cosmos-sdk/pull/17869) `NewGrant`, `NewMsgGrantAllowance` & `NewMsgRevokeAllowance` takes strings instead of `sdk.AccAddress`
* [#16535](https://github.com/cosmos/cosmos-sdk/pull/16535) Use collections for `FeeAllowance`, `FeeAllowanceQueue`.
* [#18815](https://github.com/cosmos/cosmos-sdk/pull/18815) Add the implementation of the `UpdatePeriodReset` interface to update the value of the `PeriodReset` field.
