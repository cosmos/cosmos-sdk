# Cosmos SDK v0.40.1 "Stargate" Release Notes

This is a bug fix release to the *Cosmos SDK 0.40* "Stargate" release series. No breaking changes are introduced, thus no migration should be needed.
Among the various bugfixes, this release introduces important security patches.

See the [Cosmos SDK v0.40.1 milestone](https://github.com/cosmos/cosmos-sdk/milestone/36?closed=1) on our issue tracker for details.

### Gogo protobuf security release

[Gogoprotobuf](https://github.com/gogo/protobuf) released a bugfix addressing [CVE-2021-3121](https://nvd.nist.gov/vuln/detail/CVE-2021-3121). Cosmos SDK respective dependency has been updated and protobuf generated files were regenerated.

### Tendermint security patches

This release comes with a newer release of Tendermint that, other than fixing various bugs it also addresses a high-severity security vulnerability.
For the comprehensive list of changes introduced by Tendermint since Cosmos SDK v0.40.0, please refer to [Tendermint's v0.34.3 release notes](https://github.com/tendermint/tendermint/blob/v0.34.3/CHANGELOG.md#v0.34.3).

### Fix zero time checks

In Cosmos SDK applications, it is recommended to use `Time.Unix() <= 0` instead of `Time.IsZero()`, which may lead to unexpected results.
See [\#8085](https://github.com/cosmos/cosmos-sdk/pull/8058) for more information.

### Querying upgrade plans panics when no plan exists

The `x/upgrade` module command and REST endpoints for querying upgrade plans would panic when no plan existed. This is now resolved.

### Fix account sequence

In Cosmos SDK v0.40.0 a new structure (`SignatureV2`) for handling message signatures was introduced.
Although the `tx sign --signature-only` command was still capable of generating correct signatures, it was not returning the account's correct sequence number.

### Reproducible builds

Our automatic builds were not working correctly due to small gaps in file paths. Fixed in [\8300](https://github.com/cosmos/cosmos-sdk/pull/8300) and [\8301](https://github.com/cosmos/cosmos-sdk/pull/8301).

### Wrapper errors were not reflective

Cosmos SDK errors typically support the `Is()` method. The `Go` `errors.Is()` function compares an error to a value and should always return `true` when the receiver is passed as an argument to its own method, e.g. `err.Is(err)`. This was not a case for the error types provided by the `types/errors` package.

### Fix latest consensus state

Errors occur when using the client to send IBC transactions without flag `--absolute-timeouts`, e.g:

    gaiad tx ibc-transfer transfer

The issue was caused by incorrect interface decoding and unpacking of `Any` values and is now fixed.
