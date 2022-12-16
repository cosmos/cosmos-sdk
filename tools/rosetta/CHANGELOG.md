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

### Improvements

* [#14272](https://github.com/cosmos/cosmos-sdk/pull/14272) Use `coinbase/rosetta-sdk-go/types` packages instead of comsos fork.

### Bug Fixes

* [#14285](https://github.com/cosmos/cosmos-sdk/pull/14285) Sets tendermint errors status codes to 500

## v0.2.0 2022-12-07

### Improvements

* [#14118](https://github.com/cosmos/cosmos-sdk/pull/14118) Allow rosetta to be installed as a standalone application.
* [#14061](https://github.com/cosmos/cosmos-sdk/pull/14061) Adds openapi specification.
* [#13832](https://github.com/cosmos/cosmos-sdk/pull/13832) Correctly populates rosetta's `/network/status` endpoint response. Rosetta's data api is divided into its own go files (account, block, mempool, network).

### Bug Fixes

* [#13832](https://github.com/cosmos/cosmos-sdk/pull/13832) Wrap tendermint RPC errors to rosetta errors.

## v0.1.0 2022-11-04

**From `v0.1.0` the minimum version of Tendermint is `v0.37+`, due event type changes.**

### Improvements

* [#13583](https://github.com/cosmos/cosmos-sdk/pull/13583) Extract rosetta to its own go.mod.
