# PENDING CHANGELOG

<!----------------------------- BREAKING CHANGES ----------------------------->

## BREAKING CHANGES

### Gaia REST API

* [\#3641] Remove the ability to use a Keybase from the REST API client:
  * `password` and `generate_only` have been removed from the `base_req` object
  * All txs that used to sign or use the Keybase now only generate the tx
  * `keys` routes completely removed

### Gaia CLI

### Gaia

### SDK

* [\#3669] Ensure consistency in message naming, codec registration, and JSON
tags.

### Tendermint

<!--------------------------------- FEATURES --------------------------------->

## FEATURES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

### Tendermint

<!------------------------------- IMPROVEMENTS ------------------------------->

## IMPROVEMENTS

### Gaia REST API

* Update the `TxResponse` type allowing for the `Logs` result to be JSON
decoded automatically.

### Gaia CLI

* [\#3670] CLI support for showing bech32 addresses in Ledger devices
* [\#3711] Update `tx sign` to use `--from` instead of the deprecated `--name`
CLI flag.

### Gaia

### SDK

* \#3679 Consistent operators across Coins, DecCoins, Int, Dec
          replaced: Minus->Sub Plus->Add Div->Quo
* [\#3665] Overhaul sdk.Uint type in preparation for Coins Int -> Uint migration.
* \#3691 Cleanup error messages
* \#3456 Integrate in the Int.ToDec() convenience function
* [\#3300] Update the spec-spec, spec file reorg, and TOC updates.
* [\#3694] Push tagged docker images on docker hub when tag is created.

### Tendermint

<!--------------------------------- BUG FIXES -------------------------------->

## BUG FIXES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

* \#3728 Truncate decimal multiplication & division in distribution to ensure
         no more than the collected fees / inflation are distributed
* \#3559 fix occasional failing due to non-determinism in lcd test TestBonding
  where validator is unexpectedly slashed throwing off test calculations
* [\#3411] Include the `RequestInitChain.Time` in the block header init during
`InitChain`.

* Update the vesting specification and implementation to cap deduction from
`DelegatedVesting` by at most `DelegatedVesting`. This accounts for the case where
the undelegation amount may exceed the original delegation amount due to
truncation of undelegation tokens.

### Tendermint
