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

### Features

* [#17569](https://github.com/cosmos/cosmos-sdk/pull/17569) Introduce a new message type, `MsgBurn`, to burn coins.

### Improvements

* [#18636](https://github.com/cosmos/cosmos-sdk/pull/18636) `SendCoinsFromModuleToAccount`, `SendCoinsFromModuleToModule`, `SendCoinsFromAccountToModule`, `DelegateCoinsFromAccountToModule`, `UndelegateCoinsFromModuleToAccount`, `MintCoins` and `BurnCoins` methods now returns an error instead of panicking if any module accounts does not exist or unauthorized.

### API Breaking Changes

* [#17569](https://github.com/cosmos/cosmos-sdk/pull/17569) `BurnCoins` takes an address instead of a module name
* [#19477](https://github.com/cosmos/cosmos-sdk/pull/19477) `appmodule.Environment` is passed to bank `NewKeeper`
* [#19627](https://github.com/cosmos/cosmos-sdk/pull/19627) The genesis api has been updated to match `appmodule.HasGenesis`.
* [#19740](https://github.com/cosmos/cosmos-sdk/pull/19740) Verify `InitGenesis` and `ExportGenesis` module code and keeper code do not panic.

### Consensus Breaking Changes

* [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188) Remove creation of `BaseAccount` when sending a message to an account that does not exist
