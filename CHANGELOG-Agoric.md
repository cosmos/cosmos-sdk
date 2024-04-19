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

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking Protobuf, gRPC and REST routes used by end-users.
"CLI Breaking" for breaking CLI commands.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog (Agoric fork)

## Unreleased

### Improvements

* (auth) [#407](https://github.com/agoric-labs/cosmos-sdk/pull/407) Configurable fee collector module account in DeductFeeDecorator.

### API Breaking

* (auth, bank) Agoric/agoric-sdk#8989 Remove deprecated lien support

## `v0.46.16-alpha.agoric.2.4` - 2024-04-19

### Improvements

* (server) [#419](https://github.com/agoric-labs/cosmos-sdk/pull/419) More unit tests for flag vs toml precedence for ABCI client type.

## `v0.46.16-alpha.agoric.2.3` - 2024-04-17

### Improvements

* (baseapp) [#415](https://github.com/agoric-labs/cosmos-sdk/pull/415) Unit tests and documentation for event history.
* (server) [#416](https://github.com/agoric-labs/cosmos-sdk/pull/416) Config entry to select ABCI client type.

### Bug Fixes

* (baseapp) [#415](https://github.com/agoric-labs/cosmos-sdk/pull/415) Avoid duplicates in event history.

## `v0.46.16-alpha.agoric.2.2` - 2024-04-12

### Improvements

* (server) [#409](https://github.com/agoric-labs/cosmos-sdk/pull/409) Flag to select ABCI client type.
* (deps) [#412](https://github.com/agoric-labs/cosmos-sdk/pull/412) Bump iavl to v0.19.7

### Bug Fixes

* (baseapp) [#413](https://github.com/agoric-labs/cosmos-sdk/pull/413) Ignore events from simulated transactions

## `v0.46.16-alpha.agoric.2.1` - 2024-03-08

### Improvements

* (auth) #??? Configurable fee collector module account in DeductFeeDecorator. (backport #407)

## `v0.46.16-alpha.agoric.2` - 2024-02-08

* Agoric/agoric-sdk#8871 Have `tx gov submit-proposal` accept either new or legacy syntax

### Bug Fixes

* (crypto) [#19371](https://github.com/cosmos/cosmos-sdk/pull/19371) Avoid cli redundant log in stdout, log to stderr instead.

## `v0.46.16-alpha.agoric.1` - 2024-02-05

* Agoric/agoric-sdk#8224 Merge [cosmos/cosmos-sdk v0.46.16](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.16)

### Bug Fixes

* Agoric/agoric-sdk#8719 MsgClawback returns the wrong Type

## `v0.45.16-alpha.agoric.3` - 2023-12-04

* (vesting) [#342](https://github.com/agoric-labs/cosmos-sdk/pull/342) recipient can return clawback vesting grant to funder

## `v0.45.16-alpha.agoric.2` - 2023-11-08

### Bug Fixes

* (baseapp) [#337](https://github.com/agoric-labs/cosmos-sdk/pull/337) revert #305 which causes test failures in agoric-sdk

## `v0.45.16-alpha.agoric.1` - 2023-09-22

### Improvements

* Agoric/agoric-sdk#8223 Merge [cosmos/cosmos-sdk v0.45.16](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16)
* (vesting) [#303](https://github.com/agoric-labs/cosmos-sdk/pull/303) Improve vestcalc comments and documentation.

### Bug Fixes

* (snapshots) [#304](https://github.com/agoric-labs/cosmos-sdk/pull/304) raise the per snapshot item limit. Fixes [Agoric/agoric-sdk#8325](https://github.com/Agoric/agoric-sdk/issues/8325)
* (baseapp) [#305](https://github.com/agoric-labs/cosmos-sdk/pull/305) Make sure we don't execute blocks beyond the halt height. Port of [cosmos/cosmos-sdk#16639](https://github.com/cosmos/cosmos-sdk/pull/16639)
