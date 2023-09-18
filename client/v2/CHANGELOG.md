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

<!-- ## [v2.1.0-alpha.1] to be tagged after v0.51 final-->

### API Breaking Changes

* [#17709](https://github.com/cosmos/cosmos-sdk/pull/17709) Address codecs have been removed from `autocli.AppOptions` and `flag.Builder`. Instead client/v2 uses the address codecs present in the context (introduced in [#17503](https://github.com/cosmos/cosmos-sdk/pull/17503)).

## [v2.0.0-alpha.1] <!-- to be tagged after tx support and v0.50 final-->

This is the first tagged version of client/v2.
It depends on the Cosmos SDK v0.50 release and fully supports AutoCLI.
