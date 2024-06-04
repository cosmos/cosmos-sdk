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

* [#18355](https://github.com/cosmos/cosmos-sdk/pull/18355) Added new versions for `Balance`, `Owner`, `Supply`, `NFT`, `Class` queries that receives request via query string.
* [#19367](https://github.com/cosmos/cosmos-sdk/pull/19367) `appmodule.Environment` is received on the Keeper to get access to different application services

### API Breaking Changes

* [#19740](https://github.com/cosmos/cosmos-sdk/pull/19740) `InitGenesis` and `ExportGenesis` module code and keeper code do not panic but return errors.

## [v0.1.1](https://github.com/cosmos/cosmos-sdk/releases/tag/x/nft/v0.1.1) - 2024-04-22

### Improvements

* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Upgrade SDK version due to Prometheus breaking change.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Bump `cosmossdk.io/store` to v1.1.0.

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/x/nft/v0.1.0) - 2023-11-07


