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

Change log entries are to be added to the Unreleased section from newest to oldest.
Each entry must include the Github issue reference in the following format:

* [#<issue-number>] Changelog message.

-->

# Changelog

## [Unreleased]

* [#23486](https://github.com/cosmos/cosmos-sdk/pull/23486) Add `server/v2/api/swagger` server component.

## [v2.0.0-beta.2](https://github.com/cosmos/cosmos-sdk/releases/tag/server/v2.0.0-beta.2)

* [#23411](https://github.com/cosmos/cosmos-sdk/pull/23411) Use only command name in `IsAppRequired()` instead of command usage
* [#23333](https://github.com/cosmos/cosmos-sdk/pull/23333) Fix gRPC reflection
* [#22715](https://github.com/cosmos/cosmos-sdk/pull/22941) Add custom HTTP handler for grpc-gateway that removes the need to manually register grpc-gateway services.

## [v2.0.0-beta.1](https://github.com/cosmos/cosmos-sdk/releases/tag/server/v2.0.0-beta.1)

Initial tag of `cosmossdk.io/server/v2`.
