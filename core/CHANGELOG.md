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

## [v0.11.3](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.11.3)

* [#24088](https://github.com/cosmos/cosmos-sdk/pull/24088) Convert store interface type definitions to type aliases for compatibility with `cosmossdk.io/collections` `v1.2.x`.

## [v0.11.2](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.11.2)

* [#21298](https://github.com/cosmos/cosmos-sdk/pull/21298) Backport [#19265](https://github.com/cosmos/cosmos-sdk/pull/19265) to core.
* [#21298](https://github.com/cosmos/cosmos-sdk/pull/21298) Clean-up after depinject upgrade.

## [v0.11.1](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.11.1)

* [#21022](https://github.com/cosmos/cosmos-sdk/pull/21022) Upgrade depinject to v1.0.0.

## [v0.11.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.11.0)

* [#17468](https://github.com/cosmos/cosmos-sdk/pull/17468) Add `appmodule.HasPreBlocker` interface.

## [v0.10.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.10.0)

* [#17383](https://github.com/cosmos/cosmos-sdk/pull/17383) Add `appmoduke.UpgradeModule` interface.

## [v0.9.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.9.0)

* [#16739](https://github.com/cosmos/cosmos-sdk/pull/16739) Add `AppHash` to header.Info.

## [v0.8.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.8.0)

* [#15519](https://github.com/cosmos/cosmos-sdk/pull/15519) Update `comet.VoteInfo` for CometBFT v0.38.
* [#16310](https://github.com/cosmos/cosmos-sdk/pull/16310) Add `gas.Service` and `gas.GasMeter` interfaces.

## [v0.7.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.7.0)

* [#15850](https://github.com/cosmos/cosmos-sdk/pull/15850) Add `comet` and `header` packages.
* [#15923](https://github.com/cosmos/cosmos-sdk/pull/15923) Add `appmodule.HasPrepareCheckState` `appmodule.HasPrecommit` extension interfaces.
* [#15434](https://github.com/cosmos/cosmos-sdk/pull/15434) Add `coin.ParseCoin` for parsing a coin from a string.
* [#15999](https://github.com/cosmos/cosmos-sdk/pull/15999) Add `genesis.GenesisTxHandler` interface.

## [v0.6.1](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.6.1)

* [#15364](https://github.com/cosmos/cosmos-sdk/pull/15364) Add address codec to core.

## [v0.6.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.6.0)

* [#15045](https://github.com/cosmos/cosmos-sdk/pull/15045) Add error return parameter to `RegisterServices` method from `appmodule.HasServices` interface.
* [#14859](https://github.com/cosmos/cosmos-sdk/pull/14859) Simplify event service interface.

## [v0.5.1](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.5.1)

* [#14686](https://github.com/cosmos/cosmos-sdk/pull/14686) Add event service.
* [#14735](https://github.com/cosmos/cosmos-sdk/pull/14735) Specify event listener API.

## [v0.5.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.5.0)

* [#14604](https://github.com/cosmos/cosmos-sdk/pull/14604) Add begin/end block extension interfaces.
* [#14605](https://github.com/cosmos/cosmos-sdk/pull/14605) Add register services extension interface.

## [v0.4.1](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.4.1)

* [#14329](https://github.com/cosmos/cosmos-sdk/pull/14329) Implement basic core API genesis source and target.

## [v0.4.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.4.0)

* [#14416](https://github.com/cosmos/cosmos-sdk/pull/14416) Update core to use cosmos-db instead of tm-db.
* [#14326](https://github.com/cosmos/cosmos-sdk/pull/14326) Remove `appmodule.Service` from core.

## [v0.3.4](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.3.4)

* [#14223](https://github.com/cosmos/cosmos-sdk/pull/14223) Add genesis API.

## [v0.3.3](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.3.3)

* [#14227](https://github.com/cosmos/cosmos-sdk/pull/14227) Add store API.
* [#13696](https://github.com/cosmos/cosmos-sdk/pull/13696) Update `FormatCoins` where empty coins are rendered as "zero".

## [v0.3.2](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.3.2)

* [#13631](https://github.com/cosmos/cosmos-sdk/pull/13631) Add ADR 033 (inter-module communication) Client interface.

## [v0.3.1](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.3.1)

* [#13306](https://github.com/cosmos/cosmos-sdk/pull/13306) Move `FormatCoins` to core.
* [#13607](https://github.com/cosmos/cosmos-sdk/pull/13115) Add `AppModule` tag interface.

## [v0.3.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.3.0)

* [#13115](https://github.com/cosmos/cosmos-sdk/pull/13115) Update core module to depinject changes.

## [v0.2.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.2.0)

* [#12020](https://github.com/cosmos/cosmos-sdk/pull/12020) Use cosmossdk.io/depinject instead of container.
* [#12367](https://github.com/cosmos/cosmos-sdk/pull/12367) Add support for golang_bindings in app.yaml.

## [v0.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/core%2Fv0.1.0)

* [#11914](https://github.com/cosmos/cosmos-sdk/pull/11914) Add core module with app config support.
