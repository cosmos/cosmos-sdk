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
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Feature

* [#15320](https://github.com/cosmos/cosmos-sdk/pull/15320) Add current sequence getter (`LastInsertedSequence`) for auto increment tables.

### API Breaking Changes

* [#14822](https://github.com/cosmos/cosmos-sdk/pull/14822) Migrate to cosmossdk.io/core genesis API

### State-machine Breaking Changes

* [#12273](https://github.com/cosmos/cosmos-sdk/pull/12273) The timestamp key encoding was reworked to properly handle nil values. Existing users will need to manually migrate their data to the new encoding before upgrading.
