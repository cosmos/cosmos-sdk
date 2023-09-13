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

### Features

* (x/evidence) [14724](https://github.com/cosmos/cosmos-sdk/pull/14724) The `x/evidence` module is extracted to have a separate go.mod file which allows it be a standalone module.
* (keeper) [#15420](https://github.com/cosmos/cosmos-sdk/pull/15420) Move `BeginBlocker` to the keeper folder & make HandleEquivocation private

### API Breaking Changes

* [#16008](https://github.com/cosmos/cosmos-sdk/pull/16008) NewKeeper now takes in a KVStoreService instead of KVStoreKey, most functions use context.Context instead of sdk.Context and `IterateEvidence` callback function now returns an error to stop interation (`errors.ErrStopIterating`).
* (keeper) [#15825](https://github.com/cosmos/cosmos-sdk/pull/15825) Evidence constructor now requires an `address.Codec` (`import "cosmossdk.io/core/address"`)
* [#16336](https://github.com/cosmos/cosmos-sdk/pull/16336) Use collections for state management:
    * Removed: keeper `SetEvidence`, `GetEvidence`, `IterateEvidences`, `GetAllEvidences`, `MustMarshalEvidence`, `MustUnmarshalEvidence`, `MarshalEvidence`, `UnmarshalEvidence`
* [#17688](https://github.com/cosmos/cosmos-sdk/pull/17688) Remove `comet.Info` as an arg to evidence

### Client Breaking Changes

* [#16625](https://github.com/cosmos/cosmos-sdk/pull/16625) The `simd q evidence evidence` command supports only querying an evidence by hash. For querying all evidences, use `simd q evidence list` instead.
