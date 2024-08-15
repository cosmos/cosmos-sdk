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

`SimApp` is an application built using the Cosmos SDK for testing and educational purposes.
It won't be tagged or intented to be imported in an application.
This changelog is aimed to help developers understand the wiring changes between SDK versions.
It is an exautive list of changes that completes the SimApp section in the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#simapp)

## v0.50 to v0.51

Always refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md) to understand the changes.

* [#20409](https://github.com/cosmos/cosmos-sdk/pull/20409) Add `tx` as `SkipStoreKeys` in `app_config.go`.
* [#20485](https://github.com/cosmos/cosmos-sdk/pull/20485) The signature of `x/upgrade/types.UpgradeHandler` has changed to accept `appmodule.VersionMap` from `module.VersionMap`.  These types are interchangeable, but usages of `UpradeKeeper.SetUpgradeHandler` may need to adjust their usages to match the new signature.
* [#20740](https://github.com/cosmos/cosmos-sdk/pull/20740) Update `genutilcli.Commands` to use the genutil modules from the module manager.
* [#20771](https://github.com/cosmos/cosmos-sdk/pull/20771) Use client/v2 `GetNodeHomeDirectory` helper in `app.go` and use the `DefaultNodeHome` constant everywhere in the app.

<!-- TODO: move changelog.md elements to here -->

## v0.47 to v0.50

No changelog is provided for this migration. Please refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#v050x)

## v0.46 to v0.47

No changelog is provided for this migration. Please refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#v047x)

## v0.45 to v0.46

No changelog is provided for this migration. Please refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#v046x)
