# Upgrade Reference

This document provides a reference for upgrading from `v0.54.x` to `v0.55.x` of Cosmos SDK.


For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.55.x/CHANGELOG.md).

## Table of Contents

* [Breaking Changes](#breaking-changes)
    * [Removed: SIGN_MODE_TEXTUAL](#removed-sign_mode_textual)
    * [`query.ParsePagination` rejects limits above 10,000](#queryparsepagination-rejects-limits-above-10000)
* [New Features and Non-Breaking Changes](#new-features-and-non-breaking-changes)
    * [ML-DSA-65 Validator Keys](#ml-dsa-65-validator-consensus-keys) 



## Breaking Changes

### Removed: SIGN_MODE_TEXTUAL

`SIGN_MODE_TEXTUAL` (proto enum value `2`) and its entire implementation have been removed:

- `x/tx/signing/textual/` — all renderers, the CBOR encoder, test data, and internal protos
- `x/auth/tx/textual.go` and `ConfigOptions.TextualCoinMetadataQueryFn`
- Ledger + SIGN_MODE_TEXTUAL integration in `client/` flags and tx factory

The proto enum value `2` and string `"SIGN_MODE_TEXTUAL"` are **reserved** to prevent future reuse. ADR-050 is archived.

**Required action** if your app enabled SIGN_MODE_TEXTUAL:

1. Remove `TextualCoinMetadataQueryFn` from your `tx.ConfigOptions`:

```go
// Before
txConfig, err := tx.NewTxConfigWithOptions(cdc, tx.ConfigOptions{
    TextualCoinMetadataQueryFn: ...,
})

// After — field removed, omit it
txConfig, err := tx.NewTxConfigWithOptions(cdc, tx.ConfigOptions{...})
```

2. Remove any `SIGN_MODE_TEXTUAL` cases from signing mode handler switch statements.

3. Remove Ledger wiring that depended on `SIGN_MODE_TEXTUAL`.

### `query.ParsePagination` rejects limits above 10,000

`types/query.ParsePagination` now returns an `InvalidArgument` error ("limit must not exceed 10000") when `PageRequest.Limit` exceeds `query.MaxLimit` (10,000). Previously the value was silently capped to 10,000, so a caller requesting a larger page size got a smaller page back with no indication anything was reduced.

`query.Paginate` (used by most gRPC query handlers via `initPageRequestDefaults`) is unaffected and still silently caps oversized limits to 10,000.

**Required action:** any client or indexer that intentionally requests `limit > 10000` from an endpoint backed by `ParsePagination` must lower its requested limit or handle the new error.

## New Features and Non-Breaking Changes

These changes are informational and optional to adopt during the upgrade; they are not required for a successful migration.

### ML-DSA-65 Validator Consensus Keys

Cosmos SDK v0.54 registers the NIST ML-DSA-65 (FIPS 204) post-quantum signature scheme as a supported validator consensus key type. The new `cosmos.crypto.mldsa65.PubKey` / `PrivKey` proto messages, Amino routes (`cometbft/PubKeyMlDsa65`, `cometbft/PrivKeyMlDsa65`), interface-registry registration, multisig amino route, and `hd.MlDsa65Type` constant are all enabled by default.

**Action required:** none. Existing chains continue to accept only the consensus key types listed in `genesis.consensus_params.validator.pub_key_types` (still `["ed25519"]` by default). No state-machine-relevant behavior changes for chains that do not opt in.

**To opt in (new chains):** set `genesis.consensus_params.validator.pub_key_types` to `["ml_dsa_65"]` (or a list including it). Validators must then submit `MsgCreateValidator` with a `mldsa65.PubKey`. Test harnesses can use the new `testutil/network.Config.ValidatorConsensusKeyType` field together with `genutil.InitializeNodeValidatorFilesFromMnemonicWithKeyType` to spin up an in-process testnet pinned to ML-DSA-65.

**Operational considerations:** ML-DSA-65 keys and signatures are substantially larger than ed25519 (pubkey 1952 bytes vs 32, signature 3309 bytes vs 64). Chains enabling this key type should review `consensus_params.block.max_bytes` and gossip framing limits accordingly. The cometbft commit lift in this release expanded `MaxSignatureSize` and the per-validator `MaxCommitSigBytes` to accommodate the larger signatures; downstream applications relying on the previous fixed values may need to be re-examined.

