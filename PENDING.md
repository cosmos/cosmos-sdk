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

### Gaia CLI

* [\#3670] CLI support for showing bech32 addresses in Ledger devices

### Gaia

### SDK

* [\#3665] Overhaul sdk.Uint type in preparation for Coins Int -> Uint migration.
* \#3691 Cleanup error messages
* \#3456 Integrate in the Int.ToDec() convenience function
* [\#3300] Update the spec-spec, spec file reorg, and TOC updates.

### Tendermint

<!--------------------------------- BUG FIXES -------------------------------->

## BUG FIXES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

### Tendermint
