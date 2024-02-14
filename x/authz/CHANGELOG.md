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

### Consens Breaking Changes

* [#19188](https://github.com/cosmos/cosmos-sdk/pull/19188) Remove creation of `BaseAccount` when sending a message to an account that does not exist

### Features

* [#18737](https://github.com/cosmos/cosmos-sdk/pull/18737) Added a limit of 200 grants pruned per `BeginBlock` and the `PruneExpiredGrants` message that prunes 75 expired grants on every run.

### Improvements

### API Breaking Changes

* [#18737](https://github.com/cosmos/cosmos-sdk/pull/18737) Update the keeper method `DequeueAndDeleteExpiredGrants` to take a limit argument for the number of grants to prune.
* [#16509](https://github.com/cosmos/cosmos-sdk/pull/16509) `AcceptResponse` has been moved to sdk/types/authz and the `Updated` field is now of the type `sdk.Msg` instead of `authz.Authorization`.

### Bug Fixes
