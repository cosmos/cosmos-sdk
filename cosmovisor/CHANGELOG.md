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

# Changelog

## [Unreleased]

### Features

+ [\#8590](https://github.com/cosmos/cosmos-sdk/pull/8590) File watcher for cosmovisor. Instead of parsing logs from stdin and stderr, we watch the `<DAEMON_HOME>/data/upgrade-info.json` file updates using polling mechanism.
+ [\#10128 ](https://github.com/cosmos/cosmos-sdk/pull/10128) Change default value of `DAEMON_RESTART_AFTER_UPGRADE` to `true`.

### Improvements

+ [\#10018](https://github.com/cosmos/cosmos-sdk/pull/10018) Strict boolean argument parsing: cosmovisor will fail if user will not set correctly a boolean variable. Correct values are: "true", "false", "" (not setting) - all case not sensitive.
+ [\#10036](https://github.com/cosmos/cosmos-sdk/pull/10036) Improve logs when downloading the binary.

## v0.1 2021-08-06

This is the first release and we started this changelog on 2021-07-01. See the (README)[https://github.com/cosmos/cosmos-sdk/blob/release/cosmovisor/v0.1.x/cosmovisor/CHANGELOG.md] file for the full list of features.

## Features

* [\#9652](https://github.com/cosmos/cosmos-sdk/pull/9652) Add backup option for cosmovisor.
