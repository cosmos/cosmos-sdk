# Cosmos SDK v0.40.0 "Stargate" Release Notes

This is a bug fix release to the *Cosmos SDK 0.40* "Stargate" release series.
None of the changes is a breaking change, so no migration is needed.

### Fix zero time checks

We found that `Time.IsZero` is bug prone and could lead to wrong results. In SDK, we recommend to use `Time.Unix() <= 0` instead. Details [\#8085](https://github.com/cosmos/cosmos-sdk/pull/8058) fix zero time checks.

### Fix GET /upgrade/current

The `/update/current` REST endpoint didn't work correctly. Fixed in [\#8280](https://github.com/cosmos/cosmos-sdk/pull/8280).

### FIX: return correct account sequence

In "Stargate" we introduce a new structure (`SignatureV2`) for handling message signatures. However, the response didn't have a correct account sequence for signatures (example: `tx sign --signature-only`). Signatures were correct, so this is not a breaking change. Fix: [\#8287](https://github.com/cosmos/cosmos-sdk/pull/8287).

### Reproducible builds

Our automatic builds were not working correctly due to small gaps in file paths. Fixed in [\8300](https://github.com/cosmos/cosmos-sdk/pull/8300) and [\8301](https://github.com/cosmos/cosmos-sdk/pull/8301).

### Wrapped errors Is method was not reflective.

Many errors support `Is` method, which should be reflective, ie: `err.Is(err) = true`. This was not a case for `wrappedError` and all predefined errors in `types/errors` package. Fixed in [\#8355][https://github.com/cosmos/cosmos-sdk/pull/8355].

### Fix Latest consensus state

Errors occur when using the client to send IBC transactions without flag `--absolute-timeouts`, example:

    gaiad tx ibc-transfer transfer

The problem was with decoding transactions (specifically: interface decoding and wrong use of unpacking `Any` values). Fixed in: [\#8341](https://github.com/cosmos/cosmos-sdk/pull/8341) and [\#8359](https://github.com/cosmos/cosmos-sdk/pull/8359).


### Security patches for gogo proto

[Gogoprotobuf](https://github.com/gogo/protobuf) released new versions with security patches v1.3.2. We updated and regenerated protobuf generated files. [\#8350][https://github.com/cosmos/cosmos-sdk/pull/8350] abd [\#8361][https://github.com/cosmos/cosmos-sdk/pull/8361].


### Tendermint security patches

Tendermint release security patches v0.34.2 and and v0.34.3. We updated Tendermint dependency to the latest version.


## Known issues

No known issues.


## Update instructions

* Download or rebuild the app based on the latest SDK version
* restart the node

NOTE: no migration is needed.
