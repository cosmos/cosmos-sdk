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

## Unreleased

### Features

* [#15414](https://github.com/cosmos/cosmos-sdk/pull/15414) Add basic transaction decoding support.

### API Breaking

* [#15581](https://github.com/cosmos/cosmos-sdk/pull/15581) `GetSignersOptions` and `directaux.SignModeHandlerOptions` now
require a `signing.ProtoFileResolver` interface instead of `protodesc.Resolver`.
* [#15742](https://github.com/cosmos/cosmos-sdk/pull/15742) The `direct_aux` package has been renamed to `directaux` in line with Go conventions. No other types were changed during the package rename.
* [#15748](https://github.com/cosmos/cosmos-sdk/pull/15748) Rename signing.SignerData.ChainId to .ChainID, in line with Go conventions.

### Bug Fixes

* (signing/textual) [#15730](https://github.com/cosmos/cosmos-sdk/pull/15730) make IntValueRenderer.Parse: gracefully handle "" + fuzz

## v0.4.0

### API Breaking

* [#13793](https://github.com/cosmos/cosmos-sdk/pull/13793) `direct_aux.NewSignModeHandler` constructor function now returns an additional error argument.
* [#15278](https://github.com/cosmos/cosmos-sdk/pull/15278) Move `x/tx/{textual,aminojson}` into `x/tx/signing`.
* [#15302](https://github.com/cosmos/cosmos-sdk/pull/15302) `textual.NewSignModeHandler` now takes an options struct instead of a simple coin querier argument. It also returns an error.

### Improvements

* [#15302](https://github.com/cosmos/cosmos-sdk/pull/15302) Add support for a custom registry (e.g. gogo's MergedRegistry) to be plugged into SIGN_MODE_TEXTUAL.
* [#15557](https://github.com/cosmos/cosmos-sdk/pull/15557) Implement unknown field filtering.
* [#15515](https://github.com/cosmos/cosmos-sdk/pull/15515) Implement SIGN_MODE_LEGACY_AMINO_JSON handler.
