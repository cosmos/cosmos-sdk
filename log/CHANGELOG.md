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

## [v1.6.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.6.0) - 2025-05-13

* [#24720](https://github.com/cosmos/cosmos-sdk/pull/24720) add `VerboseModeLogger` extension interface and `VerboseLevel` configuration option for increasing log verbosity during sensitive operations such as upgrades.

## [v1.5.1](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.5.1) - 2025-03-07

* [#23928](https://github.com/cosmos/cosmos-sdk/pull/23928) Bump sonic json library to [v1.3.1](https://github.com/bytedance/sonic/releases/tag/v1.13.1) for Go 1.24 compatibility.

## [v1.5.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.5.0) - 2024-11-07

* [#22466](https://github.com/cosmos/cosmos-sdk/pull/22466) Disable coloring in testing logger.
* [#22233](https://github.com/cosmos/cosmos-sdk/pull/22233) Use sonic json library for faster json handling.
* [#22347](https://github.com/cosmos/cosmos-sdk/pull/22347) Add `cosmossdk.io/log/slog` to allow using a standard library log/slog-backed logger. This required to bump the minimum go version of `cosmossdk.io/log` to 1.21.

## [v1.4.1](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.4.1) - 2024-08-16

* [#21326](https://github.com/cosmos/cosmos-sdk/pull/21326) Avoid context key collision.

## [v1.4.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.4.0) - 2024-08-07

* [#21045](https://github.com/cosmos/cosmos-sdk/pull/21045) Add `WithContext` method implementations to make all returned loggers compatible with `cosmossdk.io/core/log.Logger` (v1) without a direct dependency.

## [v1.3.1](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.3.1) - 2024-02-05

* [#19346](https://github.com/cosmos/cosmos-sdk/pull/19346) Upgrade zerolog to v1.32.0.
* [#19346](https://github.com/cosmos/cosmos-sdk/pull/19346) `#15956` now works thanks to the upgrade of `zerolog`.

## [v1.3.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.3.0) - 2024-01-10

* [#18916](https://github.com/cosmos/cosmos-sdk/pull/18916) Introduce an option for setting hooks.
* [#18429](https://github.com/cosmos/cosmos-sdk/pull/18429) Support customization of log json marshal.
* [#18898](https://github.com/cosmos/cosmos-sdk/pull/18898) Add `WARN` level.

## [v1.2.1](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.2.1) - 2023-08-25

* [#17532](https://github.com/cosmos/cosmos-sdk/pull/17532) Proper marshalling of `fmt.Stringer` (follow-up of [#17205](https://github.com/cosmos/cosmos-sdk/pull/17205)).

## [v1.2.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.2.0) - 2023-07-31

* [#17194](https://github.com/cosmos/cosmos-sdk/pull/17194) Avoid repeating parse log level in `ParseLogLevel`.
* [#17205](https://github.com/cosmos/cosmos-sdk/pull/17205) Fix types that do not implement the `json.Marshaler` interface.
* [#15956](https://github.com/cosmos/cosmos-sdk/pull/15956) Introduce an option for enabling error stack trace.

## [v1.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.1.0) - 2023-04-27

* [#15956](https://github.com/cosmos/cosmos-sdk/pull/15956) Introduce options to configure logger (enable/disable colored output, customize log timestamps).

## [v1.0.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v1.0.0) - 2023-03-30

* [#15601](https://github.com/cosmos/cosmos-sdk/pull/15601) Introduce logger options. These options allow to configure the logger with filters, different level and output format.

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/log/v0.1.0) - 2023-03-13

* Introducing a standalone SDK logger package (`comossdk.io/log`).
  It replaces CometBFT logger and provides a common interface for all SDK components.
  The default logger (`NewLogger`) is using [zerolog](https://github.com/rs/zerolog),
  but it can be easily replaced with any implementation that implements the `log.Logger` interface.
