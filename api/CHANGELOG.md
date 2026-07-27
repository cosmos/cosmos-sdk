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

Change log entries are to be added to the Unreleased section from newest to oldest.
Each entry must include the Github issue reference in the following format:

* [#<issue-number>] Changelog message.

-->

# Changelog

## [Unreleased]

## [v1.1.0](https://github.com/cosmos/cosmos-sdk/releases/tag/api/v1.1.0) - 2026-07-27

### Features
* [#26678](https://github.com/cosmos/cosmos-sdk/pull/26678) Regenerate the API module against the latest protos:
  * Add the `cosmos/crypto/mldsa65` package with `PubKey` and `PrivKey` for ML-DSA-65 (FIPS 204) validator keys.
  * Add the `cosmos/crypto/secp256k1eth` package with `PubKey` and `PrivKey` for Ethereum-compatible secp256k1 validator keys.
  * Add `Params.sig_verify_cost_mldsa65` to `cosmos/auth/v1beta1/auth.proto`.
  * Add `ConsKeyEvidenceExpiry` to `cosmos/staking/v1beta1/staking.proto`.
  * Add `ConsensusKeyRotationHistory.evidence_expiry_time` and `.evidence_expiry_height` to `cosmos/staking/v1beta1/genesis.proto`. `maturity_time` may now be zero, meaning the re-rotation gate is retired and 
* [#26481](https://github.com/cosmos/cosmos-sdk/pull/26481) Add staking `Params.key_rotation_fee` to `cosmos/staking/v1beta1/staking.proto`.
* [#26460](https://github.com/cosmos/cosmos-sdk/pull/26460) Add `ConsensusKeyRotationHistory` and `PendingConsensusKeyRotation` to `cosmos/staking/v1beta1/genesis.proto`.
* [#26440](https://github.com/cosmos/cosmos-sdk/pull/26440) Add `Msg/RotateConsPubKey` with `MsgRotateConsPubKey` and `MsgRotateConsPubKeyResponse` to `cosmos/staking/v1beta1/tx.proto`.
* [#25607](https://github.com/cosmos/cosmos-sdk/pull/25607) Add `tendermint.types.AuthorityParams` and `ConsensusParams.authority` and `cosmos.consensus.v1.MsgUpdateParams.auth`.

### API Breaking

* [#26456](https://github.com/cosmos/cosmos-sdk/pull/26456) Remove the `cosmos/msg/textual/v1` package and the `SIGN_MODE_TEXTUAL` value from `cosmos.tx.signing.v1beta1.SignMode`. Enum number 2 and the name are reserved.
* [#26428](https://github.com/cosmos/cosmos-sdk/pull/26428) Change the gRPC-gateway route for `Query/AccountAddressByID` from `/cosmos/auth/v1beta1/address_by_id/{id}` to `/cosmos/auth/v1beta1/address_by_id/{account_id}`.
* [#26421](https://github.com/cosmos/cosmos-sdk/pull/26421) Remove `cosmos/protocolpool`
* [#25981](https://github.com/cosmos/cosmos-sdk/pull/25981) Remove the `cosmos/benchmark` package.
* [#25546](https://github.com/cosmos/cosmos-sdk/pull/25546) Remove the `cosmos/params` package.

## [v0.9.0](https://github.com/cosmos/cosmos-sdk/releases/tag/api/v0.9.0) - 2025-03-31

### Features

* [#23815](https://github.com/cosmos/cosmos-sdk/pull/23815) `x/epochs` API files
* [#23708](https://github.com/cosmos/cosmos-sdk/pull/23708) `unordered` transaction support

### Improvements

* [#24227](https://github.com/cosmos/cosmos-sdk/pull/24227) Minor dependency bumps


