# PENDING CHANGELOG

<!----------------------------- BREAKING CHANGES ----------------------------->

## BREAKING CHANGES

### Gaia REST API

* [\#3641] Remove the ability to use a Keybase from the REST API client:
  * `password` and `generate_only` have been removed from the `base_req` object
  * All txs that used to sign or use the Keybase now only generate the tx
  * `keys` routes completely removed
* [\#3692] Update tx encoding and broadcasting endpoints:
  * Remove duplicate broadcasting endpoints in favor of POST @ `/txs`
    * The `Tx` field now accepts a `StdTx` and not raw tx bytes
  * Move encoding endpoint to `/txs/encode`

### Gaia CLI

### Gaia

### SDK

* [\#3669] Ensure consistency in message naming, codec registration, and JSON
tags.
* [\#3666] Improve coins denom validation.
* [\#3751] Disable (temporarily) support for ED25519 account key pairs.

### Tendermint
* [\#3804] Update to Tendermint `v0.31.0-dev0`

<!--------------------------------- FEATURES --------------------------------->

## FEATURES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

* [\#3719](https://github.com/cosmos/cosmos-sdk/issues/3719) DBBackend can now be set at compile time.
  Defaults: goleveldb. Supported: cleveldb.

### Tendermint

<!------------------------------- IMPROVEMENTS ------------------------------->

## IMPROVEMENTS

### Gaia REST API

* Update the `TxResponse` type allowing for the `Logs` result to be JSON
decoded automatically.

### Gaia CLI

* [\#3653] Prompt user confirmation prior to signing and broadcasting a transaction.
* [\#3670] CLI support for showing bech32 addresses in Ledger devices
* [\#3711] Update `tx sign` to use `--from` instead of the deprecated `--name`
CLI flag.
* [\#3738] Improve multisig UX:
  * `gaiacli keys show -o json` now includes constituent pubkeys, respective weights and threshold
  * `gaiacli keys show --show-multisig` now displays constituent pubkeys, respective weights and threshold
  * `gaiacli tx sign --validate-signatures` now displays multisig signers with their respective weights
* [\#3730](https://github.com/cosmos/cosmos-sdk/issues/3730) Improve workflow for
`gaiad gentx` with offline public keys, by outputting stdtx file that needs to be signed.
* [\#3761](https://github.com/cosmos/cosmos-sdk/issues/3761) Querying account related information using custom querier in auth module

### Gaia

### SDK

* \#3753 Remove no-longer-used governance penalty parameter
* \#3679 Consistent operators across Coins, DecCoins, Int, Dec
          replaced: Minus->Sub Plus->Add Div->Quo
* [\#3665] Overhaul sdk.Uint type in preparation for Coins Int -> Uint migration.
* \#3691 Cleanup error messages
* \#3456 Integrate in the Int.ToDec() convenience function
* [\#3300] Update the spec-spec, spec file reorg, and TOC updates.
* [\#3694] Push tagged docker images on docker hub when tag is created.
* [\#3716] Update file permissions the client keys directory and contents to `0700`.

### Tendermint

* [\#3699] Upgrade to Tendermint 0.30.1

<!--------------------------------- BUG FIXES -------------------------------->

## BUG FIXES

### Gaia REST API

### Gaia CLI

* [\#3731](https://github.com/cosmos/cosmos-sdk/pull/3731) `keys add --interactive` bip32 passphrase regression fix
* [\#3714](https://github.com/cosmos/cosmos-sdk/issues/3714) Fix USB raw access issues with gaiacli when installed via snap

### Gaia

* [\#3777](https://github.com/cosmso/cosmos-sdk/pull/3777) `gaiad export` no longer panics when the database is empty

### SDK

* \#3728 Truncate decimal multiplication & division in distribution to ensure
         no more than the collected fees / inflation are distributed
* \#3727 Return on zero-length (including []byte{}) PrefixEndBytes() calls
* \#3559 fix occasional failing due to non-determinism in lcd test TestBonding
  where validator is unexpectedly slashed throwing off test calculations
* [\#3411] Include the `RequestInitChain.Time` in the block header init during
`InitChain`.
* [\#3717] Update the vesting specification and implementation to cap deduction from
`DelegatedVesting` by at most `DelegatedVesting`. This accounts for the case where
the undelegation amount may exceed the original delegation amount due to
truncation of undelegation tokens.
* [\#3717] Ignore unknown proposers in allocating rewards for proposers, in case
  unbonding period was just 1 block and proposer was already deleted.
* [\#3726] Cap(clip) reward to remaining coins in AllocateTokens.

### Tendermint
