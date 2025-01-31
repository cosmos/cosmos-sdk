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

### Improvements
* [#23568](https://github.com/cosmos/cosmos-sdk/pull/23568) Remove auto migration and fix restore cmd
* [#23013](https://github.com/cosmos/cosmos-sdk/pull/23013) Support memDB for sims

### Bug Fixes

* [#23552](https://github.com/cosmos/cosmos-sdk/pull/23552) Fix pebbleDB integration

## [v2.0.0-beta.2](https://github.com/cosmos/cosmos-sdk/releases/tag/store/v2.0.0-beta.2)

* [#22336](https://github.com/cosmos/cosmos-sdk/pull/22336) Finish migration manager.
* [#23157](https://github.com/cosmos/cosmos-sdk/pull/23157) Remove support for RocksDB.

## [v2.0.0-beta.1](https://github.com/cosmos/cosmos-sdk/releases/tag/store/v2.0.0-beta.1)

Initial tag of `cosmossdk.io/store/v2`.
