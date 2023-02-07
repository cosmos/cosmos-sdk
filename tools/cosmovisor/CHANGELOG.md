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

## Client Breaking Changes

* [#14881](https://github.com/cosmos/cosmos-sdk/pull/14881) Cosmovisor supports only upgrade plan with a checksum. This is enforced by the `x/upgrade` module for better security.

## Improvements

* [#14881](https://github.com/cosmos/cosmos-sdk/pull/14881) Refactor Cosmovisor to use `x/upgrade` validation logic.
* [#14881](https://github.com/cosmos/cosmos-sdk/pull/14881) Refactor Cosmovisor to depend only on the `x/upgrade` module.

## v1.4.0 2022-10-23

### API Breaking Changes

* [#13603](https://github.com/cosmos-sdk/pull/13603) Rename cosmovisor package to `cosmossdk.io/tools/cosmovisor`.

## v1.3.0 2022-09-11

### Improvements

* [#12921](https://github.com/cosmos/cosmos-sdk/pull/12918) Add documentation about expected plan path name.
* [#12918](https://github.com/cosmos/cosmos-sdk/pull/12918) Automatically set version using module version.

### Bug Fixes

* [#13221](https://github.com/cosmos/cosmos-sdk/pull/13221) Fix `go install`.

## v1.2.0 2022-07-26

### Features

* [\#12464](https://github.com/cosmos/cosmos-sdk/pull/12464) Create the `cosmovisor init` command.
* [\#12188](https://github.com/cosmos/cosmos-sdk/pull/12188) Add a `DAEMON_RESTART_DELAY` for allowing a node operator to define a delay between the node halt (for upgrade) and backup.
* [\#11823](https://github.com/cosmos/cosmos-sdk/pull/11823) Refactor `cosmovisor` CLI to use `cobra`.
* [\#11731](https://github.com/cosmos/cosmos-sdk/pull/11731) `cosmovisor version -o json` returns the cosmovisor version and the result of `simd --output json --long` in one JSON object.

### Bug Fixes

* [\#12005](https://github.com/cosmos/cosmos-sdk/pull/12005) Fix cosmovisor binary usage for pre-upgrade

### CLI Breaking Changes

* [\#12188](https://github.com/cosmos/cosmos-sdk/pull/12188) Remove the possibility to set a time with only a number. `DAEMON_POLL_INTERVAL` env variable now only supports a duration (e.g. `100ms`, `30s`, `20m`).

## v1.1.0 2022-02-10

### Features

* [\#10285](https://github.com/cosmos/cosmos-sdk/pull/10316) Added `run` command to run the associated app.
* [\#10649](https://github.com/cosmos/cosmos-sdk/pull/10649) Customize backup directory. Added new env variable: `DAEMON_DATA_BACKUP_DIR`. If it is set, cosmovisor will backup the app data in `DAEMON_DATA_BACKUP_DIR` before running the update.

### Deprecated

* [\#10285](https://github.com/cosmos/cosmos-sdk/pull/10316) Running `cosmovisor` without the `run` argument.

### Bug Fixes

* [\#10458](https://github.com/cosmos/cosmos-sdk/pull/10458) Fix version when using 'go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0' to install cosmovisor.

## v1.0.0 2021-09-30

### Features

* [\#8590](https://github.com/cosmos/cosmos-sdk/pull/8590) File watcher for cosmovisor. Instead of parsing logs from stdin and stderr, we watch the `<DAEMON_HOME>/data/upgrade-info.json` file updates using polling mechanism.
* [\#9999](https://github.com/cosmos/cosmos-sdk/pull/10103) Added `version` command that returns the cosmovisor version and the application version.
* [\#9973](https://github.com/cosmos/cosmos-sdk/pull/10056) Added support for pre-upgrade command in Cosmovisor to be called before the binary is upgraded. Added new environmental variable `DAEMON_PREUPGRADE_MAX_RETRIES` that holds the maximum number of times to reattempt pre-upgrade before failing.
* [\#10126](https://github.com/cosmos/cosmos-sdk/pull/10229) Added `help`.

### Improvements

* [\#10018](https://github.com/cosmos/cosmos-sdk/pull/10018) Strict boolean argument parsing: cosmovisor will fail if user will not set correctly a boolean variable. Correct values are: "true", "false", "" (not setting) - all case not sensitive.
* [\#10036](https://github.com/cosmos/cosmos-sdk/pull/10036) Improve logs when downloading the binary.
* [\#10217](https://github.com/cosmos/cosmos-sdk/pull/10217) Replacing logging to use zerolog.

### CLI Breaking

* [\#10128](https://github.com/cosmos/cosmos-sdk/pull/10128) Change default value of `DAEMON_RESTART_AFTER_UPGRADE` to `true`.

## v0.1 2021-08-06

This is the first release and we started this changelog on 2021-07-01. See the [README](https://github.com/cosmos/cosmos-sdk/blob/release/cosmovisor/v0.1.x/cosmovisor/CHANGELOG.md) file for the full list of features.

## Features

* [\#9652](https://github.com/cosmos/cosmos-sdk/pull/9652) Add backup option for cosmovisor.
