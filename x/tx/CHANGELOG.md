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

## API Breaking

* [#15278](https://github.com/cosmos/cosmos-sdk/pull/15278) Move `x/tx/{textual,aminojson}` into `x/tx/signing`.
* [#15302](https://github.com/cosmos/cosmos-sdk/pull/15302) `textual.NewSignModeHandler` now takes an options struct instead of a simple coin querier argument. It also returns an error.

## Improvements

* [#15302](https://github.com/cosmos/cosmos-sdk/pull/15302) Add support for a custom registry (e.g. gogo's MergedRegistry) to be plugged into SIGN_MODE_TEXTUAL.

