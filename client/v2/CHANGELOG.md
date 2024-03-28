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

## [v2.0.0-beta.2] - 2024-XX-XX

### Improvements

* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Upgrade SDK version to due prometheus breaking change.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Bump `cosmossdk.io/store` to v1.1.0.
* [#19618](https://github.com/cosmos/cosmos-sdk/pull/19618) Marshal enum as string in queries.
* [#19060](https://github.com/cosmos/cosmos-sdk/pull/19060) Use client context from root (or enhanced) command in autocli commands.
    * Note, the given command must have a `client.Context` in its context.
* [#19216](https://github.com/cosmos/cosmos-sdk/pull/19216) Do not overwrite TxConfig, use directly the one provided in context. TxConfig should always be set in the `client.Context` in `root.go` of an app.

### Bug Fixes

* [#19377](https://github.com/cosmos/cosmos-sdk/pull/19377) Partly fix comment parsing in autocli.
* [#19060](https://github.com/cosmos/cosmos-sdk/pull/19060) Simplify key flag parsing logic in flag handler.

## [v2.0.0-beta.1] - 2023-11-07

This is the first tagged version of client/v2.
It depends on the Cosmos SDK v0.50 release and fully supports AutoCLI.
