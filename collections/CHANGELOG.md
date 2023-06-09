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

## [v0.2.0](https://github.com/cosmos/cosmos-sdk/releases/tag/collections%2Fv0.2.0)

### Features

* [#16074](https://github.com/cosmos/cosmos-sdk/pull/16074) – Makes the generic Collection interface public, still highly unstable.

### API Breaking

* [#16127](https://github.com/cosmos/cosmos-sdk/pull/16127) – In the `Walk` method the call back function being passed is allowed to error.

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/collections%2Fv0.1.0)

Collections `v0.1.0` is released! Check out the [docs](https://docs.cosmos.network/main/packages/collections) to know how to use the APIs.
