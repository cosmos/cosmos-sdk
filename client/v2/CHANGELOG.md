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

## [v2.0.0-beta.9](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.9) - 2025-04-24

### Features

* [#24359](https://github.com/cosmos/cosmos-sdk/pull/24359) Support governance proposals.

### Improvements

* [#22890](https://github.com/cosmos/cosmos-sdk/pull/22890)  Added support for flattening inner message fields in autocli as positional arguments.

### Bug Fixes

* (cli) [#24330](https://github.com/cosmos/cosmos-sdk/pull/24330) Use the gogoproto merge registry as a file resolver instead of the interface registry.
* [#21853](https://github.com/cosmos/cosmos-sdk/pull/21853) Fix `*big.Int` unmarshalling in txs.

## [v2.0.0-beta.8](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.8) - 2025-01-29

* [#23544](https://github.com/cosmos/cosmos-sdk/pull/23544) Support map in queries for autocli.

## [v2.0.0-beta.7](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.7) - 2024-12-10

### Bug Fixes

* [#22817](https://github.com/cosmos/cosmos-sdk/pull/22817) Add DecCoin support in autocli flag builder.

## [v2.0.0-beta.6](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.6) - 2024-11-21

### Improvements

* [#22443](https://github.com/cosmos/cosmos-sdk/pull/22443) Add keyring flags in query commands.
* [#21936](https://github.com/cosmos/cosmos-sdk/pull/21936) Print possible enum values in error message after an invalid input was provided.

### Bug Fixes

* [#21809](https://github.com/cosmos/cosmos-sdk/pull/21809) Correctly handle enhanced sub commands.
* [#22576](https://github.com/cosmos/cosmos-sdk/pull/22576) Fix duplicate command addition in `autocli` when custom enhanced command has a different name than module name

## [v2.0.0-beta.5](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.5) - 2024-09-18

### Improvements

* [#21936](https://github.com/cosmos/cosmos-sdk/pull/21936) Print possible enum values in error message after an invalid input was provided.

## [v2.0.0-beta.4](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.4) - 2024-07-16

### Improvements

* [#21712](https://github.com/cosmos/cosmos-sdk/pull/21712) Marshal `type` field as proto message url in queries instead of amino name.

### Bug Fixes

* [#20964](https://github.com/cosmos/cosmos-sdk/pull/20964) Fix `GetNodeHomeDirectory` helper in `client/v2/helpers` to respect the `(PREFIX)_HOME` environment variable.

## [v2.0.0-beta.3](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.3) - 2024-07-15

### Features

* [#20771](https://github.com/cosmos/cosmos-sdk/pull/20771) Add `GetNodeHomeDirectory` helper in `client/v2/helpers`.

## [v2.0.0-beta.2](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.2) - 2024-06-19

### Features

* [#19039](https://github.com/cosmos/cosmos-sdk/pull/19039) Add support for pubkey in autocli.

### Improvements

* [#19646](https://github.com/cosmos/cosmos-sdk/pull/19646) Use keyring from command context.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Upgrade SDK version due to prometheus breaking change.
* (deps) [#19810](https://github.com/cosmos/cosmos-sdk/pull/19810) Bump `cosmossdk.io/store` to v1.1.0.
* [#20083](https://github.com/cosmos/cosmos-sdk/pull/20083) Integrate latest version of cosmos-proto and improve version filtering.
* [#19618](https://github.com/cosmos/cosmos-sdk/pull/19618) Marshal enum as string in queries.
* [#19060](https://github.com/cosmos/cosmos-sdk/pull/19060) Use client context from root (or enhanced) command in autocli commands.
  * Note, the given command must have a `client.Context` in its context.
* [#19216](https://github.com/cosmos/cosmos-sdk/pull/19216) Do not overwrite TxConfig, use directly the one provided in context. TxConfig should always be set in the `client.Context` in `root.go` of an app.
* [#20266](https://github.com/cosmos/cosmos-sdk/pull/20266) Add ability to override the short description in AutoCLI-generated top-level commands.

### Bug Fixes

* [#19976](https://github.com/cosmos/cosmos-sdk/pull/19976) Add encoder for `cosmos.base.v1beta1.DecCoin`.
* [#19377](https://github.com/cosmos/cosmos-sdk/pull/19377) Partly fix comment parsing in autocli.
* [#19060](https://github.com/cosmos/cosmos-sdk/pull/19060) Simplify key flag parsing logic in flag handler.
* [#20033](https://github.com/cosmos/cosmos-sdk/pull/20033) Respect output format from client ctx.

### API Breaking Changes

* [#19646](https://github.com/cosmos/cosmos-sdk/pull/19646) Remove keyring from `autocli.AppOptions` and `flag.Builder` options.

## [v2.0.0-beta.1](https://github.com/cosmos/cosmos-sdk/tree/client/v2.0.0-beta.1) - 2023-11-07

This is the first tagged version of client/v2.
It depends on the Cosmos SDK v0.50 release and fully supports AutoCLI.
