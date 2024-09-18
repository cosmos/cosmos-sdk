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

Since v0.13.0, x/tx follows Cosmos SDK semver: https://github.com/cosmos/cosmos-sdk/blob/main/RELEASES.md
-->

# Changelog

## [Unreleased]

### Improvements

* [#21712](https://github.com/cosmos/cosmos-sdk/pull/21712) Add `AminoNameAsTypeURL` option to Amino JSON encoder.
* [#21073](https://github.com/cosmos/cosmos-sdk/pull/21073) In Context use sync.Map `getSignersFuncs` map from concurrent writes, we also need to call Validate when using the legacy app.

## [v0.13.3](https://github.com/cosmos/cosmos-sdk/releases/tag/x/tx/v0.13.3) - 2024-04-22

### Improvements

* [#20049](https://github.com/cosmos/cosmos-sdk/pull/20049) Sort JSON attributes for `inline_json` encoder.

## [v0.13.2](https://github.com/cosmos/cosmos-sdk/releases/tag/x/tx/v0.13.2) - 2024-04-12

### Features

* [#19786](https://github.com/cosmos/cosmos-sdk/pull/19786)/[#19919](https://github.com/cosmos/cosmos-sdk/pull/19919) Add "inline_json" option to Amino JSON encoder.

### Improvements

* [#19845](https://github.com/cosmos/cosmos-sdk/pull/19845) Use hybrid resolver instead of only protov2 registry

### Bug Fixes

* [#19955](https://github.com/cosmos/cosmos-sdk/pull/19955) Don't shadow Amino marshalling error messages

## [v0.13.1](https://github.com/cosmos/cosmos-sdk/releases/tag/x/tx/v0.13.1) - 2024-03-05

### Features

* [#19618](https://github.com/cosmos/cosmos-sdk/pull/19618) Add enum as string option to encoder.

### Improvements

* [#18857](https://github.com/cosmos/cosmos-sdk/pull/18857) Moved `FormatCoins` from `core/coins` to this package under `signing/textual`.

### Bug Fixes

* [#19265](https://github.com/cosmos/cosmos-sdk/pull/19265) Reject denoms that contain a comma.

## [v0.13.0](https://github.com/cosmos/cosmos-sdk/releases/tag/x/tx/v0.13.0) - 2023-12-19

### Improvements

* [#18740](https://github.com/cosmos/cosmos-sdk/pull/18740) Support nested messages when fetching signers up to a default depth of 32.

## v0.12.0

### Improvements

* [#18309](https://github.com/cosmos/cosmos-sdk/pull/18309) Update encoder so that amino types default to msg type url.

## v0.11.0

### Improvements

* [#17787](https://github.com/cosmos/cosmos-sdk/pull/17787) Drop tip support.

## v0.10.0

### Features

* [#17681](https://github.com/cosmos/cosmos-sdk/pull/17681) Add encoder `DefineTypeEncoding` method for defining custom type encodings.
* [#17600](https://github.com/cosmos/cosmos-sdk/pull/17600) Add encoder `DefineScalarEncoding` method for defining custom scalar encodings.
* [#17600](https://github.com/cosmos/cosmos-sdk/pull/17600) Add indent option to encoder.

## v0.9.1

### Improvements

* [#16936](https://github.com/cosmos/cosmos-sdk/pull/16936) Remove extra whitespace when marshalling module accounts.

## v0.9.0

### Bug Fixes

* [#16681](https://github.com/cosmos/cosmos-sdk/pull/16681): Catch and fix `(*Decoder).Decode` crash from invalid length prefix in Tx bytes.

### Improvements

* [#16846](https://github.com/cosmos/cosmos-sdk/pull/16846): Harmonize interface `signing.TypeResolver` with the rest of the codebase (orm and client/v2).
* [#16684](https://github.com/cosmos/cosmos-sdk/pull/16684): Use `io.WriteString`+`fmt.Fprintf` to remove unnecessary `string`->`[]byte` roundtrip.

## v0.8.0

### Improvements

* [#16340](https://github.com/cosmos/cosmos-sdk/pull/16340): add `DefineCustomGetSigners` API function.

## v0.7.0

### API Breaking

* [#16044](https://github.com/cosmos/cosmos-sdk/pull/16044): rename aminojson.NewAminoJSON -> aminojson.NewEncoder.
* [#16047](https://github.com/cosmos/cosmos-sdk/pull/16047): aminojson.NewEncoder now takes EncoderOptions as an argument.
* [#16254](https://github.com/cosmos/cosmos-sdk/pull/16254): aminojson.Encoder.Marshal now sorts all fields like encoding/json.Marshal does, hence no more need for sdk.\*SortJSON.

## v0.6.2

### Improvements

* [#15873](https://github.com/cosmos/cosmos-sdk/pull/15873): add `Validate` method and only check for errors when `Validate` is explicitly called.

## v0.6.1

### Improvements

* [#15871](https://github.com/cosmos/cosmos-sdk/pull/15871)
    * `HandlerMap` now has a `DefaultMode()` getter method
    * Textual types use `signing.ProtoFileResolver` instead of `protoregistry.Files`

## v0.6.0

### API Breaking

* [#15709](https://github.com/cosmos/cosmos-sdk/pull/15709):
    * `GetSignersContext` has been renamed to `signing.Context`
    * `GetSigners` now returns `[][]byte` instead of `[]string`
    * `GetSignersOptions` has been renamed to `signing.Options` and requires `address.Codec`s for account and validator addresses
    * `GetSignersOptions.ProtoFiles` has been renamed to `signing.Options.FileResolver`

### Bug Fixes

* [#15849](https://github.com/cosmos/cosmos-sdk/pull/15849) Fix int64 usage for 32 bit platforms.

## v0.5.1

### Features

* [#15414](https://github.com/cosmos/cosmos-sdk/pull/15414) Add basic transaction decoding support.

## v0.5.0

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
