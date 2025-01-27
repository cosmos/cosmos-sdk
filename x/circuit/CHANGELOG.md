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

## [v0.2.0-rc.1](https://github.com/cosmos/cosmos-sdk/releases/tag/x/circuit/v0.2.0-rc.1) - 2024-12-18

### API Breaking Changes

* [#19041](https://github.com/cosmos/cosmos-sdk/pull/19041) `appmodule.Environment` is received on the Keeper to get access to different application services

## [v0.1.1](https://github.com/cosmos/cosmos-sdk/releases/tag/x/circuit/v0.1.1) - 2024-04-22

### Improvements

* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Upgrade SDK version due to Prometheus breaking change.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Bump `cosmossdk.io/store` to v1.1.0.
* (feat) [#22459](https://github.com/cosmos/cosmos-sdk/pull/22459) Allow msg Reset with empty msgURL
* (feat) [#22460](https://github.com/cosmos/cosmos-sdk/pull/22460) Add validation for permission when an account is assigned and validation for msgURL

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/x/circuit/v0.1.0) - 2023-11-07
